package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProviderAdapters(t *testing.T) {
	// Substitute the trusted test server's transport only for these sequential tests.
	old := http.DefaultTransport
	defer func() { http.DefaultTransport = old }()
	for _, kind := range []string{"openai", "anthropic", "deepseek", "gemini"} {
		t.Run(kind, func(t *testing.T) {
			t.Setenv("TEST_API_KEY", "test-credential")
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var input map[string]any
				if json.NewDecoder(r.Body).Decode(&input) != nil {
					t.Error("invalid request")
				}
				if r.Method != "POST" {
					t.Error("not POST")
				}
				var body any
				switch kind {
				case "openai":
					if r.URL.Path != "/responses" || r.Header.Get("Authorization") != "Bearer test-credential" || input["instructions"] == nil {
						t.Error("bad OpenAI request")
					}
					body = map[string]any{"status": "completed", "output": []any{map[string]any{"content": []any{map[string]string{"type": "output_text", "text": safeJSON}}}}}
				case "anthropic":
					if r.URL.Path != "/messages" || r.Header.Get("x-api-key") != "test-credential" || input["system"] == nil {
						t.Error("bad Anthropic request")
					}
					body = map[string]any{"stop_reason": "end_turn", "content": []any{map[string]string{"type": "text", "text": safeJSON}}}
				case "deepseek":
					if r.URL.Path != "/chat/completions" || r.Header.Get("Authorization") != "Bearer test-credential" {
						t.Error("bad DeepSeek request")
					}
					body = map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]string{"content": safeJSON}}}}
				case "gemini":
					if r.URL.Path != "/models/test:generateContent" || r.Header.Get("x-goog-api-key") != "test-credential" || input["systemInstruction"] == nil {
						t.Error("bad Gemini request")
					}
					body = map[string]any{"candidates": []any{map[string]any{"finishReason": "STOP", "content": map[string]any{"parts": []any{map[string]string{"text": safeJSON}}}}}}
				}
				var payloadText string
				switch kind {
				case "openai":
					payloadText = input["input"].(string)
				case "anthropic":
					payloadText = input["messages"].([]any)[0].(map[string]any)["content"].(string)
				case "deepseek":
					payloadText = input["messages"].([]any)[1].(map[string]any)["content"].(string)
				case "gemini":
					payloadText = input["contents"].([]any)[0].(map[string]any)["parts"].([]any)[0].(map[string]any)["text"].(string)
				}
				var payload reviewPayload
				if err := json.Unmarshal([]byte(payloadText), &payload); err != nil || !reflect.DeepEqual(payload.RecentHistory, sampleHistory()) || payload.Command != "pwd" {
					t.Error("history missing or altered in provider request", err)
				}
				json.NewEncoder(w).Encode(body)
			}))
			defer s.Close()
			http.DefaultTransport = s.Client().Transport
			v, e := askWithHistory(context.Background(), Provider{Kind: kind, Endpoint: s.URL, KeyEnv: "TEST_API_KEY", Model: "test"}, "alice", "pwd", sampleHistory())
			if e != nil || *v.Score != 0 {
				t.Fatalf("verdict %v error %v", v, e)
			}
		})
	}
}
func TestDuplicateVerdict(t *testing.T) {
	if _, e := parseVerdict(`{"offensive":true,"offensive":false,"danger_score":0,"reasons":[],"destinations":[]}`); e == nil {
		t.Fatal("accepted duplicate")
	}
}

func TestGeminiOverloadRetries(t *testing.T) {
	for _, tc := range []struct {
		name, kind                  string
		status, failures, wantCalls int
		success                     bool
	}{
		{"recovers", "gemini", 503, 1, 2, true},
		{"exhausted", "gemini", 503, 10, 3, false},
		{"bad_credentials", "gemini", 403, 10, 1, false},
		{"other_provider", "openai", 503, 10, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				var input map[string]any
				if json.NewDecoder(r.Body).Decode(&input) != nil || input["systemInstruction"] == nil {
					t.Error("retry lost request body")
				}
				if r.Header.Get("x-goog-api-key") != "test-key" {
					t.Error("retry lost credentials")
				}
				if calls <= tc.failures {
					w.WriteHeader(tc.status)
					w.Write([]byte("SECRET response body"))
					return
				}
				json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"finishReason": "STOP", "content": map[string]any{"parts": []any{map[string]string{"text": safeJSON}}}}}})
			}))
			defer s.Close()
			// Use the Gemini wire format even when checking the retry policy for another kind.
			req, _ := http.NewRequest("POST", s.URL, strings.NewReader(`{"systemInstruction":{}}`))
			req.Header.Set("x-goog-api-key", "test-key")
			resp, err := requestWithGeminiRetry(s.Client(), req, tc.kind)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if (resp.StatusCode == 200) != tc.success || calls != tc.wantCalls {
				t.Fatalf("status=%d calls=%d", resp.StatusCode, calls)
			}
		})
	}
}

func TestGeminiRetryHonorsDeadline(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(503) }))
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", s.URL, strings.NewReader(`{}`))
	_, err := requestWithGeminiRetry(s.Client(), req, "gemini")
	if !errors.Is(err, context.DeadlineExceeded) || calls != 1 {
		t.Fatalf("error=%v calls=%d", err, calls)
	}
}
