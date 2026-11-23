package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	configexamples "ssh-ballpit/examples"
)

var examples = configexamples.Files

const usage = "usage: ssh-ballpit configure|host-key|login-openai|check|providers-check|providers-test|serve [flags]"

type cliOptions struct {
	configPath     string
	authorizedKeys string
	command        string
	user           string
	providerName   string
	codexHome      string
	codexBinary    string
	hostKeyOutput  string
	whitelistFiles []string
	blacklistFiles []string
	suggestionsDir string
}

// checkOutput is shared by local-policy and model-review CLI responses.
// An empty Source means the configured reviewers made the decision.
type checkOutput struct {
	Safe      bool     `json:"safe"`
	Source    string   `json:"source,omitempty"`
	Reviewers []Result `json:"reviewers"`
}

type providerCheckResult struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type providersCheckOutput struct {
	OK        bool                  `json:"ok"`
	Providers []providerCheckResult `json:"providers"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Arguments and output are explicit so CLI tests do not replace process globals.
func run(args []string, output io.Writer) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	action := args[0]
	switch action {
	case "configure", "host-key", "login-openai", "check", "providers-check", "providers-test", "serve":
	default:
		return fmt.Errorf("unknown subcommand %q", action)
	}

	options, err := parseCLIOptions(action, args[1:])
	if err != nil {
		return err
	}
	// These commands do not require an existing server configuration.
	switch action {
	case "configure":
		return writeExampleConfig(options.configPath)
	case "host-key":
		return generateHostKey(options.hostKeyOutput)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if action == "login-openai" {
		return loginFromOptions(ctx, options)
	}

	config, err := readConfig(options.configPath)
	if err != nil {
		return err
	}
	if options.authorizedKeys != "" {
		config.AuthorizedKeys = options.authorizedKeys
	}
	if err := config.commandRules.addFiles(options.whitelistFiles, options.blacklistFiles); err != nil {
		return err
	}
	switch action {
	case "serve":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
		if options.suggestionsDir != "" {
			config.suggestions, err = newSuggestions(options.suggestionsDir)
			if err != nil {
				return err
			}
			slog.Info("writing whitelist suggestions", "path", config.suggestions.path)
		}
		return serve(ctx, config)
	case "check":
		return checkCommand(ctx, config, options.user, options.command, output)
	case "providers-check":
		return checkProviders(ctx, config, output)
	case "providers-test":
		// Diagnostics always call every provider, regardless of command rules.
		return reviewCommand(ctx, config, "provider-test", "printf '%s\\n' hello", output)
	}
	return fmt.Errorf("unhandled subcommand %q", action)
}

// checkProviders probes the complete review path for each configured provider.
// A valid denial still proves the provider was reachable and authenticated.
func checkProviders(ctx context.Context, config Config, output io.Writer) error {
	results, _ := review(ctx, config, "provider-check", "uname -a")
	response := providersCheckOutput{
		OK:        len(results) > 0,
		Providers: make([]providerCheckResult, len(results)),
	}
	for i, result := range results {
		response.Providers[i] = providerCheckResult{
			Name:  config.Providers[i].Name,
			Kind:  config.Providers[i].Kind,
			OK:    result.Error == "",
			Error: result.Error,
		}
		if result.Error != "" {
			response.OK = false
		}
	}
	if err := json.NewEncoder(output).Encode(response); err != nil {
		return fmt.Errorf("write provider check result: %w", err)
	}
	if !response.OK {
		return errors.New("provider check failed")
	}
	return nil
}

func parseCLIOptions(action string, args []string) (cliOptions, error) {
	defaultConfig := os.Getenv("SSHBALLPIT_CONFIG")
	if defaultConfig == "" {
		defaultConfig = "config/server.yaml"
	}

	var options cliOptions
	flags := flag.NewFlagSet(action, flag.ContinueOnError)
	flags.StringVar(&options.configPath, "config", defaultConfig, "server configuration file")
	flags.StringVar(&options.authorizedKeys, "authorized-keys", "", "override custom authorized_keys path")
	flags.StringVar(&options.command, "command", "", "command to review (never executes)")
	flags.StringVar(&options.user, "user", "", "frontend username for review")
	flags.StringVar(&options.providerName, "provider", "", "subscription provider name for login (when more than one is configured)")
	flags.StringVar(&options.codexHome, "codex-home", "", "dedicated private Codex authentication directory")
	flags.StringVar(&options.codexBinary, "codex-binary", "codex", "Codex executable")
	flags.StringVar(&options.hostKeyOutput, "out", "host_key", "new host key file")
	if action == "check" || action == "serve" {
		flags.Func("whitelist-file", "additional whitelist YAML path (repeatable)", func(path string) error {
			options.whitelistFiles = append(options.whitelistFiles, path)
			return nil
		})
		flags.Func("blacklist-file", "additional blacklist YAML path (repeatable)", func(path string) error {
			options.blacklistFiles = append(options.blacklistFiles, path)
			return nil
		})
	}
	if action == "serve" {
		flags.StringVar(&options.suggestionsDir, "suggest-whitelist", "", "write a new suggested whitelist YAML in this directory")
	}
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	if flags.NArg() != 0 {
		return options, errors.New("unexpected positional arguments; use --command for command text")
	}
	return options, nil
}

func writeExampleConfig(path string) error {
	if strings.EqualFold(filepath.Ext(path), ".json") {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		return writeExampleFile("config.example.json", path)
	}

	base := filepath.Dir(path)
	if err := os.MkdirAll(filepath.Join(base, "providers"), 0700); err != nil {
		return err
	}
	if err := writeExampleFile("server.yaml", path); err != nil {
		return err
	}
	for _, name := range []string{"providers/ollama.yaml", "whitelist.yaml", "blacklist.yaml"} {
		if err := writeExampleFile(name, filepath.Join(base, name)); err != nil {
			return err
		}
	}
	return nil
}

func writeExampleFile(name, path string) error {
	data, err := examples.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read embedded example %s: %w", name, err)
	}
	return exclusiveWrite(path, data)
}

func loginFromOptions(ctx context.Context, options cliOptions) error {
	if options.codexHome != "" {
		return login(ctx, options.codexHome, options.codexBinary)
	}

	config, err := readConfig(options.configPath)
	if err != nil {
		return err
	}
	var candidates []Provider
	for _, provider := range config.Providers {
		if provider.Kind == "openai-subscription" &&
			(options.providerName == "" || provider.Name == options.providerName) {
			candidates = append(candidates, provider)
		}
	}
	if len(candidates) != 1 {
		return errors.New("configure one subscription provider, or select it with --provider NAME")
	}
	provider := candidates[0]
	if provider.CodexBinary != "" {
		options.codexBinary = provider.CodexBinary
	}
	return login(ctx, provider.CodexHome, options.codexBinary)
}

func checkCommand(ctx context.Context, config Config, user, command string, output io.Writer) error {
	mapping, ok := config.Users[user]
	if !ok {
		return errors.New("--user must name a configured frontend account")
	}

	source, safe, decisionErr := localCommandDecision(config, command)
	if source != "" {
		response := checkOutput{Safe: safe, Source: source, Reviewers: []Result{}}
		return errors.Join(decisionErr, writeCheckOutput(output, response))
	}
	return reviewCommand(ctx, config, mapping.BackendUser, command, output)
}

func reviewCommand(ctx context.Context, config Config, backendUser, command string, output io.Writer) error {
	results, safe := review(ctx, config, backendUser, command)
	if err := writeCheckOutput(output, checkOutput{Safe: safe, Reviewers: results}); err != nil {
		return err
	}
	if !safe {
		return errors.New("review denied")
	}
	return nil
}

func writeCheckOutput(output io.Writer, response checkOutput) error {
	if err := json.NewEncoder(output).Encode(response); err != nil {
		return fmt.Errorf("write review result: %w", err)
	}
	return nil
}
