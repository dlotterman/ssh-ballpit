package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func testKey(t *testing.T, dir, name string) ssh.Signer {
	t.Helper()
	_, key, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	block, e := ssh.MarshalPrivateKey(key, "")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, name), pem.EncodeToMemory(block), 0600); e != nil {
		t.Fatal(e)
	}
	s, e := ssh.NewSignerFromKey(key)
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestSSHRelay(t *testing.T) {
	t.Run("plaintext_commands", func(t *testing.T) { testSSHRelay(t, false, true) })
	t.Run("strict", func(t *testing.T) { testSSHRelay(t, false, false) })
	t.Run("accept_all_keys", func(t *testing.T) { testSSHRelay(t, true, false) })
}
func testSSHRelay(t *testing.T, lab, logCommands bool) {
	logs := captureLogs(t)
	defer func() {
		text := logs.String()
		for _, want := range []string{"ssh authentication succeeded", "unknown_user", "ssh public key rejected", "ssh handshake failed", "ssh channel rejected", "ssh request rejected"} {
			if !strings.Contains(text, want) {
				t.Errorf("missing integration log %s", want)
			}
		}
	}()

	dir := t.TempDir()
	front := testKey(t, dir, "front")
	clientKey := testKey(t, dir, "client")
	backendKey := testKey(t, dir, "backend")
	backendHost := testKey(t, dir, "backend-host")
	if !lab {
		os.WriteFile(filepath.Join(dir, "authorized"), ssh.MarshalAuthorizedKey(clientKey.PublicKey()), 0600)
	}
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer listener.Close()
	trustedHostKey := backendHost.PublicKey()
	if lab {
		trustedHostKey = front.PublicKey()
	} // Deliberate backend host-key mismatch in lab mode.
	os.WriteFile(filepath.Join(dir, "known"), []byte(knownhosts.Line([]string{listener.Addr().String()}, trustedHostKey)+"\n"), 0600)
	var executions atomic.Int32
	backendConf := &ssh.ServerConfig{PublicKeyCallback: func(m ssh.ConnMetadata, k ssh.PublicKey) (*ssh.Permissions, error) {
		if m.User() != "local-alice" || ssh.FingerprintSHA256(k) != ssh.FingerprintSHA256(backendKey.PublicKey()) {
			return nil, fmt.Errorf("bad identity")
		}
		return nil, nil
	}}
	backendConf.AddHostKey(backendHost)
	go func() {
		for {
			raw, e := listener.Accept()
			if e != nil {
				return
			}
			go func() {
				defer raw.Close()
				s, channels, requests, e := ssh.NewServerConn(raw, backendConf)
				if e != nil {
					return
				}
				defer s.Close()
				go ssh.DiscardRequests(requests)
				for newCh := range channels {
					ch, requests, e := newCh.Accept()
					if e != nil {
						return
					}
					for req := range requests {
						if req.Type != "exec" {
							req.Reply(false, nil)
							continue
						}
						executions.Add(1)
						req.Reply(true, nil)
						data, _ := io.ReadAll(ch)
						if len(data) != 0 {
							fmt.Fprint(ch, "unexpected stdin")
						}
						fmt.Fprint(ch, "backend stdout\n")
						fmt.Fprint(ch.Stderr(), "backend stderr\n")
						finish(ch, 7)
						break
					}
				}
			}()
		}
	}()
	var deny atomic.Bool
	var reviews atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reviews.Add(1)
		score := 0
		if deny.Load() {
			score = 90
		}
		answer := fmt.Sprintf(`{"offensive":false,"danger_score":%d,"reasons":[],"destinations":[]}`, score)
		json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": answer}})
	}))
	defer provider.Close()
	probe, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	addr := probe.Addr().String()
	probe.Close()
	c := Config{Listen: addr, HostKey: filepath.Join(dir, "front"), AuthorizedKeys: filepath.Join(dir, "authorized"), Backend: listener.Addr().String(), BackendKey: filepath.Join(dir, "backend"), KnownHosts: filepath.Join(dir, "known"), Users: map[string]User{"alice": {BackendUser: "local-alice", Fingerprints: []string{ssh.FingerprintSHA256(clientKey.PublicKey())}}}, Providers: []Provider{{Name: "mock", Kind: "ollama", Endpoint: provider.URL, Model: "mock", Threshold: 30}}, MaxConnections: 8, SessionSeconds: 5, ReviewSeconds: 2}
	c.AcceptAllKeys = lab
	c.LogCommands = logCommands
	if lab {
		c.Users["alice"] = User{BackendUser: "local-alice"}
	}
	metricsProbe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	c.MetricsListen = metricsProbe.Addr().String()
	metricsProbe.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, c) }()
	defer func() {
		cancel()
		select {
		case e := <-done:
			if e != nil {
				t.Error(e)
			}
		case <-time.After(3 * time.Second):
			t.Error("server failed to stop")
		}
	}()
	clientConf := &ssh.ClientConfig{User: "alice", Auth: []ssh.AuthMethod{ssh.PublicKeys(clientKey)}, HostKeyCallback: ssh.FixedHostKey(front.PublicKey()), Timeout: time.Second}
	dial := func() *ssh.Client {
		t.Helper()
		var cl *ssh.Client
		var err error
		for i := 0; i < 100; i++ {
			cl, err = ssh.Dial("tcp", addr, clientConf)
			if err == nil {
				return cl
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal(err)
		return nil
	}
	cl := dial()
	session, e := cl.NewSession()
	if e != nil {
		t.Fatal(e)
	}
	session.Stdin = strings.NewReader("unreviewed input must not reach backend\n")
	output, e := session.CombinedOutput("uname -a")
	if exit, ok := e.(*ssh.ExitError); !ok || exit.ExitStatus() != 7 {
		t.Fatalf("exit: %v", e)
	}
	if string(output) != "backend stdout\nbackend stderr\n" && string(output) != "backend stderr\nbackend stdout\n" {
		t.Fatalf("output %q", output)
	}
	cl.Close()
	deny.Store(true)
	// Approved repeats execute again even while the reviewer would now deny them.
	cl = dial()
	session, _ = cl.NewSession()
	_, e = session.CombinedOutput("uname -a")
	if exit, ok := e.(*ssh.ExitError); !ok || exit.ExitStatus() != 7 {
		t.Fatalf("cached execution: %v", e)
	}
	cl.Close()
	if reviews.Load() != 1 || executions.Load() != 2 {
		t.Fatal("repeat was reviewed again or did not execute")
	}
	if !strings.Contains(logs.String(), `"hit":true`) {
		t.Fatal("cache hit was not logged")
	}
	cl = dial()
	session, _ = cl.NewSession()
	_, e = session.CombinedOutput("uname -r")
	if exit, ok := e.(*ssh.ExitError); !ok || exit.ExitStatus() != 126 {
		t.Fatalf("denial: %v", e)
	}
	cl.Close()
	if executions.Load() != 2 {
		t.Fatalf("denied command reached backend")
	}
	cl = dial()
	if _, e = cl.Dial("tcp", "127.0.0.1:22"); e == nil {
		t.Fatal("forwarding accepted")
	}
	session, _ = cl.NewSession()
	if e = session.Shell(); e == nil {
		t.Fatal("shell accepted")
	}
	session.Close()
	cl.Close()
	const rejectedCommand = "bash # token='log-test-secret'\n"
	cl = dial()
	session, _ = cl.NewSession()
	_, e = session.CombinedOutput(rejectedCommand)
	if exit, ok := e.(*ssh.ExitError); !ok || exit.ExitStatus() != 126 {
		t.Fatalf("static denial: %v", e)
	}
	cl.Close()
	var stdout, stderr bytes.Buffer
	code, err := execute(context.Background(), c, "local-alice", "uname -a", &stdout, &stderr, backendKey, ssh.FixedHostKey(front.PublicKey()))
	if lab {
		if err != nil || code != 7 || executions.Load() != 3 {
			t.Fatalf("lab backend mismatch not accepted: code=%d err=%v", code, err)
		}
		if stdout.String() != "backend stdout\n" || stderr.String() != "backend stderr\n" {
			t.Fatalf("backend output streams mixed or lost: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	} else if err == nil || code != 255 || executions.Load() != 2 {
		t.Fatal("backend host-key mismatch was not rejected")
	}
	clientConf.User = "root"
	if cl, e := ssh.Dial("tcp", addr, clientConf); e == nil {
		cl.Close()
		t.Fatal("unmapped user accepted")
	}
	listener.Close()
	deny.Store(false)
	clientConf.User = "alice"
	cl = dial()
	session, e = cl.NewSession()
	if e != nil {
		t.Fatal(e)
	}
	_, e = session.CombinedOutput("uname -a")
	if exit, ok := e.(*ssh.ExitError); !ok || exit.ExitStatus() != 255 {
		t.Fatalf("backend failure exit: %v", e)
	}
	cl.Close()
	if !strings.Contains(logs.String(), "backend failed") || !strings.Contains(logs.String(), "connection_refused") {
		t.Fatal("missing backend connection failure diagnostics")
	}
	if lab {
		found := false
		for _, line := range strings.Split(logs.String(), "\n") {
			var event map[string]any
			if json.Unmarshal([]byte(line), &event) == nil && event["msg"] == "ssh untrusted key accepted" {
				found = true
				if event["level"] != "ERROR" || event["signature_verified"] != true || event["fingerprint"] != ssh.FingerprintSHA256(clientKey.PublicKey()) || event["authorized_key"] != strings.TrimSpace(string(ssh.MarshalAuthorizedKey(clientKey.PublicKey()))) || event["backend_user"] != "local-alice" {
					t.Fatal("missing verified key collection fields")
				}
			}
		}
		if !found || !strings.Contains(logs.String(), "ACCEPT ALL SSH KEYS ENABLED") {
			t.Fatal("missing loud lab logs")
		}
	}
	if lab {
		found := false
		for _, line := range strings.Split(logs.String(), "\n") {
			var event map[string]any
			if json.Unmarshal([]byte(line), &event) == nil && event["msg"] == "ssh untrusted backend host key accepted" {
				found = true
				if event["level"] != "ERROR" || event["signature_verified"] != true || event["fingerprint"] != ssh.FingerprintSHA256(backendHost.PublicKey()) || event["known_hosts_entry"] != knownhosts.Line([]string{listener.Addr().String()}, backendHost.PublicKey()) {
					t.Fatal("missing backend host-key collection fields")
				}
			}
		}
		if !found {
			t.Fatal("backend mismatch bypass not logged")
		}
	}
	response, err := http.Get("http://" + c.MetricsListen + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	metricBytes, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || response.StatusCode != 200 {
		t.Fatalf("metrics scrape: %v status=%d", err, response.StatusCode)
	}
	metricsText := string(metricBytes)
	for _, want := range []string{
		`ssh_ballpit_commands_total{outcome="nonzero_exit"} 2`,
		`ssh_ballpit_commands_total{outcome="static_denied"} 1`,
		`ssh_ballpit_commands_total{outcome="review_denied"} 1`,
		`ssh_ballpit_commands_total{outcome="backend_error"} 1`,
		`ssh_ballpit_approval_cache_lookups_total{result="hit"} 2`,
		`ssh_ballpit_approval_cache_entries 1`,
		`ssh_ballpit_reviews_total{outcome="approved",reviewer="mock"} 1`,
		`ssh_ballpit_reviews_total{outcome="denied",reviewer="mock"} 1`,
		`ssh_ballpit_execution_duration_seconds_count 3`,
		`go_goroutines`, `process_cpu_seconds_total`,
	} {
		if !strings.Contains(metricsText, want) {
			t.Errorf("missing metric %s", want)
		}
	}
	for _, secret := range []string{"local-alice", ssh.FingerprintSHA256(clientKey.PublicKey()), "uname -a", "command_sha256", "remote_addr"} {
		if strings.Contains(metricsText, secret) {
			t.Errorf("sensitive metric label %s", secret)
		}
	}
	expectedHostBypasses := "ssh_ballpit_backend_host_key_bypasses_total 0"
	expectedBypasses := "ssh_ballpit_key_authorization_bypasses_total 0"
	expectedLab := "ssh_ballpit_accept_all_keys_enabled 0"
	if lab {
		expectedLab = "ssh_ballpit_accept_all_keys_enabled 1"
		expectedBypasses = "ssh_ballpit_key_authorization_bypasses_total 6"
		expectedHostBypasses = "ssh_ballpit_backend_host_key_bypasses_total 2"
	}
	if !strings.Contains(metricsText, expectedHostBypasses) {
		t.Fatal("incorrect backend bypass metric")
	}
	if !strings.Contains(metricsText, expectedBypasses) {
		t.Fatal("incorrect successful bypass counter")
	}
	if !strings.Contains(metricsText, expectedLab) {
		t.Fatal("incorrect lab mode metric")
	}
	loggedCommands := map[string]int{}
	for _, line := range strings.Split(logs.String(), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal("command text broke JSON log framing", err)
		}
		if event["msg"] == "ssh command received" {
			command, ok := event["command"].(string)
			if !ok {
				t.Fatal("missing command text")
			}
			loggedCommands[command]++
			if event["user"] != "alice" || event["backend_user"] != "local-alice" || event["command_sha256"] == "" {
				t.Fatal("missing command context")
			}
		}
	}
	if logCommands {
		if len(loggedCommands) != 3 || loggedCommands["uname -a"] != 3 || loggedCommands["uname -r"] != 1 || loggedCommands[rejectedCommand] != 1 {
			t.Fatalf("unexpected command logs: %v", loggedCommands)
		}
	} else if len(loggedCommands) != 0 || strings.Contains(logs.String(), "log-test-secret") || strings.Contains(logs.String(), "uname -a") {
		t.Fatal("plaintext commands logged while disabled")
	}
	deadline := time.Now().Add(time.Second)
	for !strings.Contains(logs.String(), "ssh handshake failed") && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
}

func TestBackendHostKeys(t *testing.T) {
	dir := t.TempDir()
	key := testKey(t, dir, "host").PublicKey()
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2222}
	for _, lab := range []bool{false, true} {
		for _, fixture := range []string{"missing", "empty", "malformed", "trusted"} {
			t.Run(fmt.Sprintf("%s/lab=%t", fixture, lab), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "known_hosts")
				contents := map[string]string{
					"empty":     "",
					"malformed": "not a known_hosts entry\n",
					"trusted":   knownhosts.Line([]string{addr.String()}, key) + "\n",
				}
				if fixture != "missing" {
					if err := os.WriteFile(path, []byte(contents[fixture]), 0600); err != nil {
						t.Fatal(err)
					}
				}
				verify, err := backendHostKeys(Config{KnownHosts: path, AcceptAllKeys: lab})
				wantError := fixture == "malformed" || (fixture == "missing" && !lab)
				if (err != nil) != wantError {
					t.Fatalf("loader error = %v, want error %t", err, wantError)
				}
				if err != nil {
					return
				}
				err = verify(addr.String(), addr, key)
				if fixture == "trusted" {
					if err != nil {
						t.Fatal(err)
					}
				} else if !strings.HasPrefix(connectionReason(err), "host_key_unknown:") {
					t.Fatalf("empty trust store should report unknown key, got %v", err)
				}
			})
		}
	}
}
