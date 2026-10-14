package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/crypto/ssh"
)

func exampleCommandRules(t *testing.T) commandRules {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"whitelist.yaml", "blacklist.yaml"} {
		b, err := examples.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	rules, err := readCommandRules(dir)
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func TestCommandRuleExamples(t *testing.T) {
	c := Config{commandRules: exampleCommandRules(t)}
	for _, tc := range []struct {
		command, source string
	}{
		{"ls -al /home/*", "whitelist"},
		{"ls -al /home/alice/projects", "whitelist"},
		{"ls -al /home/", "whitelist"},
		{"ls -al /home/'a b'", "whitelist"},
		{"systemd-analyze --user verify", "whitelist"},
		{"systemd-analyze --user verify foo.service", ""},
		{"systemd-analyze --user", ""},
		{"ls -al /tmp/home/alice", ""},
		{"uptime", ""},
		{"reboot", "blacklist"},
		{" reboot \n", "blacklist"},
		{"uptime; reboot", "blacklist"},
		{"ls -al /home/alice && reboot", "blacklist"},
		{"ls -al /home/alice\nreboot", "blacklist"},
		{"printf hello | reboot", "blacklist"},
		{"MODE=test reboot > /tmp/log", "blacklist"},
		{"if true; then reboot; fi", "blacklist"},
		{"dnf", "blacklist"},
		{"dnf install cowsay", "blacklist"},
		{"dnf\tinstall cowsay", "blacklist"},
		{"dnf 'multiline\nargument'", "blacklist"},
		{"ls -al /home/alice; dnf upgrade", "blacklist"},
		{"printf reboot", ""},
		{"reboot-helper", ""},
		{"ls -al /home/alice; touch /tmp/file", ""},
		{"ls -al /home/alice && touch /tmp/file", ""},
		{"ls -al /home/alice | tee /tmp/file", ""},
		{"ls -al /home/alice > /tmp/file", ""},
		{"ls -al /home/alice &", ""},
		{"ls -al /home/$USER", ""},
		{"ls -al /home/${USER:=alice}", ""},
		{"ls -al /home/$((1 + 1))", ""},
		{"ls -al /home/$(id)", "static_policy"},
		{"ls -al /home/a; bash", "static_policy"},
	} {
		t.Run(tc.command, func(t *testing.T) {
			source, safe, err := localCommandDecision(c, tc.command)
			if source != tc.source || safe != (source == "whitelist") || (err != nil) != (source == "blacklist" || source == "static_policy") {
				t.Fatalf("source=%q safe=%t err=%v; want %q", source, safe, err, tc.source)
			}
		})
	}
}

func TestCommandRulesStrictLoading(t *testing.T) {
	for _, name := range []string{"whitelist.yaml", "blacklist.yaml"} {
		for _, content := range []string{
			"", "null", "{}", "patterns: null", "patterns: reboot", "unknown: []",
			"patterns: []\nextra: true", "patterns: []\npatterns: []",
			"patterns: []\n---\npatterns: []", "patterns: ['[']",
			"patterns: ['']", "patterns: ['   ']", "patterns: [null]",
			"patterns: [123]", "patterns: [true]", "patterns: [{}]",
			"patterns: ['(?=reboot)']", "#" + strings.Repeat("x", 1<<20),
		} {
			t.Run(name+"/"+content[:min(len(content), 50)], func(t *testing.T) {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := readCommandRules(dir); err == nil {
					t.Fatal("accepted malformed command rules")
				}
			})
		}
	}
	dir := t.TempDir()
	if rules, err := readCommandRules(dir); err != nil || len(rules.whitelist)+len(rules.blacklist) != 0 {
		t.Fatalf("missing optional files: %+v %v", rules, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "whitelist.yaml"), []byte("patterns: []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readCommandRules(dir); err != nil {
		t.Fatal("explicitly empty rules rejected", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "blacklist.yaml"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := readCommandRules(dir); err == nil {
		t.Fatal("unreadable rule file ignored")
	}
	dir = t.TempDir()
	if err := os.Symlink(filepath.Join(dir, "missing.yaml"), filepath.Join(dir, "blacklist.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := readCommandRules(dir); err == nil {
		t.Fatal("dangling blacklist symlink treated as missing optional file")
	}
}

func TestCommandRulesConfigAndAnchors(t *testing.T) {
	for _, format := range []string{"yaml", "json"} {
		t.Run(format, func(t *testing.T) {
			path := yamlFixture(t, "kind: ollama\nmodel: test\nendpoint: http://example.test\nthreshold: 30\n")
			if format == "json" {
				path = filepath.Join(filepath.Dir(path), "server.json")
				b, err := examples.ReadFile("config.example.json")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0600); err != nil {
					t.Fatal(err)
				}
			}
			dir := filepath.Dir(path)
			for _, name := range []string{"whitelist.yaml", "blacklist.yaml"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte("patterns: ['reboot|dnf', '(?m)^pwd$']\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			c, err := readConfig(path)
			if err != nil {
				t.Fatal(err)
			}
			if source, safe, _ := localCommandDecision(c, "reboot"); source != "blacklist" || safe {
				t.Fatal("blacklist did not override whitelist")
			}
			for _, command := range []string{"reboot-now", "mydnf", "echo\npwd\necho"} {
				if matchesCommand(c.commandRules.whitelist, command) {
					t.Fatalf("not a full-command match: %q", command)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "blacklist.yaml"), []byte("patterns: ['[']"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := readConfig(path); err == nil {
				t.Fatal("configuration accepted invalid regex")
			}
		})
	}
}

func TestBroadWhitelistCannotBypassStaticOrShellStructure(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "whitelist.yaml"), []byte("patterns: ['(?s:.*)']"), 0600); err != nil {
		t.Fatal(err)
	}
	rules, err := readCommandRules(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{
		"", "bash", "echo $(id)", "/bin/ls", "ls\x00", strings.Repeat("x", maxCommandBytes+1),
		"PATH=/tmp ls", "ls; id", "ls | id", "ls > /tmp/file", "ls &", "! ls",
		"(ls)", "{ ls; }", "for x in a; do ls; done", "f() { ls; }; f",
		"ls $PATH", "ls $((a=1))", "ls # comment", "ls\n# comment",
	} {
		if source, safe, _ := localCommandDecision(Config{commandRules: rules}, command); safe || source == "whitelist" {
			t.Fatalf("whitelist bypass for %q", command)
		}
	}
}

// A channel fixture exercises the complete SSH command decision handler without
// running submitted programs. Approved commands reach the invalid backend and
// return 255; local/model denials return 126 before attempting that connection.
type ruleTestChannel struct {
	bytes.Buffer
	stderr bytes.Buffer
	status uint32
}

func (c *ruleTestChannel) Close() error          { return nil }
func (c *ruleTestChannel) CloseWrite() error     { return nil }
func (c *ruleTestChannel) Stderr() io.ReadWriter { return &c.stderr }
func (c *ruleTestChannel) SendRequest(name string, reply bool, payload []byte) (bool, error) {
	if name == "exit-status" {
		var status struct{ Status uint32 }
		if err := ssh.Unmarshal(payload, &status); err != nil {
			return false, err
		}
		c.status = status.Status
	}
	return true, nil
}

func TestCommandRulesSession(t *testing.T) {
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": safeJSON}})
	}))
	defer provider.Close()
	c := Config{commandRules: exampleCommandRules(t), Backend: "invalid-address", ReviewSeconds: 2,
		Providers: []Provider{{Name: "mock", Kind: "ollama", Model: "test", Endpoint: provider.URL, Threshold: 30}}}
	c.metrics = newDaemonMetrics(c)
	reviewer := newApprovalReviewer(c)
	// Seed an existing model approval to prove blacklist is checked before cache.
	if _, safe, _ := reviewer.check(context.Background(), "alice", "local", "reboot"); !safe {
		t.Fatal("could not seed cache")
	}
	for _, tc := range []struct {
		command, source string
		status          uint32
		calls           int32
	}{
		{"reboot", "blacklist", 126, 1},
		{"ls -al /home/alice; reboot", "blacklist", 126, 1},
		{"systemd-analyze --user verify", "whitelist", 255, 1},
		{"ls -al /home/alice", "whitelist", 255, 1},
		{"uptime", "reviewers", 255, 2},
		{"uptime", "approval_cache", 255, 2},
		{"bash", "static_policy", 126, 2},
	} {
		ch := &ruleTestChannel{}
		requests := make(chan *ssh.Request, 1)
		requests <- &ssh.Request{Type: "exec", Payload: ssh.Marshal(struct{ Command string }{tc.command})}
		close(requests)
		handleSession(context.Background(), c, "alice", "local", ch, requests, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), reviewer)
		if ch.status != tc.status || calls.Load() != tc.calls {
			t.Fatalf("%q: exit=%d calls=%d stderr=%q", tc.command, ch.status, calls.Load(), ch.stderr.String())
		}
		history := reviewer.historySnapshot("alice", "local").Approved
		if tc.status == 126 {
			history = reviewer.historySnapshot("alice", "local").Rejected
		}
		if len(history) == 0 || history[len(history)-1].Source != tc.source {
			t.Fatalf("%q: incorrect decision history: %+v", tc.command, history)
		}
	}
	if len(reviewer.entries) != 2 {
		t.Fatal("whitelist approvals entered model cache")
	}
	recorder := httptest.NewRecorder()
	c.metrics.handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(recorder.Body.String(), `ssh_ballpit_commands_total{outcome="blacklist_denied"} 2`) {
		t.Fatal("blacklist denial metric missing")
	}
}

func runRulesCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var output bytes.Buffer
	err := run(args, &output)
	return output.String(), err
}

func TestCommandRulesCLI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.yaml")
	if _, err := runRulesCLI(t, "configure", "--config", path); err != nil {
		t.Fatal(err)
	}
	c, err := readConfig(path)
	if err != nil || len(c.commandRules.whitelist) != 2 || len(c.commandRules.blacklist) != 2 {
		t.Fatalf("configure did not generate usable rules: %v", err)
	}
	if _, err := runRulesCLI(t, "configure", "--config", path); err == nil {
		t.Fatal("configure overwrote existing files")
	}
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": safeJSON}})
	}))
	defer provider.Close()
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "providers", "ollama.yaml"),
		[]byte("kind: ollama\nmodel: test\nthreshold: 30\nendpoint: "+provider.URL+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		command, source string
		safe            bool
		calls           int32
	}{
		{"systemd-analyze --user verify", "whitelist", true, 0},
		{"reboot", "blacklist", false, 0},
		{"bash", "static_policy", false, 0},
		{"uptime", "", true, 1},
	} {
		output, err := runRulesCLI(t, "check", "--config", path, "--user", "alice", "--command", tc.command)
		var result struct {
			Safe   bool   `json:"safe"`
			Source string `json:"source"`
		}
		if json.Unmarshal([]byte(output), &result) != nil || result.Safe != tc.safe || result.Source != tc.source || (err == nil) != tc.safe || calls.Load() != tc.calls {
			t.Fatalf("check %q: output=%s err=%v calls=%d", tc.command, output, err, calls.Load())
		}
	}
	// Provider diagnostics must still work when a command rule matches everything.
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "blacklist.yaml"), []byte("patterns: ['.*']"), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := runRulesCLI(t, "providers-test", "--config", path); err != nil || calls.Load() != 2 {
		t.Fatalf("providers-test skipped provider: %s %v calls=%d", output, err, calls.Load())
	}
}
