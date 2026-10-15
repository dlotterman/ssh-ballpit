package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type logBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}
func (b *logBuffer) String() string { b.mu.Lock(); defer b.mu.Unlock(); return b.b.String() }
func captureLogs(t *testing.T) *logBuffer {
	t.Helper()
	b := &logBuffer{}
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(b, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })
	return b
}

type authMetadata struct{ user string }

func (m authMetadata) User() string        { return m.user }
func (authMetadata) SessionID() []byte     { return nil }
func (authMetadata) ClientVersion() []byte { return []byte("SSH-2.0-test") }
func (authMetadata) ServerVersion() []byte { return []byte("SSH-2.0-test") }
func (authMetadata) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 12345}
}
func (authMetadata) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2222}
}
func TestAuthenticationLogReasons(t *testing.T) {
	logs := captureLogs(t)
	dir := t.TempDir()
	testKey(t, dir, "host")
	key := testKey(t, dir, "client")
	other := testKey(t, dir, "other")
	os.WriteFile(filepath.Join(dir, "authorized"), ssh.MarshalAuthorizedKey(key.PublicKey()), 0600)
	c := Config{HostKey: filepath.Join(dir, "host"), AuthorizedKeys: filepath.Join(dir, "authorized"), Users: map[string]User{"alice": {BackendUser: "alice", Fingerprints: []string{ssh.FingerprintSHA256(key.PublicKey())}}, "unmapped": {}, "other": {BackendUser: "other"}}}
	cfg, e := serverConfig(c)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		user   string
		key    ssh.PublicKey
		reason string
	}{{"root", key.PublicKey(), "unknown_user"}, {"unmapped", key.PublicKey(), "missing_backend_user"}, {"alice", other.PublicKey(), "key_not_in_authorized_keys"}, {"other", key.PublicKey(), "key_not_allowed_for_user"}} {
		_, e := cfg.PublicKeyCallback(authMetadata{tc.user}, tc.key)
		if e == nil {
			t.Fatal("authentication accepted")
		}
		cfg.AuthLogCallback(authMetadata{tc.user}, "publickey", e)
		if !strings.Contains(logs.String(), tc.reason) {
			t.Errorf("missing reason %s", tc.reason)
		}
	}
	if _, e := cfg.PublicKeyCallback(authMetadata{"alice"}, key.PublicKey()); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(logs.String(), "authentication succeeded") {
		t.Fatal("public key probe logged as successful authentication")
	}
	cfg.AuthLogCallback(authMetadata{"alice"}, "publickey", nil)
	cfg.AuthLogCallback(authMetadata{"alice"}, "password", errors.New("password=DO-NOT-LOG"))
	text := logs.String()
	for _, want := range []string{"192.0.2.1:12345", ssh.FingerprintSHA256(key.PublicKey()), "ssh authentication succeeded", "unsupported_authentication_method"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(text, "DO-NOT-LOG") {
		t.Fatal("credential leaked")
	}
}
func TestConnectionReasons(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{io.EOF, "peer_disconnected"}, {syscall.ECONNREFUSED, "connection_refused"}, {&knownhosts.KeyError{}, "host_key_unknown"}, {&knownhosts.KeyError{Want: []knownhosts.KnownKey{{}}}, "host_key_mismatch"}, {&ssh.ServerAuthError{}, "no_accepted_authentication"}, {errors.New("ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey]"), "backend_authentication_rejected"}, {errors.New("peer secret DO-NOT-LOG"), "ssh_connection_or_protocol_error"},
	} {
		got := connectionReason(tc.err)
		if !strings.Contains(got, tc.want) || strings.Contains(got, "DO-NOT-LOG") {
			t.Errorf("reason %q", got)
		}
	}
}

func TestProviderTransportReasons(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{&net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, "connection_refused"},
		{&net.OpError{Op: "dial", Err: syscall.EHOSTUNREACH}, "network_unreachable"},
		{&net.DNSError{Err: "lookup failed", Name: "secret.example"}, "dns_error"},
		{context.DeadlineExceeded, "timeout"},
		{errors.New("secret URL or peer response"), "http_transport_error"},
	} {
		got := providerTransportReason(tc.err)
		if got != tc.want || strings.Contains(got, "secret") {
			t.Errorf("provider transport reason: got %q, want %q", got, tc.want)
		}
	}
}

