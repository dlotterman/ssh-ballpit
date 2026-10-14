package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const diagnosticResponseLimit = 16384

func verdictSchema() map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"offensive", "danger_score", "reasons", "destinations"},
		"properties": map[string]any{
			"offensive":    map[string]any{"type": "boolean"},
			"danger_score": map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
			"reasons":      map[string]any{"type": "array", "items": map[string]string{"type": "string"}},
			"destinations": map[string]any{"type": "array", "items": map[string]string{"type": "string"}},
		},
	}
}

type reviewDiagnostic struct {
	unwrapped                                       bool
	enabled                                         bool
	provider                                        Provider
	user, command, endpoint, stage, transportReason string
	started                                         time.Time
	request, response                               []byte
	status, answerBytes                             int
	secrets                                         []string
}

// Quote each shell argument, including embedded quotes, substitutions and newlines.
func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

func (d reviewDiagnostic) log(err error) {
	if !d.enabled {
		return
	}
	response := string(d.response)
	for _, secret := range d.secrets {
		if secret == "" {
			continue
		}
		response = strings.ReplaceAll(response, secret, "[REDACTED]")
		escaped, _ := json.Marshal(secret)
		response = strings.ReplaceAll(response, string(escaped[1:len(escaped)-1]), "[REDACTED]")
	}
	truncated := len(response) > diagnosticResponseLimit
	if truncated {
		response = response[:diagnosticResponseLimit]
	}
	fields := []any{
		"reviewer", d.provider.Name, "kind", d.provider.Kind, "model", d.provider.Model,
		"endpoint_origin", endpointOrigin(d.provider.Endpoint),
		"user", d.user, "command_sha256", fmt.Sprintf("%x", sha256.Sum256([]byte(d.command))),
		"stage", d.stage, "duration_ms", time.Since(d.started).Milliseconds(),
		"http_status", d.status, "response_bytes", len(d.response),
		"response_excerpt", response, "response_truncated", truncated,
		"verdict_bytes", d.answerBytes, "ollama_markdown_unwrapped", d.unwrapped, "transport_reason", d.transportReason,
	}
	if d.provider.Kind == "ollama" && len(d.request) > 0 {
		// No credential headers are sent by the Ollama adapter. Pass data on stdin
		// so curl cannot interpret a leading @ as a filename.
		curl := "printf '%s' " + shellQuote(string(d.request)) +
			" | curl --silent --show-error --max-time 120 --request POST --header 'Content-Type: application/json' --data-binary @- --url " + shellQuote(d.endpoint)
		fields = append(fields, "ollama_curl", curl)
	}
	if err != nil {
		fields = append(fields, "error", err.Error())
		slog.Error("provider review diagnostic", fields...)
	} else {
		slog.Info("provider review diagnostic", fields...)
	}
}
