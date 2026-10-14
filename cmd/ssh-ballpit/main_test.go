package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRejectsInvalidInvocationBeforeReadingConfig(t *testing.T) {
	missingConfig := filepath.Join(t.TempDir(), "missing.yaml")
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"unknown command", []string{"typo", "--config", missingConfig}, "unknown subcommand"},
		{"unquoted command", []string{"check", "--config", missingConfig, "--command", "ls", "/home"}, "unexpected positional arguments"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := run(test.args, io.Discard)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v; want %q", err, test.want)
			}
		})
	}
}

func TestConfigureNestedJSONDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "server.JSON")
	args := []string{"configure", "--config", path}
	if err := run(args, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfig(path); err != nil {
		t.Fatalf("generated JSON cannot be loaded: %v", err)
	}

	// Protect the operator's subsequent edits when configure is run again.
	const edited = "operator's configuration"
	if err := os.WriteFile(path, []byte(edited), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(args, io.Discard); !errors.Is(err, os.ErrExist) {
		t.Fatalf("got %v; want existing-file error", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != edited {
		t.Fatalf("existing configuration changed: %q, %v", data, err)
	}
}

type failingOutput struct{ err error }

func (w failingOutput) Write([]byte) (int, error) { return 0, w.err }

func TestCheckReportsOutputFailure(t *testing.T) {
	outputErr := errors.New("output closed")
	config := Config{
		Users:        map[string]User{"alice": {BackendUser: "backend"}},
		commandRules: exampleCommandRules(t),
	}
	for _, command := range []string{
		"systemd-analyze --user verify", // Locally approved.
		"reboot",                        // Locally denied.
		"uptime",                        // Model path; no reviewers means denial.
	} {
		t.Run(command, func(t *testing.T) {
			err := checkCommand(context.Background(), config, "alice", command, failingOutput{outputErr})
			if !errors.Is(err, outputErr) {
				t.Fatalf("output failure was lost: %v", err)
			}
		})
	}
}
