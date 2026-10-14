package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func diagnosticEvents(t *testing.T, logs string) []map[string]any {
	t.Helper()
	var events []map[string]any
	for _, line := range strings.Split(logs, "\n") {
		var event map[string]any
		if json.Unmarshal([]byte(line), &event) == nil && event["msg"] == "provider review diagnostic" {
			events = append(events, event)
		}
	}
	return events
}

func TestOllamaDiagnosticsAndReplay(t *testing.T) {
	logs := captureLogs(t)
	var requests [][]byte
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests = append(requests, body)
		var input map[string]any
		if err := json.Unmarshal(body, &input); err != nil {
			t.Error(err)
		}
		format, ok := input["format"].(map[string]any)
		if !ok || format["type"] != "object" {
			t.Error("missing verdict schema")
		}
		json.NewEncoder(w).Encode(map[string]any{"done": true, "done_reason": "stop", "message": map[string]string{"content": "Here is the verdict:\n" + safeJSON}})
	}))
	defer s.Close()
	marker := filepath.Join(t.TempDir(), "must-not-exist")
	command := "echo 'quoted'; $(touch " + marker + ") `touch " + marker + "`\nnext"
	p := Provider{Name: "ollama", Kind: "ollama", Endpoint: s.URL, Model: "test", Threshold: 30}
	for _, verbose := range []bool{false, true} {
		c := Config{Providers: []Provider{p}, ReviewSeconds: 5, VerboseReviewLogs: verbose}
		results, safe := reviewWithHistory(context.Background(), c, "alice", command, sampleHistory())
		if safe || !strings.Contains(results[0].Error, "invalid verdict JSON") {
			t.Fatal("malformed verdict did not fail closed", results)
		}
		if !verbose && strings.Contains(logs.String(), "provider review diagnostic") {
			t.Fatal("verbose logging enabled by default")
		}
	}
	events := diagnosticEvents(t, logs.String())
	if len(events) != 1 {
		t.Fatalf("events=%d", len(events))
	}
	event := events[0]
	if event["stage"] != "verdict_parse" || event["level"] != "ERROR" || event["http_status"] != float64(200) || !strings.Contains(event["response_excerpt"].(string), "Here is the verdict:") {
		t.Fatal("missing diagnostic context", event)
	}
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl not installed for replay check")
	}
	if out, err := exec.Command("sh", "-c", event["ollama_curl"].(string)).CombinedOutput(); err != nil {
		t.Fatalf("replay failed: %v %s", err, out)
	}
	if len(requests) != 3 || string(requests[1]) != string(requests[2]) {
		t.Fatal("replay changed request payload")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("shell evaluated payload")
	}
}

func TestDiagnosticRedactionAndLimit(t *testing.T) {
	logs := captureLogs(t)
	secret := "secret-with-quote-\"-and-slash-\\"
	escaped, _ := json.Marshal(secret)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte(secret + " " + string(escaped) + strings.Repeat("x", diagnosticResponseLimit*2)))
	}))
	defer s.Close()
	p := Provider{Name: "test", Kind: "openai", Endpoint: s.URL, Model: "test", APIKey: secret}
	_, err := askWithDiagnostics(context.Background(), p, "alice", "pwd", recentHistory{}, true)
	if err == nil || err.Error() != "provider HTTP 503" {
		t.Fatal(err)
	}
	events := diagnosticEvents(t, logs.String())
	if len(events) != 1 {
		t.Fatal("missing diagnostic")
	}
	e := events[0]
	excerpt := e["response_excerpt"].(string)
	if strings.Contains(excerpt, secret) || strings.Contains(excerpt, string(escaped)) || !strings.Contains(excerpt, "[REDACTED]") || len(excerpt) > diagnosticResponseLimit || e["response_truncated"] != true {
		t.Fatal("unsafe response excerpt")
	}
	if _, ok := e["ollama_curl"]; ok {
		t.Fatal("non-Ollama replay included")
	}
}

func TestOllamaIncompleteResponses(t *testing.T) {
	for _, tc := range []struct{ name, content, reason, want string }{
		{"empty", "", "stop", "empty verdict content"},
		{"array", "[]", "stop", "invalid verdict object"},
		{"truncated", safeJSON, "length", "incomplete provider response"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(map[string]any{"done": true, "done_reason": tc.reason, "message": map[string]string{"content": tc.content}})
			}))
			defer s.Close()
			_, err := ask(context.Background(), Provider{Kind: "ollama", Endpoint: s.URL, Model: "test"}, "alice", "pwd")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}

func TestOllamaFencedVerdicts(t *testing.T) {
	for _, tc := range []struct {
		name, answer string
		valid        bool
	}{
		{"plain", safeJSON, true},
		{"json_fence", "```json\n" + safeJSON + "\n```", true},
		{"bare_fence", "```\n" + safeJSON + "\n```", true},
		{"crlf", "```json\r\n" + safeJSON + "\r\n```", true},
		{"prose", "Here is JSON:\n```json\n" + safeJSON + "\n```", false},
		{"extra_object", "```json\n" + safeJSON + " {}\n```", false},
		{"extra_fence", "```json\n" + safeJSON + "\n```\n```json\n" + safeJSON + "\n```", false},
		{"missing_close", "```json\n" + safeJSON, false},
		{"missing_fields", "```json\n{}\n```", false},
		{"duplicate", "```json\n" + strings.Replace(safeJSON, `"offensive":false`, `"offensive":true,"offensive":false`, 1) + "\n```", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			answer, _ := unwrapOllamaVerdict(tc.answer)
			_, err := parseVerdict(answer)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t error=%v", tc.valid, err)
			}
		})
	}
	// Unwrapping must preserve an actual denial, rather than turn it into approval.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": "```json\n" + strings.Replace(safeJSON, `"danger_score":0`, `"danger_score":90`, 1) + "\n```"}})
	}))
	defer s.Close()
	results, safe := review(context.Background(), Config{ReviewSeconds: 5, Providers: []Provider{{Name: "ollama", Kind: "ollama", Endpoint: s.URL, Model: "test", Threshold: 30}}}, "alice", "pwd")
	if safe || results[0].Error != "" || *results[0].Verdict.Score != 90 {
		t.Fatal("fenced denial lost", results)
	}
}
