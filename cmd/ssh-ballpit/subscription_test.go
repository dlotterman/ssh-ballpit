package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// A separate Go process tests the actual subprocess/RPC boundary without using credentials.
func TestCodexProcess(t *testing.T) {
	args := strings.Join(os.Args, " ")
	if !strings.Contains(args, " -- ") {
		return
	}
	home := os.Getenv("CODEX_HOME")
	b, _ := os.ReadFile(filepath.Join(home, "scenario"))
	mode := string(b)
	if os.Getenv("OPENAI_API_KEY") != "" || os.Getenv("GEMINI_API_KEY") != "" {
		os.Exit(20)
	}
	if strings.Contains(args, "login --device-auth") {
		if !strings.Contains(args, `cli_auth_credentials_store="file"`) {
			os.Exit(21)
		}
		fmt.Println("Enter one-time code in your browser on another device")
		os.Exit(0)
	}
	if mode == "hang" {
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	if !strings.Contains(args, "--disable shell_tool") || !strings.Contains(args, "--disable plugins") {
		os.Exit(22)
	}
	sc := bufio.NewScanner(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for sc.Scan() {
		var m struct {
			ID     int            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if json.Unmarshal(sc.Bytes(), &m) != nil {
			os.Exit(23)
		}
		reply := func(v any) { enc.Encode(map[string]any{"id": m.ID, "result": v}) }
		switch m.Method {
		case "initialize":
			reply(map[string]any{})
		case "account/read":
			kind := "chatgpt"
			if mode == "apikey" {
				kind = "apiKey"
			}
			reply(map[string]any{"account": map[string]string{"type": kind}})
		case "thread/start":
			env, ok := m.Params["environments"].([]any)
			roots, rootsOK := m.Params["selectedCapabilityRoots"].([]any)
			if !ok || len(env) != 0 || !rootsOK || len(roots) != 0 || m.Params["ephemeral"] != true || m.Params["baseInstructions"] != instructions {
				os.Exit(24)
			}
			reply(map[string]any{"thread": map[string]any{"id": "thread", "environments": []any{}}, "instructionSources": []any{}, "sandbox": map[string]any{"type": "readOnly", "networkAccess": false}})
		case "turn/start":
			if mode == "history" {
				input := m.Params["input"].([]any)[0].(map[string]any)["text"].(string)
				var payload reviewPayload
				if json.Unmarshal([]byte(input), &payload) != nil || !reflect.DeepEqual(payload.RecentHistory, sampleHistory()) {
					os.Exit(25)
				}
			}
			reply(map[string]any{})
			if mode == "approval" {
				enc.Encode(map[string]any{"id": 99, "method": "item/commandExecution/requestApproval"})
				continue
			}
			text := safeJSON
			if mode == "invalid" {
				text = "{}"
			}
			kind := "agentMessage"
			if mode == "tool" {
				kind = "commandExecution"
			}
			enc.Encode(map[string]any{"method": "item/completed", "params": map[string]any{"threadId": "thread", "item": map[string]string{"type": kind, "text": text, "phase": "final_answer"}}})
			status := "completed"
			if mode == "failed" {
				status = "failed"
			}
			enc.Encode(map[string]any{"method": "turn/completed", "params": map[string]any{"threadId": "thread", "turn": map[string]string{"status": status}}})
		}
	}
	os.Exit(0)
}
func fakeCodex(t *testing.T, mode string) (string, string) {
	t.Helper()
	home := t.TempDir()
	os.Chmod(home, 0700)
	os.WriteFile(filepath.Join(home, "scenario"), []byte(mode), 0600)
	binary := filepath.Join(t.TempDir(), "codex")
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	script := "#!/bin/sh\nexec '" + strings.ReplaceAll(exe, "'", "'\\''") + "' -test.run=TestCodexProcess -- \"$@\"\n"
	if e = os.WriteFile(binary, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	return home, binary
}
func TestCodexReview(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "must-not-inherit")
	t.Setenv("GEMINI_API_KEY", "must-not-inherit")
	for _, mode := range []string{"safe", "invalid", "failed", "tool", "approval", "apikey", "hang"} {
		t.Run(mode, func(t *testing.T) {
			home, binary := fakeCodex(t, mode)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if mode == "hang" {
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}
			v, e := askCodex(ctx, Provider{CodexHome: home, CodexBinary: binary, Model: "test"}, "alice", "uname -a")
			if mode == "safe" {
				if e != nil || v == nil {
					t.Fatalf("%v", e)
				}
			} else if e == nil {
				t.Fatalf("accepted %s", mode)
			}
		})
	}
}
func TestDeviceLogin(t *testing.T) {
	home, binary := fakeCodex(t, "login")
	if e := login(context.Background(), home, binary); e != nil {
		t.Fatal(e)
	}
}
func TestCodexDirectory(t *testing.T) {
	home := t.TempDir()
	os.Chmod(home, 0755)
	if _, e := codexDirectory(home, false); e == nil {
		t.Fatal("accepted public directory")
	}
	os.Chmod(home, 0700)
	os.WriteFile(filepath.Join(home, "config.toml"), []byte(""), 0600)
	if _, e := codexDirectory(home, false); e == nil {
		t.Fatal("accepted inherited config")
	}
}

func TestCodexGeneratedSkillsCache(t *testing.T) {
	home, binary := fakeCodex(t, "safe")
	system := filepath.Join(home, "skills", ".system", "skill-creator")
	if err := os.MkdirAll(system, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(system, "SKILL.md"), []byte("generated skill"), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		v, err := askCodex(ctx, Provider{CodexHome: home, CodexBinary: binary, Model: "test"}, "alice", "pwd")
		cancel()
		if err != nil || v == nil {
			t.Fatalf("review %d: %v", i, err)
		}
	}
	if err := login(context.Background(), home, binary); err != nil {
		t.Fatal(err)
	}
}

func TestCodexRejectsCustomSkillsAndLinks(t *testing.T) {
	for _, kind := range []string{"custom", "skills_link", "system_link", "nested_link"} {
		t.Run(kind, func(t *testing.T) {
			home := t.TempDir()
			os.Chmod(home, 0700)
			skills := filepath.Join(home, "skills")
			var err error
			switch kind {
			case "custom":
				err = os.MkdirAll(filepath.Join(skills, "custom"), 0700)
			case "skills_link":
				err = os.Symlink(t.TempDir(), skills)
			case "system_link":
				os.Mkdir(skills, 0700)
				err = os.Symlink(t.TempDir(), filepath.Join(skills, ".system"))
			case "nested_link":
				os.MkdirAll(filepath.Join(skills, ".system"), 0700)
				err = os.Symlink(t.TempDir(), filepath.Join(skills, ".system", "outside"))
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := codexDirectory(home, false); err == nil {
				t.Fatal("accepted custom skills or symlink")
			}
		})
	}
}

func TestCodexHistoryPayload(t *testing.T) {
	home, binary := fakeCodex(t, "history")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if v, err := askCodexWithHistory(ctx, Provider{CodexHome: home, CodexBinary: binary, Model: "test"}, "alice", "pwd", sampleHistory()); err != nil || v == nil {
		t.Fatal("subscription history failed", err)
	}
}