func TestStartupConfigDoesNotLogSecrets(t *testing.T) {
	logs := captureLogs(t)
	dir := t.TempDir()
	secretFile := filepath.Join(dir, "credential")
	if err := os.WriteFile(secretFile, []byte("FILE-SECRET"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BALLPIT_TEST_CREDENTIAL", "ENV-SECRET")
	c := Config{Listen: "127.0.0.1:2222", Backend: "127.0.0.1:22", ReviewSeconds: 30, SessionSeconds: 300, MaxConnections: 32,
		Users: map[string]User{"alice": {BackendUser: "localalice", Fingerprints: []string{"SHA256:public"}}},
		Providers: []Provider{
			{Name: "inline", Kind: "openai", Model: "test-model", Threshold: 25, APIKey: "INLINE-SECRET", Endpoint: "https://USER-SECRET:PASS-SECRET@example.com/PATH-SECRET?key=QUERY-SECRET#FRAGMENT-SECRET"},
			{Name: "file", Kind: "gemini", KeyFile: secretFile},
			{Name: "env", Kind: "anthropic", KeyEnv: "BALLPIT_TEST_CREDENTIAL"},
			{Name: "subscription", Kind: "openai-subscription", CodexHome: dir, CredentialsFile: "LEGACY-SECRET"},
			{Name: "local", Kind: "ollama", Endpoint: "http://192.0.2.1:11434"},
		}}
	logStartupConfig(c)
	output := logs.String()
	for _, secret := range []string{"FILE-SECRET", "ENV-SECRET", "INLINE-SECRET", "USER-SECRET", "PASS-SECRET", "PATH-SECRET", "QUERY-SECRET", "FRAGMENT-SECRET", "LEGACY-SECRET"} {
		if strings.Contains(output, secret) {
			t.Errorf("startup logs leaked %s", secret)
		}
	}
	for _, want := range []string{"startup configuration", "127.0.0.1:2222", "localalice", "SHA256:public", "test-model", `"threshold":25`, "https://example.com", "http://192.0.2.1:11434", `"credential_source":"inline"`, `"credential_source":"file"`, `"credential_source":"environment"`, `"credential_source":"codex_home"`, `"credential_source":"none"`} {
		if !strings.Contains(output, want) {
			t.Errorf("missing startup setting %s", want)
		}
	}
	if got := endpointOrigin("https://SECRET%broken"); strings.Contains(got, "SECRET") {
		t.Fatal("invalid endpoint leaked")
	}
}

func TestAcceptAllKeysPolicy(t *testing.T) {
	logs := captureLogs(t)
	dir := t.TempDir()
	testKey(t, dir, "host")
	key := testKey(t, dir, "client")
	c := Config{AcceptAllKeys: true, HostKey: filepath.Join(dir, "host"), AuthorizedKeys: filepath.Join(dir, "missing"), Users: map[string]User{"alice": {BackendUser: "local-alice"}, "unmapped": {}}}
	cfg, err := serverConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	perms, err := cfg.PublicKeyCallback(authMetadata{"alice"}, key.PublicKey())
	if err != nil || perms.Extensions["key_policy_bypassed"] != "true" {
		t.Fatal("unlisted key rejected in lab mode", err)
	}
	if strings.Contains(logs.String(), `"signature_verified":true`) {
		t.Fatal("unsigned key offer logged as verified")
	}
	for _, user := range []string{"root", "unmapped"} {
		if _, err := cfg.PublicKeyCallback(authMetadata{user}, key.PublicKey()); err == nil {
			t.Fatal("unmapped account accepted")
		}
	}
	c.AcceptAllKeys = false
	if _, err := serverConfig(c); err == nil {
		t.Fatal("strict mode allowed missing keys file")
	}
	c.AcceptAllKeys = true
	os.WriteFile(c.AuthorizedKeys, []byte("not a valid key"), 0600)
	if _, err := serverConfig(c); err == nil {
		t.Fatal("malformed keys configuration ignored")
	}
	// Also bypass a configured key that is missing from this user's fingerprint list.
	os.WriteFile(c.AuthorizedKeys, ssh.MarshalAuthorizedKey(key.PublicKey()), 0600)
	cfg, err = serverConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	perms, err = cfg.PublicKeyCallback(authMetadata{"alice"}, key.PublicKey())
	if err != nil || perms.Extensions["key_rejection_reason"] != "key_not_allowed_for_user" {
		t.Fatal("user fingerprint mismatch not bypassed")
	}
	c.AcceptAllKeys = false
	cfg, err = serverConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.PublicKeyCallback(authMetadata{"alice"}, key.PublicKey()); err == nil {
		t.Fatal("strict mode bypassed user fingerprints")
	}
}
