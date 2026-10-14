# Working on ssh-ballpit

This is a Go SSH command gateway. It accepts one exec request per connection,
reviews the command, and forwards approved text unchanged to a separate local
sshd. It is an experiment, not a sandbox or transparent SSH relay.

## Code and documentation

Application Go source and tests live in `cmd/ssh-ballpit/`; the Go filenames below
are relative to that directory. Keep the repository root free of application source.

- `main.go` and `config.go`: CLI, strict YAML/legacy JSON configuration, examples.
- `server.go`: listener lifecycle, SSH sessions and command execution flow.
- `auth.go`, `backend.go`, `keys.go`: client authentication, backend relay and keys.
- `policy.go`: non-overridable static shell checks.
- `connection_errors.go`: bounded connection error categories for logs/metrics.
- `command_rules.go`: optional whitelist/blacklist YAML and regex matching.
- `cache.go`, `history.go`: approval cache and per-user decision history.
- `review.go`: concurrent reviews and unanimous approval.
- `provider_http.go`, `subscription.go`: HTTP adapters and isolated Codex reviews.
- `verdict.go`: strict verdict parsing and provider result types.
- `examples/` (repository root): starter configs, legacy JSON example and the
  small Go package embedding them. Keep the binary independent of checkout paths.
- `sbom.md` (repository root): dependency versions, license evidence and scope.
- `scripts/generate-sbom.py` (repository root): regenerate the Markdown inventory
  from the selected Go graph and an OCI image using Syft; keep license scope clear.
- `docs/configuration.md` and `docs/connection-flow.md`: behavior and lifecycle.
- `docs/agent/notes.md`: historical context; verify environment details before use.

## Readability

Prefer descriptive names, explicit control flow and small functions with one job.
Keep request payloads readable across multiple lines. Use comments to explain
security constraints and protocol quirks, rather than narrating obvious Go code.
Separate standard-library imports from external and local package imports.

## Container tooling

References to `podman` mean container tooling generally, not a requirement to use
Podman. Docker (`docker`) or another compatible container engine should work.
Adapt engine-specific flags, build-file names and volume options as needed while
preserving the intended networking, file permissions and user mappings.

## Invariants

- Keep static policy before command rules; blacklist wins over whitelist and cache.
  Whitelist bypass is restricted to one simple command. Unmatched commands require
  every configured reviewer to approve; failures and invalid responses deny.
- Keep SSH and CLI `check` decisions consistent. `providers-test` must always call
  providers, even when command rules would match its diagnostic command.
- Preserve client/backend identity separation, backend host-key verification,
  loopback-only backend addresses, EOF on backend stdin, and session deadlines.
- Do not change the host's existing sshd configuration as part of gateway work.
- Compile and validate rules at startup; do not silently ignore invalid files.
  Configuration changes require restart and a fresh in-memory approval cache.
- Never log credentials, raw commands, rule patterns or model responses by default.
  Keep opt-in command/diagnostic logging explicit and metric labels bounded.
- Keep `config/`, keys, credentials and local artifacts out of version control.
  Preserve unrelated workspace edits.

## Validation

Use the Go version required by `go.mod` or newer. Run `gofmt` on changed Go files,
`go test -mod=readonly ./...`, and `go vet -mod=readonly ./...`. Tests use local
SSH servers and mock providers; live credentials should not be needed. Add
behavioral tests for authorization changes, precedence, malformed configuration,
and bypass attempts. Run race tests when changing shared concurrent state.

Build from the repository root with `go build -o bin/ssh-ballpit ./cmd/ssh-ballpit`.
Update `Containerfile` and `.containerignore` when changing source layout.

Update examples and configuration/lifecycle documentation with behavior changes.
Keep the Mermaid diagrams in README and connection-flow documentation consistent
with the code. See `.githooks/README.md` for the staged secret/vulnerability checks;
do not disable hooks to make a commit pass.
