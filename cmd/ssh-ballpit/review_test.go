package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const safeJSON = `{"offensive":false,"danger_score":0,"reasons":[],"destinations":[]}`

func TestVerdictValidation(t *testing.T) {
	for _, s := range []string{`{}`, `null`, `{"offensive":false,"danger_score":0}`, `{"offensive":false,"danger_score":101,"reasons":[],"destinations":[]}`, safeJSON + ` {}`, "```json\n" + safeJSON + "\n```"} {
		if _, e := parseVerdict(s); e == nil {
			t.Errorf("accepted %s", s)
		}
	}
	if _, e := parseVerdict(safeJSON); e != nil {
		t.Fatal(e)
	}
}
func TestUnanimousReview(t *testing.T) {
	for _, tc := range []struct {
		name, answer string
		status       int
		safe         bool
	}{
		{"safe", safeJSON, 200, true},
		{"threshold", `{"offensive":false,"danger_score":30,"reasons":[],"destinations":[]}`, 200, false},
		{"offensive", `{"offensive":true,"danger_score":0,"reasons":[],"destinations":[]}`, 200, false},
		{"malformed", `{}`, 200, false}, {"unavailable", "", 503, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Path == "/bad/api/chat" {
					w.WriteHeader(tc.status)
					json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": tc.answer}})
				} else {
					json.NewEncoder(w).Encode(map[string]any{"done": true, "message": map[string]string{"content": safeJSON}})
				}
			}))
			defer s.Close()
			c := Config{ReviewSeconds: 2, Providers: []Provider{{Name: "good", Kind: "ollama", Endpoint: s.URL, Model: "test", Threshold: 30}, {Name: "other", Kind: "ollama", Endpoint: s.URL + "/bad", Model: "test", Threshold: 30}}}
			c.metrics = newDaemonMetrics(c)
			_, safe := review(context.Background(), c, "alice", "pwd")
			if safe != tc.safe || calls.Load() != 2 {
				t.Fatalf("safe=%v calls=%d", safe, calls.Load())
			}
			outcome := "denied"
			if tc.safe {
				outcome = "approved"
			} else if tc.name == "malformed" || tc.name == "unavailable" {
				outcome = "error"
			}
			recorder := httptest.NewRecorder()
			c.metrics.handler().ServeHTTP(recorder, httptest.NewRequest("GET", "/metrics", nil))
			want := `ssh_ballpit_reviews_total{outcome="` + outcome + `",reviewer="other"} 1`
			if !strings.Contains(recorder.Body.String(), want) {
				t.Fatalf("missing review metric %s", want)
			}
		})
	}
}
func TestReviewCancellation(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, safe := review(ctx, Config{ReviewSeconds: 1, Providers: []Provider{{Kind: "ollama", Endpoint: s.URL, Threshold: 30}}}, "alice", "pwd")
	if safe {
		t.Fatal("approved cancelled review")
	}
}
func TestPolicy(t *testing.T) {
	for _, cmd := range []string{"bash", "python3 -c 'print(1)'", "echo $(id)", "/tmp/script", "$SHELL", "env bash", "echo <(id)", ""} {
		if policy(cmd) == nil {
			t.Errorf("accepted %q", cmd)
		}
	}
	for _, cmd := range []string{"pwd", "uname -a", "printf '%s\\n' hello", "ls -l | head -n 3"} {
		if e := policy(cmd); e != nil {
			t.Errorf("%s: %v", cmd, e)
		}
	}
}
