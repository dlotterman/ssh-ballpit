package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func codexDirectory(home string, create bool) (string, error) {
	if home == "" {
		return "", fmt.Errorf("codex_home is required")
	}
	home, err := filepath.Abs(home)
	if err != nil {
		return "", err
	}
	if create {
		if err = os.MkdirAll(home, 0700); err != nil {
			return "", err
		}
	}
	info, err := os.Lstat(home)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return "", fmt.Errorf("Codex home must be a private directory (mode 0700)")
	}
	// This directory is owned by ssh-ballpit, not a user's general Codex installation.
	for _, name := range []string{"config.toml", "AGENTS.md", "hooks.json", "plugins"} {
		if _, err = os.Lstat(filepath.Join(home, name)); !os.IsNotExist(err) {
			return "", fmt.Errorf("dedicated Codex home must not contain %s", name)
		}
	}
	// Codex itself creates skills/.system, even for an isolated app-server.
	// Permit that generated cache, but not custom skills or symlinks into a
	// user's general Codex installation. Reviews still use no environments or
	// selected capability roots and reject inherited instruction sources.
	skills := filepath.Join(home, "skills")
	if info, err := os.Lstat(skills); !os.IsNotExist(err) {
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("Codex skills cache must be a directory, not a symlink")
		}
		entries, err := os.ReadDir(skills)
		if err != nil {
			return "", fmt.Errorf("cannot inspect Codex skills cache")
		}
		for _, entry := range entries {
			if entry.Name() != ".system" || !entry.IsDir() {
				return "", fmt.Errorf("dedicated Codex home must not contain custom skills (only generated skills/.system is allowed)")
			}
		}
		if err := filepath.WalkDir(skills, func(_ string, entry os.DirEntry, err error) error {
			if err != nil || entry.Type()&os.ModeSymlink != 0 || (!entry.IsDir() && !entry.Type().IsRegular()) {
				return fmt.Errorf("Codex skills cache must contain only regular files and directories")
			}
			return nil
		}); err != nil {
			return "", err
		}
	}
	return home, nil
}

