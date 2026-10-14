package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

func serverConfig(config Config) (*ssh.ServerConfig, error) {
	data, err := os.ReadFile(config.AuthorizedKeys)
	if err != nil && !(config.AcceptAllKeys && os.IsNotExist(err)) {
		return nil, err
	}
	allowed := map[string]bool{}
	for len(bytes.TrimSpace(data)) > 0 {
		key, _, options, rest, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			return nil, fmt.Errorf("invalid authorized_keys: %w", err)
		}
		if len(options) > 0 {
			return nil, fmt.Errorf("authorized_keys options are unsupported; use bare keys and user mappings")
		}
		allowed[ssh.FingerprintSHA256(key)] = true
		data = rest
	}
	if len(allowed) == 0 && !config.AcceptAllKeys {
		return nil, fmt.Errorf("empty authorized_keys")
	}
	server := &ssh.ServerConfig{
		MaxAuthTries: 3,
		PublicKeyCallback: func(meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return authorizeClientKey(config, allowed, meta, key)
		},
		AuthLogCallback: func(meta ssh.ConnMetadata, method string, err error) {
			logAuthentication(config, meta, method, err)
		},
	}

	signer, err := loadSigner(config.HostKey)
	if err != nil {
		return nil, err
	}
	server.AddHostKey(signer)
	return server, nil
}

// A configured frontend account must map to a backend user, even in lab mode.
func authorizeClientKey(config Config, allowed map[string]bool, meta ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	fingerprint := ssh.FingerprintSHA256(key)
	mapping, ok := config.Users[meta.User()]
	reason := "key_not_allowed_for_user"
	switch {
	case !ok:
		reason = "unknown_user"
	case mapping.BackendUser == "":
		reason = "missing_backend_user"
	case !allowed[fingerprint]:
		reason = "key_not_in_authorized_keys"
	default:
		for _, allowedFingerprint := range mapping.Fingerprints {
			if allowedFingerprint == fingerprint {
				return &ssh.Permissions{Extensions: map[string]string{"backend_user": mapping.BackendUser, "fingerprint": fingerprint}}, nil
			}
		}
	}
	if config.AcceptAllKeys && ok && mapping.BackendUser != "" {
		slog.Error("ssh key authorization bypass requested", "remote_addr", meta.RemoteAddr().String(), "user", meta.User(), "fingerprint", fingerprint, "reason", reason, "accept_all_keys", true, "signature_verified", false)
		return &ssh.Permissions{Extensions: map[string]string{"backend_user": mapping.BackendUser, "fingerprint": fingerprint,
			"key_policy_bypassed": "true", "key_rejection_reason": reason,
			"authorized_key": strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))}}, nil
	}
	if config.metrics != nil {
		config.metrics.keyRejections.WithLabelValues(reason).Inc()
	}
	slog.Warn("ssh public key rejected", "remote_addr", meta.RemoteAddr().String(), "user", meta.User(), "fingerprint", fingerprint, "reason", reason)
	return nil, authFailure(reason)
}

func logAuthentication(config Config, meta ssh.ConnMetadata, method string, err error) {
	// Normal method discovery does not supply a credential.
	if method == "none" && err != nil {
		return
	}
	logger := slog.Default().With("remote_addr", meta.RemoteAddr().String(), "user", meta.User(), "method", method)
	if err == nil {
		if config.metrics != nil {
			config.metrics.authentication.WithLabelValues("success").Inc()
		}
		logger.Info("ssh authentication succeeded")
		return
	}
	reason := "authentication_failed"
	var failure authFailure
	if errors.As(err, &failure) {
		reason = string(failure)
	} else if method != "publickey" {
		reason = "unsupported_authentication_method"
	}
	if config.metrics != nil {
		config.metrics.authentication.WithLabelValues("failure").Inc()
	}
	logger.Warn("ssh authentication failed", "reason", reason)
}
