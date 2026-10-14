# Agent handoff notes

Layout update: application source/tests now live in `cmd/ssh-ballpit/`, example
assets in `examples/`, and local builds in ignored `bin/`. The root-level Go
paths and build command below are historical; use
`go build -o bin/ssh-ballpit ./cmd/ssh-ballpit` for the current layout.

Last updated: 2026-10-14. These are working notes for future maintenance, not a
replacement for the source code or current runtime inspection. Paths, image IDs,
remote access, provider behavior, and tool versions below describe this session
and should be checked again when used. Do not put credentials or private keys here.

## User intent and working preferences

- Build a simple Go SSH gateway that asks every configured LLM to review commands
  before allowing execution on an existing host.
- All inspection and decisions belong inside ssh-ballpit. Leave the existing
  sshd configuration and runtime alone.
- The user's initial idea included ProxyJump. The implemented design terminates
  SSH and accepts individual exec requests; interactive shells and ProxyJump are
  explicitly unsupported. Do not describe it as a transparent SSH relay.
- Keep human deployment simple: server YAML, one YAML per provider, one device-login
  Podman command, and one daemon Podman command. Keep README compact; put detailed
  configuration and operational explanations in docs.
- Ollama gets its endpoint from its provider YAML. Every provider YAML loaded is
  required to approve; there is no majority vote or fallback that ignores a failure.
- OpenAI subscription login uses device authorization: a one-time code entered
  from another device. Do not restore a local-browser callback flow.
- Rootless Podman shares configuration and credentials with the host through a
  bind mount. Do not bake local credentials into an image.
- Preserve unrelated edits and avoid restarting remote workloads unless the
  current task calls for it. These historical notes do not authorize a new push.
- Motivation: privileged human workstations, an isolated third-party-agent machine
  with restricted LAN access and internet access, local Ollama accelerator hosts,
  and target lab computers/Pis/VMs/containers. After access is configured, the user
  wants automated review rather than being required to approve every action.
- The project is explicitly a personal experiment. The README describes it as
  not production-ready and prefers forks over external contributions.

## Repository and current state

- Workspace: repository checkout on the development host.
- Previous remote: private Forgejo repository, branch `main`.
- Latest commit before these notes: `eb126b8`, connection lifecycle documentation.
- At the start of this notes task, README.md had user-owned uncommitted edits
  (19 added / 4 removed lines). Leave those edits intact and out of the notes commit.
- LICENSE contains the standard Unlicense text. README links to it.
- `.gitignore` excludes the local binary, `config/`, several credential/key paths,
  and environment files. Still inspect the index before committing.

Recent changes, oldest to newest:

| Commit | Change |
| --- | --- |
| `29b47a3` | Client lab-key discovery and Prometheus instrumentation |
| `d8b8ab3` | Optional plaintext command logging |
| `9ff4b0f` | Last five approved and rejected commands included in reviews |
| `ea93afc` | Extend accept-all mode to destination host keys |
| `e39cec7` | Document discovering and installing destination host keys |
| `285ef77` | README architecture diagram, Unlicense, security/test hooks |
| `57dec4b` | Verbose provider diagnostics and strict Ollama fence handling |
| `eb126b8` | Connection and command lifecycle documentation |

## Source map

| File | Responsibility |
| --- | --- |
| `main.go` | CLI: configure, host-key, login-openai, check, providers-test, serve |
| `config.go` | YAML/JSON loading, provider files, defaults, validation |
| `server.go` | SSH listener/authentication, static policy, one exec session, backend relay |
| `review.go` | Shared prompt, provider HTTP adapters, verdict parser, parallel review, Gemini retries |
| `subscription.go` | Dedicated Codex home, device auth, isolated app-server review |
| `cache.go` | Approval cache and decision coordination |
| `history.go` | Per-user recent decision history and review payload |
| `diagnostics.go` | Opt-in provider diagnostics, shell quoting, Ollama schema and replay command |
| `startup.go` | Explicitly selected startup log fields; never dump whole provider structs |
| `metrics.go` | Private Prometheus registry and optional HTTP endpoint |
| `docs/configuration.md` | Configuration, troubleshooting, key discovery, logs, history, metrics |
| `docs/connection-flow.md` | Step-by-step lifecycle with Mermaid flowchart |
| `.githooks/README.md` | Dependencies and activation for commit checks |

