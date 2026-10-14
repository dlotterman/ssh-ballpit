package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Provider describes one required reviewer and its credential source.
// Never log the whole struct: APIKey can contain an inline secret.
type Provider struct {
	Name      string `json:"name" yaml:"name"`
	Kind      string `json:"kind" yaml:"kind"`
	Model     string `json:"model" yaml:"model"`
	Endpoint  string `json:"endpoint" yaml:"endpoint"`
	Threshold int    `json:"threshold" yaml:"threshold"`

	APIKey  string `json:"-" yaml:"api_key,omitempty"`
	KeyFile string `json:"-" yaml:"key_file,omitempty"`
	KeyEnv  string `json:"key_env,omitempty" yaml:"key_env,omitempty"`

	CodexHome   string `json:"codex_home,omitempty" yaml:"codex_home,omitempty"`
	CodexBinary string `json:"codex_binary,omitempty" yaml:"codex_binary,omitempty"`
	// Retained only to explain how to migrate legacy subscription configs.
	CredentialsFile string `json:"credentials_file,omitempty" yaml:"credentials_file,omitempty"`
}

// User maps an authenticated frontend account to an existing backend account.
type User struct {
	BackendUser  string   `json:"backend_user" yaml:"backend_user"`
	Fingerprints []string `json:"fingerprints" yaml:"fingerprints"`
}

type Config struct {
	// SSH listeners, keys and account mappings.
	Listen         string          `json:"listen" yaml:"listen"`
	HostKey        string          `json:"host_key" yaml:"host_key"`
	AuthorizedKeys string          `json:"authorized_keys" yaml:"authorized_keys"`
	Backend        string          `json:"backend" yaml:"backend"`
	BackendKey     string          `json:"backend_key" yaml:"backend_key"`
	KnownHosts     string          `json:"known_hosts" yaml:"known_hosts"`
	Users          map[string]User `json:"users" yaml:"users"`
	AcceptAllKeys  bool            `json:"accept_all_keys" yaml:"accept_all_keys"`

	// Review settings and resource limits.
	ProvidersDir         string     `json:"providers_dir,omitempty" yaml:"providers_dir,omitempty"`
	Providers            []Provider `json:"providers" yaml:"-"`
	ReviewSeconds        int        `json:"review_seconds" yaml:"review_seconds"`
	SessionSeconds       int        `json:"session_seconds" yaml:"session_seconds"`
	MaxConnections       int        `json:"max_connections" yaml:"max_connections"`
	ApprovalCacheEntries int        `json:"approval_cache_entries" yaml:"approval_cache_entries"`

	// Optional observability; raw commands are private by default.
	VerboseReviewLogs bool   `json:"verbose_review_logs" yaml:"verbose_review_logs"`
	LogCommands       bool   `json:"log_commands" yaml:"log_commands"`
	MetricsListen     string `json:"metrics_listen" yaml:"metrics_listen"`

	// Runtime state is initialized after decoding, never read from config files.
	metrics      *daemonMetrics
	commandRules commandRules
}

func readConfig(path string) (Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	// Legacy JSON keeps its original path semantics and inline provider list.
	if strings.EqualFold(filepath.Ext(path), ".json") {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&config); err != nil {
			return config, err
		}
		var trailing any
		if decoder.Decode(&trailing) != io.EOF {
			return config, fmt.Errorf("trailing configuration data")
		}
	} else {
		// Decoder errors can include inline credentials; report only the file.
		if err = decodeYAML(data, &config); err != nil {
			return config, fmt.Errorf("invalid YAML configuration: %s", path)
		}
		base := filepath.Dir(path)
		if config.Listen == "" {
			config.Listen = "127.0.0.1:2222"
		}
		if config.Backend == "" {
			config.Backend = "127.0.0.1:22"
		}
		config.HostKey = relativePath(base, config.HostKey, "state/host_key")
		config.AuthorizedKeys = relativePath(base, config.AuthorizedKeys, "authorized_keys")
		config.BackendKey = relativePath(base, config.BackendKey, "backend_key")
		config.KnownHosts = relativePath(base, config.KnownHosts, "known_hosts")
		config.ProvidersDir = relativePath(base, config.ProvidersDir, "providers")
		config.Providers, err = readProviders(config.ProvidersDir, filepath.Join(base, "state", "codex"))
		if err != nil {
			return config, err
		}
	}
	config.commandRules, err = readCommandRules(filepath.Dir(path))
	if err != nil {
		return config, err
	}
	return validateConfig(config)
}

