# Tests and checks

The Go tests live in [`cmd/ssh-ballpit/`](../cmd/ssh-ballpit/). There are 45
top-level tests across 12 files, with additional table-driven cases. The
[`examples/`](../examples/) package has no test file, but the tests load its
embedded configuration and rule examples.

## Run the checks

From the repository root, use Go 1.26 or newer:

```sh
GOWORK=off go test -mod=readonly -count=1 ./...
GOWORK=off go vet -mod=readonly ./...
```

`-count=1` runs the tests instead of accepting a cached result. `-mod=readonly`
rejects changes to `go.mod` or `go.sum`; `GOWORK=off` keeps an enclosing Go
workspace from changing the module under test. To focus on one behavior, run
`go test -mod=readonly -count=1 ./cmd/ssh-ballpit -run '^TestCommandRulesCLI$'`
(replace the test name as needed).

When changing shared state such as the approval cache, history, reviews, or
metrics, also run:

```sh
GOWORK=off go test -race -mod=readonly -count=1 ./...
```

The race detector needs a working C compiler. It is an additional check and is
not part of the pre-commit hook or GitHub Actions workflow.

## What the tests protect

| Test files | Behavior and reason |
| --- | --- |
| `config_test.go`, `main_test.go` | YAML and legacy JSON loading, provider defaults, malformed settings, generated examples, CLI argument errors, non-overwriting `configure`, private host-key permissions, and error propagation. Invalid configuration must stop startup without exposing secret values. |
| `review_test.go`, `adapters_test.go` | Static command policy, strict verdict JSON, unanimous approval, denial on errors or cancellation, HTTP request and response shapes for OpenAI, Anthropic, DeepSeek, and Gemini, history payloads, and Gemini retry limits. A missing, malformed, or failed review must not approve a command. |
| `command_rules_test.go` | Shipped rule examples, strict rule loading, full-command regex matches, blacklist precedence, and a broad-whitelist bypass attempt. Session and CLI tests check that static policy comes first, blacklist overrides cached approvals, whitelist applies only to a simple command, and `providers-test` still contacts the provider. |
| `cache_test.go`, `history_test.go` | Cache hits, key separation by user/backend/command bytes, eviction, no caching of denials or partial approvals, fresh state after restart, bounded per-user history, review payload snapshots, and concurrent access. These tests guard against stale authorization and cross-user disclosure. |
| `server_test.go`, `logging_test.go` | Real local SSH connections to the gateway and a test backend SSH server; client and backend identity separation; key and host-key checks; denied shells and forwarding; command output, stdin EOF, exit status, logs, and metrics. Logging tests also check bounded error reasons and that credentials or raw commands stay out of default logs. |
| `diagnostics_test.go`, `subscription_test.go` | Opt-in diagnostics, response redaction, bounded excerpts, safe Ollama replay commands, strict Ollama fenced-verdict handling, and an isolated fake Codex subprocess. Subscription tests reject tool use, approval requests, invalid responses, inherited API keys, unsafe skill paths, and timeout failures. |
| `metrics_test.go` | Metrics listener startup, occupied-port errors, shutdown, endpoint routing, and independent registries. |

The SSH integration test creates temporary Ed25519 keys, starts both SSH servers
on loopback addresses, and uses a local mock HTTP reviewer. The backend fixture
returns fixed output; it does not run submitted shell commands or connect to the
host's existing sshd. HTTP adapter tests use local mock servers. Codex tests
re-execute the Go test binary as a fake `codex` program. No live provider account,
API key, or SSH credential is needed.

The optional replay assertion in `TestOllamaDiagnosticsAndReplay` needs `curl`;
that assertion is skipped when `curl` is unavailable. The other Go tests still
run. Tests create files under temporary directories and use local TCP ports, so
the runner must allow loopback listeners.

## Commit and CI gates

The [pre-commit hook](pre-commit.md) scans the staged diff with Gitleaks, scans
a temporary checkout of the staged tree with Grype (failing on High or Critical
findings), then runs the Go test and vet commands above on that checkout. The
[GitHub Actions workflow](../.github/workflows/ci.yml) runs equivalent checks
on every push. After they pass, an image-input change pushed to `main` builds
and publishes the container. The Containerfile also runs Go tests and vet while
building the image.

These checks do not validate live LLM behavior, a deployed sshd, or the runtime
container's OS packages. Grype's source scan covers dependency findings in the
checkout; the container build is a separate check. Add focused behavioral tests
for authorization changes, precedence, malformed configuration, and bypass
attempts, then run the full suite before committing.
