package main

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type authFailure string

func (err authFailure) Error() string { return string(err) }

type backendFailure struct{ stage, reason string }

func (err backendFailure) Error() string { return err.stage + ": " + err.reason }

// Never log raw peer-controlled SSH error text: disconnect messages can contain
// arbitrary data. Return stable operator-facing reasons instead.
func connectionReason(err error) string {
	var backend backendFailure
	if errors.As(err, &backend) {
		return backend.reason
	}
	var keys *knownhosts.KeyError
	if errors.As(err, &keys) {
		if len(keys.Want) == 0 {
			return "host_key_unknown: add the verified sshd host key to known_hosts"
		}
		return "host_key_mismatch: sshd host key differs from known_hosts"
	}
	var revoked *knownhosts.RevokedError
	if errors.As(err, &revoked) {
		return "host_key_revoked"
	}
	var auth *ssh.ServerAuthError
	if errors.As(err, &auth) {
		return "no_accepted_authentication: check preceding authentication logs"
	}
	var negotiation *ssh.AlgorithmNegotiationError
	if errors.As(err, &negotiation) {
		return "no_common_ssh_algorithm"
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return "timeout"
	}
	switch {
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, syscall.EADDRINUSE):
		return "address_in_use: another process is listening on this address"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "connection_refused: check backend address and sshd listener"
	case errors.Is(err, syscall.ECONNRESET):
		return "connection_reset_by_peer"
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return "peer_disconnected"
	case errors.Is(err, net.ErrClosed):
		return "connection_closed"
	}
	text := err.Error()
	switch {
	case strings.Contains(text, "unable to authenticate"):
		return "backend_authentication_rejected: check backend_user and backend_key authorization"
	case strings.Contains(text, "too many authentication failures"):
		return "too_many_authentication_failures"
	case strings.Contains(text, "host_key_"): // ssh handshake may stringify the host-key callback error.
		for _, reason := range []string{"host_key_unknown", "host_key_mismatch", "host_key_revoked"} {
			if strings.Contains(text, reason) {
				return reason
			}
		}
	case strings.Contains(text, "version"), strings.Contains(text, "handshake failed"), strings.Contains(text, "invalid packet"):
		return "invalid_ssh_handshake_or_protocol"
	}
	return "ssh_connection_or_protocol_error"
}