func validateConfig(config Config) (Config, error) {
	if config.ApprovalCacheEntries < -1 {
		return config, fmt.Errorf("approval_cache_entries must be -1 (disabled), 0 (default), or positive")
	}
	if config.ApprovalCacheEntries == 0 {
		config.ApprovalCacheEntries = defaultApprovalCacheEntries
	}

	if len(config.Providers) == 0 || len(config.Providers) > 16 {
		return config, fmt.Errorf("between one and sixteen reviewers are required")
	}
	seen := map[string]bool{}
	for _, provider := range config.Providers {
		if provider.Name == "" || seen[provider.Name] || provider.Model == "" || provider.Threshold < 1 || provider.Threshold > 100 {
			return config, fmt.Errorf("invalid reviewer configuration")
		}
		seen[provider.Name] = true
		switch provider.Kind {
		case "ollama", "openai", "anthropic", "deepseek", "gemini", "openai-subscription":
		default:
			return config, fmt.Errorf("unsupported provider %q", provider.Kind)
		}
		if provider.Kind == "openai-subscription" {
			if provider.CodexHome == "" || provider.CredentialsFile != "" ||
				provider.Endpoint != "" || provider.KeyEnv != "" ||
				provider.APIKey != "" || provider.KeyFile != "" {
				return config, fmt.Errorf("subscription now requires codex_home; remove credentials_file, endpoint and key_env, then run login-openai --codex-home PATH")
			}
			continue
		}

		endpoint, err := url.Parse(provider.Endpoint)
		if err != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
			return config, fmt.Errorf("invalid endpoint for %s (cloud endpoints require HTTPS)", provider.Name)
		}
		allowOllamaHTTP := provider.Kind == "ollama" && endpoint.Scheme == "http"
		if endpoint.Scheme != "https" && !allowOllamaHTTP {
			return config, fmt.Errorf("invalid endpoint for %s (cloud endpoints require HTTPS)", provider.Name)
		}
	}
	if config.ReviewSeconds <= 0 {
		config.ReviewSeconds = 30
	}
	if config.SessionSeconds <= 0 {
		config.SessionSeconds = 300
	}
	if config.MaxConnections <= 0 {
		config.MaxConnections = 32
	}
	return config, nil
}

func relativePath(base, value, fallback string) string {
	if value == "" {
		value = fallback
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(base, value)
}

func decodeYAML(data []byte, out any) error {
	if len(data) > 1<<20 {
		return fmt.Errorf("YAML file too large")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return fmt.Errorf("expected one YAML document")
	}
	return nil
}

func readProviders(dir, codexHome string) ([]Provider, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var providers []Provider
	defaults := map[string]string{
		"openai":    "https://api.openai.com/v1",
		"anthropic": "https://api.anthropic.com/v1",
		"deepseek":  "https://api.deepseek.com",
		"gemini":    "https://generativelanguage.googleapis.com/v1beta",
	}
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yaml" && ext != ".yml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var provider Provider
		if decodeYAML(data, &provider) != nil {
			return nil, fmt.Errorf("invalid provider YAML: %s", path)
		}
		if provider.Name == "" {
			provider.Name = strings.TrimSuffix(entry.Name(), ext)
		}
		if provider.Endpoint == "" {
			provider.Endpoint = defaults[provider.Kind]
		}
		if provider.Kind == "openai-subscription" {
			if provider.CodexHome == "" {
				provider.CodexHome = codexHome
			} else {
				provider.CodexHome = relativePath(dir, provider.CodexHome, "")
			}
		}
		credentialSources := 0
		for _, value := range []string{provider.APIKey, provider.KeyEnv, provider.KeyFile} {
			if value != "" {
				credentialSources++
			}
		}
		if credentialSources > 1 {
			return nil, fmt.Errorf("provider %s: choose only one of api_key, key_file or key_env", provider.Name)
		}
		if provider.KeyFile != "" {
			provider.KeyFile = relativePath(dir, provider.KeyFile, "")
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func providerKey(provider Provider) (string, error) {
	if provider.KeyFile != "" {
		data, err := os.ReadFile(provider.KeyFile)
		if err != nil {
			return "", fmt.Errorf("cannot read provider credential file")
		}
		return strings.TrimSpace(string(data)), nil
	}
	if provider.APIKey != "" {
		return provider.APIKey, nil
	}
	return os.Getenv(provider.KeyEnv), nil
}