Tests include actual local SSH client/server integration, mock HTTP providers,
Codex subprocess fixtures, cache/history behavior, configuration, logging,
metrics, and diagnostics. Normal tests do not need live LLM credentials.

## Connection and execution invariants

1. Startup creates metrics/cache/history, loads keys, verifies backend is a literal
   loopback address, and starts SSH plus optional metrics listeners.
2. Capacity is bounded by `max_connections` (default 32). Each admitted connection
   has a 10-second handshake deadline. Authentication allows three tries.
3. Frontend user must map to a backend user. Strict mode requires the client key in
   both the custom authorized_keys file and the user's fingerprint list.
4. Only one SSH session channel is accepted. Only exec is supported; PTY, shell,
   subsystem, forwarding, and environment requests are rejected.
5. Static shell policy runs before cache lookup. It rejects blank/oversized/NUL
   commands, unsupported syntax, substitutions, selected interpreter/privilege/
   alternate-session entry points, and executable names with slashes or expansion.
6. Cache hit skips models. A miss runs all configured reviewers concurrently with
   the same history snapshot and review deadline. All must return a valid approval.
7. Only after approval does the daemon connect to local sshd, verify the host key,
   and authenticate with its own backend private key as the mapped backend user.
8. Exact command text is sent to backend exec. Client stdin is not forwarded;
   backend stdin is EOF. Stdout/stderr are relayed without model review.
9. Policy/review denial returns 126; backend infrastructure errors return 255;
   completed commands return their own status. Connection closes and slot is freed.

Defaults: `review_seconds: 30`, `session_seconds: 300`. Session time includes
review and execution. Disconnect/shutdown cancels work. Command size limit is
16,384 bytes. LLM review is heuristic, not a sandbox; direct sshd access bypasses it.

## Approval cache and recent history

- Cache key hashes length-prefixed frontend username, backend username, and exact
  command bytes. It is not a command-only global cache.
- Default capacity 10,000, LRU eviction; -1 disables; zero selects the default.
- No TTL. Restart clears it. External filesystem/download changes do not invalidate
  approvals, and changing history does not invalidate them either.
- Only unanimous approvals are cached. Errors and denials are not cached.
- History is scoped to the frontend/backend user pair, last five approved and last
  five rejected commands, oldest first, in memory only.
- Sources: reviewers, approval_cache, static_policy, review_error. A genuine model
  rejection is distinguished from a provider failure.
- Decisions are recorded before backend execution; approved does not mean the
  command actually succeeded. Cached repeats also enter history.
- Snapshots copy lists under a mutex. In-flight requests aren't included until
  their decisions finish. Oversized static denials are truncated in history.
- CLI check/providers-test use empty history and fresh reviews, not daemon cache.

## Key discovery and authentication

`accept_all_keys: true` covers both client key authorization and backend host-key
trust. It does not allow unknown/unmapped frontend users, bypass SSH signatures,
or bypass backend user authentication.

- Client unsigned offers: ERROR `ssh key authorization bypass requested`,
  `signature_verified: false`.
- Successful bypassed client authentication: ERROR `ssh untrusted key accepted`,
  fingerprint and full authorized_key, `signature_verified: true`.
- Backend trust bypass: ERROR `ssh untrusted backend host key accepted`, backend,
  user, reason, fingerprint, key type, and known_hosts_entry. Its verified signature
  proves host-key possession during exchange, not user login success or trust.
- Backend unknown, changed, and revoked keys can all be bypassed in this mode.
- Missing/empty trust files are permitted in lab mode. Malformed/unreadable files
  still fail startup; authorized_keys options are unsupported.
- Strict mode is the default. Discovery never silently writes trust files.
- Docs include ssh-keyscan, comparison with local public host-key fingerprints,
  replacing entries using ssh-keygen -R, and restarting to reload known_hosts.
- Host-key mismatch accepted at ERROR level is deliberately loud and does not by
  itself mean the command failed. Look at subsequent execution outcomes.

## Provider behavior and the Ollama incident

Verdicts require offensive (bool), danger_score (integer 0..100), reasons (string
array), destinations (string array). Missing, duplicate, unknown, mistyped fields,
trailing objects, and invalid JSON fail closed. Approval requires offensive=false
and score strictly below that provider's threshold.

