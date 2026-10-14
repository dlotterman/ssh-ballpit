package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

func serve(ctx context.Context, config Config) error {
	ctx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()
	config.metrics = newDaemonMetrics(config)
	logStartupConfig(config)
	if err := ensureHostKey(config.HostKey); err != nil {
		return err
	}
	conf, err := serverConfig(config)
	if err != nil {
		return err
	}
	identity, err := loadSigner(config.BackendKey)
	if err != nil {
		return err
	}
	verify, err := backendHostKeys(config)
	if err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(config.Backend)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("backend must be a literal loopback address")
	}
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		slog.Error("ssh listen failed", "listen_addr", config.Listen, "reason", connectionReason(err))
		return err
	}
	defer listener.Close()
	stopMetrics, err := startMetrics(config.MetricsListen, config.metrics, cancelServer)
	if err != nil {
		return err
	}
	defer stopMetrics()
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	slots := make(chan struct{}, config.MaxConnections)
	reviewer := newApprovalReviewer(config)
	var wg sync.WaitGroup
	defer wg.Wait()
	slog.Info("listening", "address", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("ssh accept failed", "listen_addr", config.Listen, "reason", connectionReason(err))
			return err
		}
		config.metrics.connections.Inc()
		select {
		case slots <- struct{}{}:
		default:
			config.metrics.connectionRejections.Inc()
			slog.Warn("ssh connection rejected", "remote_addr", conn.RemoteAddr().String(), "reason", "connection_limit_reached", "max_connections", config.MaxConnections)
			conn.Close()
			continue
		}
		wg.Add(1)
		config.metrics.activeConnections.Inc()
		go func() {
			defer config.metrics.activeConnections.Dec()
			defer wg.Done()
			defer func() { <-slots }()
			defer conn.Close()
			handleConnection(ctx, config, conn, conf, identity, verify, reviewer)
		}()
	}
}

// handleConnection completes authentication and permits one command session.
// The caller owns connection accounting and closes the underlying TCP connection.
func handleConnection(ctx context.Context, config Config, conn net.Conn, sshConfig *ssh.ServerConfig, identity ssh.Signer, verify ssh.HostKeyCallback, reviewer *approvalReviewer) {
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	logger := slog.Default().With("remote_addr", conn.RemoteAddr().String(), "local_addr", conn.LocalAddr().String())
	s, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		if ctx.Err() == nil {
			config.metrics.handshakeFailures.Inc()
			logger.Warn("ssh handshake failed", "reason", connectionReason(err), "error_type", fmt.Sprintf("%T", err))
		}
		return
	}
	logger = logger.With("user", s.User(), "fingerprint", s.Permissions.Extensions["fingerprint"])
	if s.Permissions.Extensions["key_policy_bypassed"] == "true" {
		config.metrics.keyBypasses.Inc()
		logger.Error("ssh untrusted key accepted", "accept_all_keys", true, "signature_verified", true,
			"backend_user", s.Permissions.Extensions["backend_user"],
			"reason", s.Permissions.Extensions["key_rejection_reason"],
			"authorized_key", s.Permissions.Extensions["authorized_key"])
	}
	defer s.Close()
	conn.SetDeadline(time.Now().Add(time.Duration(config.SessionSeconds) * time.Second))
	sessionCtx, cancel := context.WithTimeout(ctx, time.Duration(config.SessionSeconds)*time.Second)
	defer cancel()
	go func() {
		s.Wait()
		cancel()
	}()
	go ssh.DiscardRequests(reqs)
	used := false
	for ch := range chans {
		if ch.ChannelType() != "session" || used {
			config.metrics.protocolRejections.WithLabelValues("channel").Inc()
			logger.Warn("ssh channel rejected", "channel_type", ch.ChannelType(), "reason", "only_one_exec_session_allowed")
			ch.Reject(ssh.Prohibited, "only one command session is supported")
			continue
		}
		used = true
		channel, requests, err := ch.Accept()
		if err != nil {
			logger.Warn("ssh channel accept failed", "reason", connectionReason(err))
			continue
		}
		handleSession(sessionCtx, config, s.User(), s.Permissions.Extensions["backend_user"], channel, requests, identity, verify, logger, reviewer)
		s.Close()
	}
}

