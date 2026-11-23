package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProvidersCheckCallsEveryProvider(t *testing.T) {
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		// A valid denial is still a successful connectivity and authentication check.
		json.NewEncoder(w).Encode(map[string]any{
			"done":    true,
			"message": map[string]string{"content": `{"offensive":false,"danger_score":90,"reasons":[],"destinations":[]}`},
		})
	}))
	defer provider.Close()

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "providers"), 0700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "server.yaml")
	if err := os.WriteFile(configPath, []byte("users: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"first", "second"} {
		data := []byte("kind: ollama\nmodel: test\nthreshold: 30\nendpoint: " + provider.URL + "\n")
		if err := os.WriteFile(filepath.Join(dir, "providers", name+".yaml"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Local command rules must not suppress the live provider probe.
	if err := os.WriteFile(filepath.Join(dir, "blacklist.yaml"), []byte("patterns: ['.*']\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"providers-check", "--config", configPath}, &output); err != nil {
		t.Fatal(err)
	}
	var response providersCheckOutput
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.OK || len(response.Providers) != 2 || !response.Providers[0].OK || !response.Providers[1].OK || calls.Load() != 2 {
		t.Fatalf("unexpected provider check: %s; calls=%d", output.String(), calls.Load())
	}
	if strings.Contains(output.String(), "danger_score") {
		t.Fatal("provider verdict leaked into connectivity check")
	}
}

func TestProvidersCheckReportsAuthenticationFailure(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer provider.Close()
	configPath := yamlFixture(t, "kind: ollama\nmodel: test\nthreshold: 30\nendpoint: "+provider.URL+"\n")
	var output bytes.Buffer
	if err := run([]string{"providers-check", "--config", configPath}, &output); err == nil {
		t.Fatal("authentication failure returned success")
	}
	var response providersCheckOutput
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.OK || len(response.Providers) != 1 || response.Providers[0].OK || response.Providers[0].Error != "provider HTTP 401" {
		t.Fatalf("unexpected provider check: %s", output.String())
	}
}