The example Ollama-compatible endpoint is
`http://ollama.example.net:13305`, with model
`Qwen2.5-Coder-7B-Instruct-GGUF-qwen2.5-coder-7b-instruct-q4_k_m:latest`.
Do not assume all models from /api/tags are suitable text reviewers; the endpoint
also advertised image models. Do not change the selected model automatically.

On 2026-10-14 the gateway logged Ollama `invalid verdict object` while Gemini and subscription
approved. A correctly quoted `ls` reproduced the issue locally: HTTP 200, done=true,
done_reason=stop, but message.content was a valid safe verdict wrapped in a Markdown
JSON fence. This occurred even with a schema supplied in format.

Fix in 57dec4b:

- Request a structured verdict schema and explicitly prohibit Markdown in the prompt.
- For Ollama only, unwrap a single complete outer JSON/bare Markdown fence, including
  LF or CRLF variants, then run the unchanged strict verdict validator.
- Do not extract JSON from prose or accept multiple objects/fences, incomplete
  wrappers, duplicate fields, or missing fields.
- Reject Ollama responses with done=false or a nonempty done_reason other than stop.
- Distinguish empty verdict content from invalid JSON/object errors.
- Live checks of `ls` and `uname -a` passed after the change. Tests verify a fenced
  danger_score=90 remains a denial, and malformed/fenced ambiguity stays rejected.

The user initially invoked remote uname with backticks. That runs uname on the
client and sends the output as the remote command. Single quotes are required.
However, quoting was not the cause of the separate fenced-response bug.

Gemini HTTP 503 is retried at most twice with approximately 1s/2s delays plus jitter,
within the original review deadline. No automatic model substitution. Other errors
and malformed verdicts are not retried.

## Logging and metrics

- Default daemon logs are JSON to stderr. Client addresses, user, fingerprint, and
  command hash allow correlation. Provider verdict reasons aren't in normal review logs.
- `log_commands: true` logs each well-formed authenticated exec command before policy,
  including cache hits and denials. This intentionally can expose embedded secrets.
- `verbose_review_logs: true` is separate, default false, and emits a startup warning.
  It includes plaintext/history even if log_commands=false.
- Diagnostic event: `provider review diagnostic`; ERROR on failure, INFO on success.
  Includes provider/model/origin/user/hash, stage, elapsed time, HTTP status, byte
  counts, a capped 16 KiB response excerpt, truncation, transport category, and
  ollama_markdown_unwrapped. Gemini shows the final retry response.
- Known API-key values are redacted from excerpts. Credential headers are not logged.
  Arbitrary secrets in commands/model text cannot be reliably redacted.
- Ollama adds ollama_curl with the exact serialized request and endpoint. It shell-quotes
  all data and uses printf piped to curl --data-binary @-, with a 120-second timeout.
  Tests execute a replay against a local mock server, check identical request bytes,
  and verify quotes/substitutions cannot execute payload text in the local shell.
- Subscription diagnostics contain metadata/errors, not raw Codex output.
- Cache hits do not produce provider diagnostics. Turn verbose mode off after triage.
- `metrics_listen` enables plain HTTP /metrics, commonly 127.0.0.1:2112. No auth/TLS;
  bind only where monitoring access is appropriate. Default is disabled.
- Metrics cover connections/auth, client/backend trust bypasses, protocol rejection,
  review outcomes/latency, cache, execution outcomes/latency, plus Go/process metrics.
  Labels intentionally exclude usernames, fingerprints, commands, and arbitrary errors.

## OpenAI subscription isolation

- Dedicated private codex_home, mode 0700. Do not reuse the operator's general Codex home.
- Uses `codex login --device-auth` and file credential storage. Container bind mount
  preserves its login state. Runtime pins Codex CLI 0.162.1 as of this session.
- Reviews run via an isolated, ephemeral app-server session with tools/capabilities
  disabled, read-only sandbox, no sandbox network, and no inherited instructions.
- A process-wide review slot serializes subscription work to avoid concurrent
  credential rotation. Queue time counts against review timeout.
- An earlier failure rejected any skills directory. Codex itself creates skills/.system;
  current code permits that generated directory while rejecting custom skills,
  symlinks, nonregular files, configuration, plugins, hooks, and AGENTS.md in the home.
