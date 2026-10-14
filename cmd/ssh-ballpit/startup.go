package main

import (
	"log/slog"
	"net/url"
	"sort"
)

// Explicitly select log fields instead of serializing configuration structs:
// providers can contain inline credentials and may gain new secret fields.
func logStartupConfig(config Config) {
	slog.Info("startup command rules configuration", "whitelist_patterns", len(config.commandRules.whitelist),
		"blacklist_patterns", len(config.commandRules.blacklist), "precedence", "static_policy,blacklist,whitelist,approval_cache,reviewers")
	slog.Info("startup configuration",
		"listen", config.Listen, "backend", config.Backend,
		"host_key", config.HostKey, "authorized_keys", config.AuthorizedKeys,
		"backend_key", config.BackendKey, "known_hosts", config.KnownHosts,
		"providers_dir", config.ProvidersDir, "provider_count", len(config.Providers),
		"review_seconds", config.ReviewSeconds, "session_seconds", config.SessionSeconds,
		"accept_all_keys", config.AcceptAllKeys, "metrics_listen", config.MetricsListen, "log_commands", config.LogCommands, "verbose_review_logs", config.VerboseReviewLogs,
		"max_connections", config.MaxConnections, "require_all_reviewers", true)
	slog.Info("startup approval cache configuration", "max_entries", config.ApprovalCacheEntries,
		"enabled", config.ApprovalCacheEntries >= 0, "storage", "memory", "cleared_on_restart", true)
	slog.Info("startup review history configuration", "approved_limit", recentHistoryLimit, "rejected_limit", recentHistoryLimit, "scope", "frontend_backend_user", "storage", "memory")
	if config.VerboseReviewLogs {
		slog.Warn("VERBOSE REVIEW LOGGING ENABLED", "warning", "provider response excerpts and Ollama replay commands include plaintext commands and recent history, even when log_commands is false")
	}
	if config.AcceptAllKeys {
		slog.Error("ACCEPT ALL SSH KEYS ENABLED", "accept_all_keys", true, "warning", "configured users may authenticate with any private key and backend host keys are accepted; client and backend key trust failures are logged and bypassed")
	}
	names := make([]string, 0, len(config.Users))
	for name := range config.Users {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		u := config.Users[name]
		slog.Info("startup user configuration", "user", name, "backend_user", u.BackendUser,
			"fingerprints", u.Fingerprints)
	}
	for _, p := range config.Providers {
		source := "none"
		switch {
		case p.Kind == "openai-subscription":
			source = "codex_home"
		case p.APIKey != "":
			source = "inline"
		case p.KeyFile != "":
			source = "file"
		case p.KeyEnv != "":
			source = "environment"
		}
		slog.Info("startup provider configuration", "name", p.Name, "kind", p.Kind,
			"model", p.Model, "threshold", p.Threshold,
			"endpoint_origin", endpointOrigin(p.Endpoint), "credential_source", source,
			"key_file", p.KeyFile, "key_env", p.KeyEnv,
			"codex_home", p.CodexHome, "codex_binary", p.CodexBinary)
	}
}

// URL paths, queries, fragments and userinfo can contain credentials.
func endpointOrigin(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return "[redacted invalid endpoint]"
	}
	return u.Scheme + "://" + u.Host
}