func finish(ch ssh.Channel, code uint32) {
	ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{code}))
	ch.Close()
}

func handleSession(ctx context.Context, config Config, user, backendUser string, ch ssh.Channel, requests <-chan *ssh.Request, identity ssh.Signer, verify ssh.HostKeyCallback, logger *slog.Logger, reviewer *approvalReviewer) {
	defer ch.Close()
	for req := range requests {
		if req.Type != "exec" {
			if config.metrics != nil {
				config.metrics.protocolRejections.WithLabelValues("request").Inc()
			}
			logger.Warn("ssh request rejected", "request_type", req.Type, "reason", "only_exec_supported")
			req.Reply(false, nil)
			continue
		}
		var payload struct{ Command string }
		if ssh.Unmarshal(req.Payload, &payload) != nil {
			if config.metrics != nil {
				config.metrics.protocolRejections.WithLabelValues("malformed_exec").Inc()
			}
			logger.Warn("ssh request rejected", "request_type", "exec", "reason", "malformed_exec_request")
			req.Reply(false, nil)
			return
		}
		req.Reply(true, nil)
		go func() {
			for r := range requests {
				r.Reply(false, nil)
			}
		}()
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(payload.Command)))
		if config.LogCommands {
			logger.Info("ssh command received", "backend_user", backendUser, "command_sha256", hash, "command", payload.Command)
		}
		source, safe, decisionErr := localCommandDecision(config, payload.Command)
		if source != "" {
			reviewer.remember(user, backendUser, payload.Command, safe, source)
			logger.Info("command policy", "command_sha256", hash, "source", source, "safe", safe)
		}
		if decisionErr != nil {
			if config.metrics != nil {
				outcome := "static_denied"
				if source == "blacklist" {
					outcome = "blacklist_denied"
				}
				config.metrics.commands.WithLabelValues(outcome).Inc()
			}
			fmt.Fprintln(ch.Stderr(), "denied:", decisionErr)
			logger.Info("denied", "user", user, "command_sha256", hash, "reason", source)
			finish(ch, 126)
			return
		}
		if source == "" {
			var results []Result
			var hit bool
			results, safe, hit = reviewer.check(ctx, user, backendUser, payload.Command)
			logger.Info("approval cache", "command_sha256", hash, "hit", hit, "safe", safe)
			// Model reasons and provider response bodies stay out of logs.
			for _, r := range results {
				logger.Info("review", "user", user, "command_sha256", hash, "reviewer", r.Name, "safe", r.Safe, "error", r.Error)
			}
		}
		if !safe {
			if config.metrics != nil {
				config.metrics.commands.WithLabelValues("review_denied").Inc()
			}
			fmt.Fprintln(ch.Stderr(), "denied: every configured reviewer must approve")
			finish(ch, 126)
			return
		}
		started := time.Now()
		code, err := execute(ctx, config, backendUser, payload.Command, ch, ch.Stderr(), identity, verify)
		if config.metrics != nil {
			outcome := "success"
			if err != nil {
				outcome = "backend_error"
			} else if code != 0 {
				outcome = "nonzero_exit"
			}
			config.metrics.commands.WithLabelValues(outcome).Inc()
			config.metrics.executionDuration.Observe(time.Since(started).Seconds())
		}
		if err != nil {
			fmt.Fprintln(ch.Stderr(), "execution failed: backend unavailable or session interrupted")
			logger.Warn("backend failed", "backend", config.Backend, "backend_user", backendUser, "command_sha256", hash, "reason", connectionReason(err), "error", err.Error())
		}
		finish(ch, code)
		return
	}
}