- Do not weaken these isolation checks to repair login issues. Consult current source
  and official product documentation before changing app-server protocol/version.

## Development and container workflow

Host Go was not on PATH. This installation worked:

```sh
/tmp/ssh-ballpit-toolchain/go/bin/go test ./...
/tmp/ssh-ballpit-toolchain/go/bin/go vet ./...
/tmp/ssh-ballpit-toolchain/go/bin/go build -o ssh-ballpit .
```

The temporary toolchain was Go 1.27.2; go.mod and the build image use Go 1.26.
The host lacked GCC for race builds earlier. Race tests can use the Go container
if needed. Regular tests/vet run both in pre-commit and in the Containerfile.

```sh
podman build -t localhost/ssh-ballpit .
podman run --rm --name ssh-ballpit --userns=keep-id --network=host \
  -v "$PWD/config:/config:z" localhost/ssh-ballpit
```

Image defaults: serve, config /config/server.yaml, workdir /workspace. No image
USER directive: --userns=keep-id chooses the host UID. Host networking allows the
loopback backend. Rootless Podman on Linux is the documented environment.

Device login:

```sh
podman run --rm -it --userns=keep-id \
  -v "$PWD/config:/config:z" localhost/ssh-ballpit login-openai
```

Builds can spend minutes silently copying/saving layers. Poll the actual build
session and provide periodic updates. Avoid launching a build before finalizing
source/tests: one prior build captured a subsequently corrected assertion and
had to be rerun. Do not restart user containers just because an image was rebuilt.
On another machine, pull, rebuild, and recreate the container; restarting an old
container alone does not switch it to the new image.

Latest local successful image build for 57dec4b:
`49b80870348a75b2211791ffcb091b6bc731f754028f1d71b3c8d3de58ccf72a`.
Later changes before these notes were documentation-only.

## Hooks, scanning, and publishing

Versioned .githooks/pre-commit runs these security checks:

1. `gitleaks git --pre-commit --staged --redact --no-banner`.
2. Export the staged index to a private temporary directory.
3. `grype dir:<snapshot> --fail-on high`.
4. Run Go tests without cached test results and go vet on that same snapshot,
   with GOWORK=off and -mod=readonly.
5. Remove the snapshot on exit. Missing tools/errors block the commit.

Enable per clone with `git config --local core.hooksPath .githooks`; already enabled
in this checkout. Here, use:

```sh
GO=/tmp/ssh-ballpit-toolchain/go/bin/go git commit -m "Describe the change"
```

Scanners are installed under ~/.local/bin. Source scans do not cover the built
container's OS packages or installed Codex CLI. Latest scans passed the High/Critical
threshold but reported GO-2026-5932 for golang.org/x/crypto v0.58.0 with Unknown
severity. This advisory was not investigated/resolved in this work; do not report
that the scan found no vulnerabilities or add an ignore to conceal it.

For private Forgejo access use the environment's configured tooling.
Never copy its credentials into repository notes, shell output, or Git remotes.
Authenticated pushes in this session used temporary repository-scoped tokens,
a private temporary askpass helper, and token revocation in a finally block.
Do not assume an unauthenticated clone of another private repository will work.

Mermaid diagrams were parsed successfully using temporary mermaid/jsdom npm
packages under /tmp/ssh-ballpit-mermaid-check. This is not a project dependency.
A later session may need to recreate the temporary validation environment.

## Verified remote access and remaining boundaries

The user explicitly authorized trying the gateway in that session. This succeeded
from the development host:

```sh
ssh -i ~/.ssh/id_ed25519 -p 2222 -o BatchMode=yes -o ConnectTimeout=10 \
  system_user@gateway.example.net 'uptime'
```

Observed on 2026-10-14: exit 0, output `21:13:02 up 6:02, 8 users, load average:
3.24, 3.13, 3.12`. The subsequent connection-closed message was normal completion.
This establishes one successful end-to-end command, not unrestricted remote access
or proof that every future request/model call will work. It could have been a
cache hit; no accompanying server logs were inspected for that successful request.

Do not infer permission to inspect remote secrets, broaden firewall access,
modify sshd, or restart gateway services from this successful read-only test.
No remaining code fix was identified at the end of the lifecycle-docs task.
The user-owned README edits remain pending separately from these notes.