func codexCommand(ctx context.Context, binary, home string, args ...string) *exec.Cmd {
	if binary == "" {
		binary = "codex"
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	// Do not pass API keys, provider overrides or the operator's Codex environment.
	for _, name := range []string{"PATH", "HOME", "LANG", "SSL_CERT_FILE", "SSL_CERT_DIR", "HTTPS_PROXY", "HTTP_PROXY", "NO_PROXY"} {
		if v, ok := os.LookupEnv(name); ok {
			cmd.Env = append(cmd.Env, name+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, "CODEX_HOME="+home)
	cmd.Dir = home
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

func login(ctx context.Context, home, binary string) error {
	home, err := codexDirectory(home, true)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	cmd := codexCommand(ctx, binary, home, "login", "--device-auth", "-c", `cli_auth_credentials_store="file"`)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Codex device login failed: %w", err)
	}
	return nil
}

type rpcMessage struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
}

type rpcClient struct {
	input  *json.Encoder
	output *bufio.Scanner
	next   int
}

func (r *rpcClient) send(method string, params any) (int, error) {
	r.next++
	return r.next, r.input.Encode(map[string]any{"id": r.next, "method": method, "params": params})
}

func (r *rpcClient) read() (rpcMessage, error) {
	var message rpcMessage
	if !r.output.Scan() {
		return message, fmt.Errorf("Codex stream ended or exceeded size limit")
	}
	if json.Unmarshal(r.output.Bytes(), &message) != nil {
		return message, fmt.Errorf("invalid Codex response")
	}
	if message.Method != "" && len(message.ID) > 0 {
		return message, fmt.Errorf("Codex requested a tool or approval; review denied")
	}
	return message, nil
}

func (r *rpcClient) call(method string, params any) (json.RawMessage, error) {
	id, err := r.send(method, params)
	if err != nil {
		return nil, err
	}
	for {
		message, err := r.read()
		if err != nil {
			return nil, err
		}
		if string(message.ID) == fmt.Sprint(id) {
			if len(message.Error) > 0 {
				return nil, fmt.Errorf("Codex %s failed", method)
			}
			return message.Result, nil
		}
	}
}

var codexReviewSlot = make(chan struct{}, 1)

func askCodex(ctx context.Context, provider Provider, user, command string) (*Verdict, error) {
	return askCodexWithHistory(ctx, provider, user, command, recentHistory{})
}

func askCodexWithHistory(ctx context.Context, provider Provider, user, command string, history recentHistory) (*Verdict, error) {
	// Serialize processes so rotating credentials have a single owner at a time.
	select {
	case codexReviewSlot <- struct{}{}:
		defer func() { <-codexReviewSlot }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	home, err := codexDirectory(provider.CodexHome, false)
	if err != nil {
		return nil, err
	}
	args := []string{
		"app-server", "--listen", "stdio://", "--strict-config",
		"-c", `cli_auth_credentials_store="file"`,
		"-c", `forced_login_method="chatgpt"`,
		"-c", `web_search="disabled"`,
		"-c", "project_doc_max_bytes=0",
	}
	for _, feature := range []string{"shell_tool", "shell_snapshot", "apps", "plugins", "hooks", "multi_agent", "browser_use", "in_app_browser", "image_generation", "view_image", "memories", "skill_search", "code_mode_host"} {
		args = append(args, "--disable", feature)
	}
	cmd := codexCommand(ctx, provider.CodexBinary, home, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("cannot start Codex: %w", err)
	}
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()
	scanner := bufio.NewScanner(io.LimitReader(stdout, 4<<20))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	rpc := rpcClient{input: json.NewEncoder(stdin), output: scanner}
	if _, err = rpc.call("initialize", map[string]any{
		"clientInfo":   map[string]string{"name": "ssh-ballpit", "version": "0.1.0"},
		"capabilities": map[string]bool{"experimentalApi": true},
	}); err != nil {
		return nil, err
	}
	if err = rpc.input.Encode(map[string]string{"method": "initialized"}); err != nil {
		return nil, err
	}
	account, err := rpc.call("account/read", map[string]bool{"refreshToken": true})
	if err != nil {
		return nil, err
	}
	var accountResponse struct {
		Account *struct {
			Type string `json:"type"`
		} `json:"account"`
	}
	if json.Unmarshal(account, &accountResponse) != nil || accountResponse.Account == nil || accountResponse.Account.Type != "chatgpt" {
		return nil, fmt.Errorf("Codex requires ChatGPT device login")
	}
	result, err := rpc.call("thread/start", map[string]any{
		"model":                   provider.Model,
		"modelProvider":           "openai",
		"cwd":                     home,
		"ephemeral":               true,
		"approvalPolicy":          "never",
		"sandbox":                 "read-only",
		"baseInstructions":        instructions,
		"developerInstructions":   "Return the verdict only. Do not call tools.",
		"environments":            []any{},
		"dynamicTools":            []any{},
		"selectedCapabilityRoots": []any{},
	})
	if err != nil {
		return nil, err
	}
	var thread struct {
		Thread struct {
			ID           string            `json:"id"`
			Environments []json.RawMessage `json:"environments"`
		} `json:"thread"`
		InstructionSources []json.RawMessage `json:"instructionSources"`
		Sandbox            struct {
			Type          string `json:"type"`
			NetworkAccess bool   `json:"networkAccess"`
		} `json:"sandbox"`
	}
	if json.Unmarshal(result, &thread) != nil || thread.Thread.ID == "" {
		return nil, fmt.Errorf("missing Codex thread")
	}
	if thread.Thread.Environments == nil || len(thread.Thread.Environments) != 0 || len(thread.InstructionSources) != 0 || thread.Sandbox.Type != "readOnly" || thread.Sandbox.NetworkAccess {
		return nil, fmt.Errorf("Codex did not establish an isolated review session")
	}
	payload := marshalReviewPayload(user, command, history)
	id, err := rpc.send("turn/start", map[string]any{
		"threadId": thread.Thread.ID,
		"input": []any{map[string]any{
			"type":          "text",
			"text":          string(payload),
			"text_elements": []any{},
		}},
		"environments": []any{},
	})
	if err != nil {
		return nil, err
	}
	var answer string
	started := false
	for {
		message, err := rpc.read()
		if err != nil {
			return nil, err
		}
		if string(message.ID) == fmt.Sprint(id) {
			if len(message.Error) > 0 {
				return nil, fmt.Errorf("Codex turn failed")
			}
			started = true
			continue
		}
		switch message.Method {
		case "item/started", "item/completed":
			var event struct {
				ThreadID string `json:"threadId"`
				Item     struct {
					Type  string `json:"type"`
					Text  string `json:"text"`
					Phase string `json:"phase"`
				} `json:"item"`
			}
			if json.Unmarshal(message.Params, &event) != nil || event.ThreadID != thread.Thread.ID {
				return nil, fmt.Errorf("invalid Codex item")
			}
			switch event.Item.Type {
			case "userMessage", "reasoning", "agentMessage":
			default:
				return nil, fmt.Errorf("Codex attempted a tool; review denied")
			}
			if message.Method == "item/completed" && event.Item.Type == "agentMessage" && (event.Item.Phase == "final_answer" || event.Item.Phase == "") {
				answer = event.Item.Text
			}
		case "turn/completed":
			var event struct {
				ThreadID string `json:"threadId"`
				Turn     struct {
					Status string `json:"status"`
				} `json:"turn"`
			}
			if json.Unmarshal(message.Params, &event) != nil || !started || event.ThreadID != thread.Thread.ID || event.Turn.Status != "completed" {
				return nil, fmt.Errorf("Codex review did not complete")
			}
			return parseVerdict(answer)
		case "error":
			return nil, fmt.Errorf("Codex inference failed")
		}
	}
}
