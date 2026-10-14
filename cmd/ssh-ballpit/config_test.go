package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func yamlFixture(t *testing.T, provider string) string {
	t.Helper()
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "providers"), 0700)
	os.WriteFile(filepath.Join(dir, "server.yaml"), []byte("users: {}\n"), 0600)
	os.WriteFile(filepath.Join(dir, "providers", "local.yaml"), []byte(provider), 0600)
	return filepath.Join(dir, "server.yaml")
}
func TestYAMLProviders(t *testing.T) {
	path := yamlFixture(t, "kind: ollama\nmodel: test\nendpoint: http://example.test:1234\nthreshold: 30\n")
	dir := filepath.Dir(path)
	os.WriteFile(filepath.Join(dir, "providers", "cloud.yaml"), []byte("kind: gemini\nmodel: test\nkey_file: ../gemini-key\nthreshold: 20\n"), 0600)
	os.WriteFile(filepath.Join(dir, "gemini-key"), []byte("test-key\n"), 0600)
	os.WriteFile(filepath.Join(dir, "providers", "chatgpt.yaml"), []byte("kind: openai-subscription\nmodel: test\nthreshold: 25\n"), 0600)
	c, e := readConfig(path)
	if e != nil {
		t.Fatal(e)
	}
	if len(c.Providers) != 3 {
		t.Fatal("not all providers loaded")
	}
	if c.HostKey != filepath.Join(dir, "state", "host_key") || c.Backend != "127.0.0.1:22" {
		t.Fatal("bad server defaults")
	}
	if c.Providers[0].CodexHome != filepath.Join(dir, "state", "codex") {
		t.Fatal("bad shared login path")
	}
	key, e := providerKey(c.Providers[1])
	if e != nil || key != "test-key" {
		t.Fatal("bad relative key file")
	}
	if c.Providers[2].Endpoint != "http://example.test:1234" {
		t.Fatal("Ollama endpoint was not taken from YAML")
	}
}
func TestBadYAMLProviders(t *testing.T) {
	for _, p := range []string{
		"kind: ollama\nmodel: test\nthreshold: 30\n", // no endpoint
		"kind: ollama\nkind: gemini\n",               // duplicate key
		"kind: ollama\nunknown: value\n",
		"kind: ollama\n---\nkind: gemini\n",
		"kind: gemini\nmodel: test\nthreshold: 30\napi_key: secret\nkey_env: GEMINI_API_KEY\n",
		"kind: gemini\nmodel: test\nthreshold: 101\n",
	} {
		path := yamlFixture(t, p)
		if _, e := readConfig(path); e == nil {
			t.Errorf("accepted %q", p)
		}
	}
}
func TestYAMLEmptyAndDuplicateReviewers(t *testing.T) {
	path := yamlFixture(t, "kind: ollama\nmodel: test\nendpoint: http://example.test\nthreshold: 30\nname: same\n")
	dir := filepath.Join(filepath.Dir(path), "providers")
	b, _ := os.ReadFile(filepath.Join(dir, "local.yaml"))
	os.WriteFile(filepath.Join(dir, "other.yaml"), b, 0600)
	if _, e := readConfig(path); e == nil {
		t.Fatal("duplicate names accepted")
	}
	os.Remove(filepath.Join(dir, "local.yaml"))
	os.Remove(filepath.Join(dir, "other.yaml"))
	if _, e := readConfig(path); e == nil {
		t.Fatal("empty provider list accepted")
	}
}
func TestYAMLSecretError(t *testing.T) {
	path := yamlFixture(t, "kind: gemini\nmodel: test\nthreshold: secret-value\n")
	_, e := readConfig(path)
	if e == nil || strings.Contains(e.Error(), "secret-value") {
		t.Fatal("secret leaked in parse error")
	}
}
func TestHostKeyPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "host_key")
	if e := ensureHostKey(path); e != nil {
		t.Fatal(e)
	}
	first, _ := os.ReadFile(path)
	if e := ensureHostKey(path); e != nil {
		t.Fatal(e)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Fatal("host key changed")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("private key permissions")
	}
}

func TestLabAndMetricsYAML(t *testing.T) {
	path := yamlFixture(t, "kind: ollama\nmodel: test\nendpoint: http://example.test\nthreshold: 30\n")
	c, err := readConfig(path)
	if err != nil || c.AcceptAllKeys || c.LogCommands || c.VerboseReviewLogs || c.MetricsListen != "" {
		t.Fatal("lab mode or metrics enabled by default", err)
	}
	if err := os.WriteFile(path, []byte("verbose_review_logs: true\nlog_commands: true\naccept_all_keys: true\nmetrics_listen: 127.0.0.1:2112\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err = readConfig(path)
	if err != nil || !c.AcceptAllKeys || !c.LogCommands || !c.VerboseReviewLogs || c.MetricsListen != "127.0.0.1:2112" {
		t.Fatal("lab and metrics settings not loaded", err)
	}
}
