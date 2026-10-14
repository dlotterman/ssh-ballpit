package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func ask(ctx context.Context, provider Provider, user, command string) (*Verdict, error) {
	return askWithHistory(ctx, provider, user, command, recentHistory{})
}

func askWithHistory(ctx context.Context, provider Provider, user, command string, history recentHistory) (*Verdict, error) {
	return askWithDiagnostics(ctx, provider, user, command, history, false)
}

func askWithDiagnostics(ctx context.Context, provider Provider, user, command string, history recentHistory, verbose bool) (verdict *Verdict, resultErr error) {
	diagnostic := reviewDiagnostic{
		enabled:  verbose,
		provider: provider,
		user:     user,
		command:  command,
		started:  time.Now(),
		stage:    "request_setup",
	}
	defer func() { diagnostic.log(resultErr) }()

	if provider.Kind == "openai-subscription" {
		diagnostic.stage = "subscription_review"
		return askCodexWithHistory(ctx, provider, user, command, history)
	}
	payload := marshalReviewPayload(user, command, history)
	messages := []any{
		map[string]string{"role": "system", "content": instructions},
		map[string]string{"role": "user", "content": string(payload)},
	}
	endpoint := strings.TrimRight(provider.Endpoint, "/")
	var body any
	headers := map[string]string{}
	if provider.Kind != "ollama" {
		key, err := providerKey(provider)
		if err != nil {
			return nil, err
		}
		if key == "" {
			return nil, fmt.Errorf("missing credential environment variable")
		}
		diagnostic.secrets = []string{key}
		headers["Authorization"] = "Bearer " + key
	}
	switch provider.Kind {
	case "ollama":
		endpoint += "/api/chat"
		body = map[string]any{
			"model":    provider.Model,
			"messages": messages,
			"stream":   false,
			"format":   verdictSchema(),
			"options":  map[string]any{"temperature": 0},
		}
	case "deepseek":
		endpoint += "/chat/completions"
		body = map[string]any{
			"model":           provider.Model,
			"messages":        messages,
			"response_format": map[string]string{"type": "json_object"},
		}
	case "openai":
		endpoint += "/responses"
		body = map[string]any{
			"model":        provider.Model,
			"instructions": instructions,
			"input":        string(payload),
			"store":        false,
			"text":         map[string]any{"format": map[string]string{"type": "json_object"}},
		}
	case "anthropic":
		endpoint += "/messages"
		headers["x-api-key"] = strings.TrimPrefix(headers["Authorization"], "Bearer ")
		delete(headers, "Authorization")
		headers["anthropic-version"] = "2023-06-01"
		body = map[string]any{
			"model":      provider.Model,
			"system":     instructions,
			"messages":   messages[1:],
			"max_tokens": 1024,
		}
	case "gemini":
		endpoint += "/models/" + url.PathEscape(provider.Model) + ":generateContent"
		headers["x-goog-api-key"] = strings.TrimPrefix(headers["Authorization"], "Bearer ")
		delete(headers, "Authorization")
		body = map[string]any{
			"systemInstruction": map[string]any{
				"parts": []any{map[string]string{"text": instructions}},
			},
			"contents": []any{map[string]any{
				"role":  "user",
				"parts": []any{map[string]string{"text": string(payload)}},
			}},
			"generationConfig": map[string]any{"responseMimeType": "application/json"},
		}
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode provider request: %w", err)
	}
	diagnostic.endpoint = endpoint
	diagnostic.request = requestBody
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	diagnostic.stage = "http_request"
	resp, err := requestWithGeminiRetry(client, req, provider.Kind)
	if err != nil {
		diagnostic.transportReason = connectionReason(err)
		return nil, fmt.Errorf("provider request failed")
	}
	defer resp.Body.Close()
	diagnostic.status = resp.StatusCode
	diagnostic.stage = "response_body"
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1048577))
	diagnostic.response = responseBody
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("provider HTTP %d", resp.StatusCode)
	}
	if err != nil || len(responseBody) > 1048576 {
		return nil, fmt.Errorf("invalid provider response size")
	}

	diagnostic.stage = "response_envelope"
	answer, err := providerAnswer(provider.Kind, responseBody)
	if err != nil {
		return nil, err
	}
	diagnostic.stage = "verdict_parse"
	diagnostic.answerBytes = len(answer)
	if provider.Kind == "ollama" {
		answer, diagnostic.unwrapped = unwrapOllamaVerdict(answer)
	}
	return parseVerdict(answer)
}

// providerAnswer unwraps each API's response and checks its completion marker.
// The returned text still needs strict verdict validation before it can approve anything.
func providerAnswer(kind string, data []byte) (string, error) {
	var envelope struct {
		Done       bool   `json:"done"`
		DoneReason string `json:"done_reason"`
		Status     string `json:"status"`
		StopReason string `json:"stop_reason"`
		Message    struct {
			Content string `json:"content"`
		} `json:"message"`
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Candidates []struct {
			FinishReason string `json:"finishReason"`
			Content      struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if json.Unmarshal(data, &envelope) != nil {
		return "", fmt.Errorf("invalid provider response")
	}
	var answer string
	switch kind {
	case "ollama":
		if !envelope.Done || (envelope.DoneReason != "" && envelope.DoneReason != "stop") {
			return "", fmt.Errorf("incomplete provider response")
		}
		answer = envelope.Message.Content
	case "deepseek":
		if len(envelope.Choices) == 1 && envelope.Choices[0].FinishReason == "stop" {
			answer = envelope.Choices[0].Message.Content
		}
	case "anthropic":
		if envelope.StopReason != "end_turn" {
			return "", fmt.Errorf("incomplete provider response")
		}
		for _, v := range envelope.Content {
			if v.Type == "text" {
				answer += v.Text
			}
		}
	case "gemini":
		if len(envelope.Candidates) == 1 && envelope.Candidates[0].FinishReason == "STOP" {
			for _, v := range envelope.Candidates[0].Content.Parts {
				answer += v.Text
			}
		}
	case "openai":
		if envelope.Status != "completed" {
			return "", fmt.Errorf("incomplete provider response")
		}
		for _, v := range envelope.Output {
			for _, t := range v.Content {
				if t.Type == "output_text" {
					answer += t.Text
				}
			}
		}
	}
	return answer, nil
}

// Retry temporary Gemini overload without switching models or extending the
// caller's review deadline. Never retry a verdict or a malformed response.
func requestWithGeminiRetry(client *http.Client, req *http.Request, kind string) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := client.Do(req)
		if err != nil || kind != "gemini" || resp.StatusCode != http.StatusServiceUnavailable || attempt == 2 {
			return resp, err
		}
		resp.Body.Close()
		delay := time.Second*time.Duration(1<<attempt) + time.Duration(rand.IntN(250))*time.Millisecond
		timer := time.NewTimer(delay)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
		next := req.Clone(req.Context())
		next.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
		req = next
	}
}
