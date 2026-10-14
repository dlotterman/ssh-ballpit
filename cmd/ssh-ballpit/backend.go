package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// A missing known_hosts file is an empty trust store only in explicit lab mode.
func backendHostKeys(config Config) (ssh.HostKeyCallback, error) {
	verify, err := knownhosts.New(config.KnownHosts)
	if err != nil && config.AcceptAllKeys && os.IsNotExist(err) {
		return knownhosts.New()
	}
	return verify, err
}

// stdout and stderr are written concurrently; callers must provide separate
// writers or a shared writer that explicitly supports concurrent writes.
func execute(ctx context.Context, config Config, user, command string, stdout, stderr io.Writer, identity ssh.Signer, verify ssh.HostKeyCallback) (uint32, error) {
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", config.Backend)
	if err != nil {
		return 255, backendFailure{stage: "connect", reason: connectionReason(err)}
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(identity)},
		HostKeyCallback: func(host string, addr net.Addr, key ssh.PublicKey) error {
			if err := verify(host, addr, key); err != nil {
				if config.AcceptAllKeys {
					if config.metrics != nil {
						config.metrics.hostKeyBypasses.Inc()
					}
					// x/crypto/ssh verifies the exchange signature before this callback.
					slog.Error("ssh untrusted backend host key accepted", "accept_all_keys", true,
						"backend", config.Backend, "backend_user", user, "remote_addr", addr.String(),
						"fingerprint", ssh.FingerprintSHA256(key), "key_type", key.Type(),
						"reason", connectionReason(err), "signature_verified", true,
						"known_hosts_entry", knownhosts.Line([]string{host}, key))
					return nil
				}
				return backendFailure{stage: "host_key_verification", reason: connectionReason(err)}
			}
			return nil
		},
	}
	clientConn, channels, requests, err := ssh.NewClientConn(conn, config.Backend, clientConfig)
	if err != nil {
		return 255, backendFailure{stage: "handshake", reason: connectionReason(err)}
	}
	conn.SetDeadline(time.Time{})
	client := ssh.NewClient(clientConn, channels, requests)
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return 255, backendFailure{stage: "open_session", reason: connectionReason(err)}
	}
	defer session.Close()
	session.Stdout = stdout
	session.Stderr = stderr
	// nil Stdin provides EOF; client data can never become an unchecked program.
	err = session.Run(command)
	if err == nil {
		return 0, nil
	}
	if status, ok := err.(*ssh.ExitError); ok {
		return uint32(status.ExitStatus()), nil
	}
	return 255, backendFailure{stage: "execute", reason: connectionReason(err)}
}
