package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Verdict uses pointers so missing fields differ from valid false/zero values.
type Verdict struct {
	Offensive    *bool    `json:"offensive"`
	Score        *int     `json:"danger_score"`
	Reasons      []string `json:"reasons"`
	Destinations []string `json:"destinations"`
}

type Result struct {
	Name    string   `json:"name"`
	Verdict *Verdict `json:"verdict,omitempty"`
	Error   string   `json:"error,omitempty"`
	Safe    bool     `json:"safe"`
}

func parseVerdict(content string) (*Verdict, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty verdict content")
	}
	// Reject duplicate fields as well as missing, unknown and incorrectly typed fields.
	check := json.NewDecoder(strings.NewReader(content))
	token, err := check.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid verdict JSON: expected an object at the start of content")
	}
	if token != json.Delim('{') {
		return nil, fmt.Errorf("invalid verdict object: expected a JSON object")
	}
	seen := map[string]bool{}
	for check.More() {
		key, err := check.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid verdict")
		}
		name, ok := key.(string)
		if !ok || seen[name] {
			return nil, fmt.Errorf("duplicate verdict field")
		}
		seen[name] = true
		var raw json.RawMessage
		if check.Decode(&raw) != nil {
			return nil, fmt.Errorf("invalid verdict")
		}
	}
	var verdict Verdict
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&verdict); err != nil {
		return nil, fmt.Errorf("invalid verdict JSON")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("incomplete or invalid verdict")
	}
	if verdict.Offensive == nil || verdict.Score == nil || verdict.Reasons == nil || verdict.Destinations == nil {
		return nil, fmt.Errorf("incomplete or invalid verdict")
	}
	if *verdict.Score < 0 || *verdict.Score > 100 {
		return nil, fmt.Errorf("incomplete or invalid verdict")
	}
	return &verdict, nil
}

// Some Ollama-compatible servers ignore format and return a single Markdown
// fence. Remove only an exact outer wrapper; strict JSON validation still runs.
func unwrapOllamaVerdict(answer string) (string, bool) {
	trimmed := strings.TrimSpace(answer)
	for _, prefix := range []string{"```json\n", "```\n", "```json\r\n", "```\r\n"} {
		if strings.HasPrefix(trimmed, prefix) && strings.HasSuffix(trimmed, "\n```") {
			return strings.TrimSuffix(strings.TrimPrefix(trimmed, prefix), "\n```"), true
		}
	}
	return answer, false
}
