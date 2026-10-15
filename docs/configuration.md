# Configuration reference

For the request lifecycle, see [Connection and command flow](connection-flow.md).

## Directory layout

```text
config/
  server.yaml
  whitelist.yaml          # optional command regexes; skip LLM review on match
  blacklist.yaml          # optional command regexes; deny on match
  authorized_keys
  backend_key
  known_hosts
  providers/
    ollama.yaml
    openai-subscription.yaml
  state/                  # created automatically; preserve across restarts
    host_key
    codex/
```

The README's single writable bind mount maps this directory to `/config`. Paths
in `server.yaml` resolve relative to that file. Paths in provider YAML resolve
relative to the provider file. The default Codex home is `state/codex` beside
`server.yaml`; login and review use the same directory. Host usage defaults to
`config/server.yaml`, and the image defaults to `/config/server.yaml`.

Do not commit `config/`. Keep it private (directory mode 0700), with private keys
and provider YAML containing API keys mode 0600. `:z` applies a shared SELinux
container label to this dedicated directory. `--userns=keep-id` keeps generated
files owned by your host user. Do not mount/relabel your entire home or ~/.ssh.

## Server YAML

Optional fields and defaults:

| Field | Default |
| --- | --- |
| `listen` | `127.0.0.1:2222` |
| `accept_all_keys` | `false` |
| `log_commands` | `false` |
| `verbose_review_logs` | `false` |
| `metrics_listen` | empty (disabled) |
| `backend` | `127.0.0.1:22` |
| `authorized_keys` | `authorized_keys` |
| `backend_key` | `backend_key` |
| `known_hosts` | `known_hosts` |
| `host_key` | `state/host_key` (generated if absent) |
| `providers_dir` | `providers` |
| `review_seconds` | `30` |
| `session_seconds` | `300` |
| `max_connections` | `32` |
| `approval_cache_entries` | `10000` (`-1` disables; `0` uses default) |

`users` maps frontend names to `backend_user` and a list of allowed client key
`fingerprints`. Use `ssh-keygen -lf client.pub -E sha256` to get fingerprints. A
client key must match both the custom authorized_keys file and that user's
fingerprint list. Only bare public keys are supported; authorized_keys options
are rejected. Backend private keys must be unencrypted and already authorized
by existing sshd. Backend authentication is separate from client authentication.
Verify the backend host-key fingerprint independently before adding known_hosts;
nonstandard ports use `[127.0.0.1]:PORT` entries. The backend must be a literal
loopback address. Host networking lets the container reach the host's loopback;
there is no `-p` mapping. With Podman Machine, loopback refers to the VM.

## Command whitelist and blacklist

Optional command rules live beside the server configuration. See the
[whitelist and blacklist guide](command-rules.md) for setup, regex matching,
decision order, and CLI checks.

## Provider YAML

Each `.yaml` or `.yml` file is one required reviewer. Files load in filename order;
non-YAML files are ignored. Empty directories, unknown fields, duplicate YAML
keys, duplicate reviewer names and multiple YAML documents are errors. All
configured reviewers must approve; no provider is silently skipped.

Required fields: `kind`, `model`, `threshold` (integer 1–100). Optional `name`
defaults to the filename without extension. Model names must be available to
your account. Scores must be strictly below the threshold, and `offensive` must
be false. Provider failures, timeouts and invalid verdicts deny execution.

| Kind | Endpoint default |
| --- | --- |
| `ollama` | **Required in YAML** |
| `openai` | `https://api.openai.com/v1` |
| `anthropic` | `https://api.anthropic.com/v1` |
| `deepseek` | `https://api.deepseek.com` |
| `gemini` | `https://generativelanguage.googleapis.com/v1beta` |
| `openai-subscription` | Local Codex app-server; no endpoint or API key |

API providers accept one of `api_key`, `key_file` (plain key text, relative to
provider YAML), or `key_env` (environment-variable name). For example:

```yaml
kind: gemini
model: gemini-3.1-flash-lite
key_file: ../gemini-key
threshold: 30
```

Use `--env-file /host/path/providers.env` with Podman if choosing `key_env`.
The file must contain plain unquoted `KEY=value` lines, not shell commands.
Commands that reach model review are sent to every configured provider; avoid
secrets in command text. Later reviews can also include locally decided commands
in recent history.
Audit logs record command hashes, usernames and reviewer outcomes, not raw
commands or model explanations. `check` prints explanations for the operator.

## OpenAI subscription

The image includes Codex 0.162.1. `login-openai` uses the subscription YAML's
`codex_home` and `codex_binary`, defaulting to the shared state directory and
`codex` on PATH. If multiple subscription reviewers are configured, select one
with `login-openai --provider NAME`. A dedicated custom Codex home must be mode
0700 and contain no personal configuration, instructions, custom skills, hooks or plugins.
Codex automatically creates a `skills/.system` cache; this is allowed and does not
need to be deleted. Reviews select no environments or capability roots and reject
inherited instruction sources. Symlinks in the skills cache are rejected.
If an older build reports `dedicated Codex home must not contain skills` after
login or a first review, update and rebuild the image, then recreate the container.
Keep the existing credential directory; another device login is not required.

Device login may need enabling in ChatGPT security settings or workspace
permissions. [OpenAI device login documentation](https://learn.chatgpt.com/docs/auth#login-on-headless-devices).
Codex manages token refresh. Stop the daemon before logging in again; run one
daemon per Codex home. Review subprocesses are serialized so credential refresh
has one owner; queue time counts toward the review deadline. Increase
`review_seconds` for slower models or multiple subscription reviewers.

Review sessions are ephemeral, with environment access, apps, plugins, shell
tools, hooks and multi-agent features disabled. Returned session restrictions
are checked. Unexpected tools, approval requests or failed turns deny execution.
Live subscription inference requires completing device authorization first.

## Other commands

Append a subcommand to the same Podman run command used in the README:

```text
providers-test
check --user alice --command 'uname -a'
serve --authorized-keys /config/other_authorized_keys
```

`providers-test` and `check` never execute submitted commands. To run detached,
add `-d`; use `podman logs ssh-ballpit` and `podman stop ssh-ballpit` afterward.
The native binary accepts the same commands with `--config PATH`. `configure`
creates starter server, Ollama, whitelist and blacklist YAML files without
overwriting existing files.
Legacy JSON configuration with inline providers remains readable for migration;
new deployments should use YAML. An old direct-OAuth `credentials_file` must be
replaced by device login; those tokens are not imported.

## Execution limits

Only one exec session per SSH connection is accepted. Shells, PTYs, SFTP,
forwarding and client environment requests are rejected. Backend stdin receives
EOF; unreviewed client input is never forwarded. Shell substitutions, dynamic
executable names and common interpreter entry points are rejected before model
review. Approved text passes unchanged to sshd. Stdout, stderr and exit status
return to the client; denial uses 126 and backend failures use 255.

Disconnects and deadlines close the backend connection but cannot guarantee that
already-started or detached processes terminate. Models only see command text
and the destination username; mutable scripts, PATH, startup files, program shell
escapes and downloaded content can hide behavior. This does not fetch downloads,
resolve redirects, inspect repositories or verify signatures. OS permissions
remain the execution boundary.

## Development

See [CI/CD and builds](ci-cd.md) for local builds and the GitHub publish flow,
and [tests and checks](tests.md) for test coverage and fixtures.

## Gemini availability

On 2026-10-14, live tests with this project's key found that `gemini-3.8-flash`
and `gemini-3.6-flash` returned HTTP 503 (high demand). The newer Interactions
API also returned 503 for 3.8 Flash. `gemini-3.1-flash-lite` returned valid review
verdicts, with occasional 503s, and is now the example model. Availability may
vary by account and time; this is not a comprehensive evaluation of model safety.

For an existing deployment, change only `model` in your Gemini provider YAML to
`gemini-3.1-flash-lite`, then recreate the container. Keep your existing key and
threshold. Test the configured reviewers without executing a command:

```sh
podman run --rm --userns=keep-id -v "$PWD/config:/config:z" \
  docker.io/dlotterman/ssh-ballpit:latest providers-test
```

Gemini HTTP 503 responses are retried at most twice, after roughly one and two
seconds with jitter, within `review_seconds`. Persistent failures still deny the
command, and the daemon never substitutes a different model automatically.
Other HTTP errors and invalid verdicts are not retried.
Google documents 503 as temporary overload or unavailability:
https://ai.google.dev/gemini-api/docs/troubleshooting


## Connection and authentication logs

The daemon writes JSON logs to stderr. In a container, use:

```sh
podman logs -f ssh-ballpit
```

Authentication logs include `remote_addr`, `user`, `method`, and a `reason` on
failure. Rejected public-key offers also include the SHA256 `fingerprint`.
An accepted key offer is not logged as a successful login until SSH verifies
actual authentication. The normal initial `none` method-discovery probe is omitted.

| Reason | What to check |
| --- | --- |
| `unknown_user` | Frontend username must exist in `server.yaml` → `users`. |
| `missing_backend_user` | The user's `backend_user` must be set. |
| `key_not_in_authorized_keys` | Add the intended client public key to the configured authorized_keys file. |
| `key_not_allowed_for_user` | Add that key's fingerprint to the intended user's mapping. |
| `unsupported_authentication_method` | Use public-key authentication; passwords are not supported. |
| `connection_limit_reached` | Existing connections have filled `max_connections`. |
| `timeout` / `peer_disconnected` | The peer stalled or closed the connection during the handshake. |
| `no_common_ssh_algorithm` | Client and server could not negotiate compatible SSH algorithms. |
| `connection_refused` | Check `backend` and whether local sshd is listening. |
| `host_key_unknown` / `host_key_mismatch` | Verify sshd's host key and the configured known_hosts entry. |
| `backend_authentication_rejected` | Check the backend username and whether sshd accepts the backend key. |

Some reasons include explanatory text after the code. Failed handshakes also
include local/remote addresses and the error type. Backend failures identify the
backend address, backend username, command hash and failure stage. Rejected
channels and session requests identify their type. Raw password values, private
keys, and peer-supplied disconnect/error messages are not logged. Command text
is omitted by default; see the opt-in setting below.

On startup, the daemon logs its effective listener/backend addresses, SSH file
paths, timeouts, connection limit, user mappings and public key fingerprints.
Each provider gets a summary with its model, danger threshold, endpoint origin,
and credential source (inline, file, environment, Codex home, or none).
Credential file paths and environment variable names are shown, but their contents
and values are never read for logging. Endpoint paths, userinfo, queries and
fragments are omitted because they may contain credentials. Startup summaries
appear before SSH initialization; the separate `listening` event confirms readiness.

## Approved-command cache

The daemon remembers commands approved by **every** configured reviewer. An exact
repeat from the same frontend user, mapped to the same backend user, skips all
LLM requests. Whitespace and every command byte must match. SSH authentication,
static policy, command rules and backend execution still run on approved requests; command
output is never cached. An approval is cached even if subsequent execution fails.

The cache stores only hashes in memory, holds at most `approval_cache_entries`
entries, and evicts the least recently used entry when full. It has no time-based
expiry and clears when the daemon restarts. Provider configuration is fixed for
the life of the daemon, so restarting to change models or thresholds also clears
old approvals. Set `approval_cache_entries: -1` in `server.yaml` to disable it.

Denials, partial approvals, provider errors and malformed verdicts are never
cached. Concurrent first requests can each call the reviewers; once an approval
is stored, later exact repeats use it. `check` applies local command policy, then
performs fresh model reviews when needed; `providers-test` always calls models.
Neither CLI command uses the daemon cache. JSON logs include `msg: "approval cache"`, `hit`, `safe`, and the
command hash, without command text. With `log_commands: true`, a separate
`ssh command received` event records the command even on cache hits.

The cache reuses a review decision, not a guarantee that files, URL destinations
or other external state are unchanged. Restart the daemon to clear approvals
when you need fresh decisions after such changes.

## Lab key discovery

In `server.yaml`, enable:

```yaml
accept_all_keys: true
users:
  alice:
    backend_user: alice
    fingerprints: []
```

This accepts any client public key **for configured users with a backend mapping**.
Clients must still prove possession of the private key. Unknown users and users
without a backend mapping are rejected. This also accepts destination SSH host
keys that fail trust checks, including unknown, changed, or revoked keys. Command
review and backend user authentication remain required. Missing or empty
`authorized_keys` and `known_hosts` files are allowed in this mode; malformed or
unreadable files and unsupported authorized_keys options still fail startup.

Startup emits `ACCEPT ALL SSH KEYS ENABLED` at ERROR level. Each successful login
with an otherwise unauthorized key emits `ssh untrusted key accepted` at ERROR
level, including `user`, `backend_user`, `remote_addr`, `fingerprint`, `reason`,
`authorized_key` (the public-key line), and `signature_verified: true`. Unsigned key
offers have a different message and `signature_verified: false`.

Collect successfully authenticated fingerprints for inspection:

```sh
podman logs ssh-ballpit 2>&1 | jq -r '
  select(.msg == "ssh untrusted key accepted" and .signature_verified == true)
  | [.user, .remote_addr, .fingerprint] | @tsv' | sort -u
```

Print the corresponding public keys for one user:

```sh
podman logs ssh-ballpit 2>&1 | jq -r '
  select(.msg == "ssh untrusted key accepted" and .signature_verified == true and .user == "alice")
  | .authorized_key' | sort -u
```

Each bypassed backend host-key check emits `ssh untrusted backend host key accepted`
at ERROR level, with `backend`, `backend_user`, `remote_addr`, `fingerprint`,
`key_type`, `reason`, and `known_hosts_entry`. Its `signature_verified: true` means
the server proved possession of the host private key during key exchange; it does
not establish trust or mean backend user authentication succeeded.

Collect destination fingerprints and their corresponding known_hosts lines:

```sh
podman logs ssh-ballpit 2>&1 | jq -r '
  select(.msg == "ssh untrusted backend host key accepted" and .signature_verified == true)
  | [.backend, .fingerprint, .known_hosts_entry] | @tsv' | sort -u
```

After you trust the destination fingerprint, copy its `known_hosts_entry` into
the configured `known_hosts` file, replacing any stale entry for that destination.
The logged line includes the port when needed. Review any existing revocation
before removing it.

After you trust the collected client identities, put their fingerprints under that
user's `fingerprints` list, and their public-key lines in `authorized_keys`.
Set `accept_all_keys: false` and restart to enforce them. No keys are automatically
added to configuration. Logged public keys are public material; private keys and
provider credentials are never included.

### Discover and save local sshd host keys

On the host running ssh-ballpit, run these commands from the
repository directory. Set the address and port to match `backend` in
`config/server.yaml`, and the file path to the host-side bind mount for
`known_hosts`. The defaults below match `127.0.0.1:22` and `config/known_hosts`.
These commands require OpenSSH client tools.

Discover the keys directly from sshd and display their fingerprints:

```bash
backend_host=127.0.0.1
backend_port=22
known_hosts=config/known_hosts

ssh-keyscan -T 5 -p "$backend_port" "$backend_host" > "$known_hosts.scan" &&
ssh-keygen -lf "$known_hosts.scan" -E sha256
```

Before installing them, compare the scanned fingerprints with the host's own
public host-key files (use the configured public-key paths if sshd uses custom
ones):

```bash
for key in /etc/ssh/ssh_host_*_key.pub; do
  [ -f "$key" ] && ssh-keygen -lf "$key" -E sha256
done
```

Every scanned fingerprint should match a trusted local host key. `ssh-keyscan`
alone does not authenticate the server. It discovers multiple key types, so you
can save all of sshd's offered keys without copying a string from the logs.

After verifying the fingerprints, run this in the same shell to replace entries
for this destination while preserving entries for other hosts:

```bash
(
  set -e
  test -s "$known_hosts.scan"
  ssh-keygen -lf "$known_hosts.scan" -E sha256 >/dev/null
  host_entry="$backend_host"
  if [ "$backend_port" != 22 ]; then
    host_entry="[$backend_host]:$backend_port"
  fi
  touch "$known_hosts"
  ssh-keygen -R "$host_entry" -f "$known_hosts"
  cat "$known_hosts.scan" >> "$known_hosts"
  podman restart ssh-ballpit
)
```

`ssh-keygen -R` backs up the previous file as `config/known_hosts.old` with these
defaults. Review any existing revocation before replacing entries. The scan file
contains only public keys and can be removed afterward. Restarting reloads the
trust file; it does not change the host's sshd configuration.


## Prometheus metrics

Enable the HTTP endpoint in `server.yaml`:

```yaml
metrics_listen: 127.0.0.1:2112
```

Scrape `http://127.0.0.1:2112/metrics`. For a remote Prometheus server, bind to the
host's trusted LAN address or `0.0.0.0:2112` instead. The endpoint is plain HTTP
without authentication, so expose it only to your monitoring network. The existing
`--network=host` Podman command needs no additional port mappings.

Example Prometheus configuration (use this host's address):

```yaml
scrape_configs:
  - job_name: ssh-ballpit
    scrape_interval: 15s
    static_configs:
      - targets: ["ballpit-host:2112"]
```

Metrics use the official Prometheus Go client and `/metrics` handler, following
[the Go instrumentation guide](https://prometheus.io/docs/guides/go-application/).
The listener starts and stops with the daemon. A configured port that cannot be
bound fails startup. An empty `metrics_listen` disables the endpoint.

| Metric (`ssh_ballpit_` prefix) | Meaning / labels |
| --- | --- |
| `connections_total`, `connections_active` | Accepted TCP connections and active connections, including handshakes |
| `connection_limit_rejections_total`, `handshake_failures_total` | Capacity and SSH handshake failures |
| `authentication_attempts_total` | `result`: success/failure; normal none-method discovery excluded |
| `public_key_rejections_total` | Rejected key offers by fixed `reason`; an offer alone does not prove possession |
| `key_authorization_bypasses_total` | Successfully authenticated connections that bypassed key authorization |
| `backend_host_key_bypasses_total` | Backend host-key trust failures bypassed after signature verification |
| `accept_all_keys_enabled` | Client and backend key lab mode: 1 enabled, 0 disabled |
| `protocol_rejections_total` | `kind`: channel/request/malformed_exec |
| `commands_total` | `outcome`: static_denied/blacklist_denied/review_denied/success/nonzero_exit/backend_error |
| `reviews_total` | `reviewer`, `outcome`: approved/denied/error; retries count as one logical review |
| `review_duration_seconds` | Histogram by `reviewer`, including retries and subscription queue time |
| `approval_cache_lookups_total` | `result`: hit/miss/disabled |
| `approval_cache_entries` | Current number of cached approvals |
| `execution_duration_seconds` | Histogram of backend connection and execution time |

Standard `go_*` and `process_*` metrics are also exposed. Counters reset on daemon
restart. Cache hits do not increment review metrics because no provider is called.
Labels contain only configured reviewer names and fixed categories, never client
addresses, usernames, fingerprints, commands, hashes, credentials or raw errors.

## Plaintext command logging

To include command text in the daemon's JSON logs, set in `server.yaml`:

```yaml
log_commands: true
```

Restart the daemon after changing this setting. The default is `false`.
Each authenticated, well-formed SSH exec request emits one INFO event with
`msg: "ssh command received"`, the exact text in `command`, `command_sha256`,
frontend/backend usernames, client address and key fingerprint. It is logged
before policy checks and LLM review, so this covers rejected commands, approved
commands, cache hits and commands whose backend execution fails. This event
records receipt, not approval or execution success.

JSON escaping keeps quotes, tabs and newlines inside one log record. Command
text is not redacted: passwords or tokens included in the command will also be
logged. Provider credentials remain omitted. Response bodies are omitted unless
`verbose_review_logs` is enabled. Command text is never added to Prometheus
metrics. Set both logging options to `false` and restart to stop recording
command text; existing log entries remain.

## Verbose provider diagnostics

For malformed responses or provider failures, enable this in `server.yaml` and
restart the container:

```yaml
verbose_review_logs: true
```

Each actual provider review emits `provider review diagnostic`, at ERROR level
on failure and INFO on success. Fields include reviewer, model, endpoint origin,
user, command hash, processing stage, elapsed time, HTTP status, response byte
count, a response excerpt capped at 16 KiB, and whether it was truncated. Known
API key values are redacted from response excerpts; credential headers are never
logged. Subscription reviews report status/error metadata without raw Codex
output. Gemini diagnostics describe the final response after any retries.

**This option includes plaintext commands and recent history even when
`log_commands: false`.** Arbitrary secrets inside commands or model responses
cannot be reliably redacted. Startup logs a warning when enabled. Turn it off and
restart after troubleshooting.

For Ollama, `ollama_curl` contains a shell-quoted replay of the exact HTTP request
body, including the system prompt, current command, history, model and format.
It uses stdin for the JSON body and a 120-second curl timeout. It sends the review
request only; it does not execute the SSH command. It includes the full endpoint
URL, so any sensitive URL path is also visible. No Ollama API credentials are sent.

Save the latest failed Ollama replay from the logs:

```sh
umask 077
podman logs ssh-ballpit 2>&1 | jq -r '
  select(.msg == "provider review diagnostic" and .kind == "ollama" and .error != null)
  | .ollama_curl // empty' | tail -n 1 > ollama-replay.sh
cat ollama-replay.sh
```

Inspect the script, then run `bash ollama-replay.sh` to see the raw server response.
Do not run scripts taken from untrusted logs. A replay can produce a different
answer if model state or history changes. Approval-cache hits do not call providers
and therefore produce no provider diagnostics. Use `providers-test` for a fresh
review without SSH execution.

If an Ollama diagnostic has `stage: "http_request"` and `http_status: 0`, the
gateway did not receive an HTTP response. Its `transport_reason` describes a
bounded network category such as `connection_refused`, `network_unreachable`,
`timeout`, or `http_transport_error`; it does not report an SSH backend error.
Check reachability from the gateway host, using the endpoint in the provider
YAML:

```sh
ollama_endpoint='http://your-ollama-host:11434'
curl --verbose --connect-timeout 3 --max-time 5 \
  "$ollama_endpoint/api/version"
```

On the Ollama host, check that the service or proxy is running and listening
on the configured address and port (`ss -ltnp`). If a local request succeeds
but the gateway host cannot connect, check the listen address and network
firewall. A failed provider review denies the command even when every other
reviewer approves.

`empty verdict content` means the provider returned no usable verdict text.
`invalid verdict JSON` or `invalid verdict object` indicates malformed JSON or
an unexpected structure (for example, prose, an array, or invalid fields). Ollama
requests use a [structured JSON schema](https://docs.ollama.com/capabilities/structured-outputs)
for the four required verdict fields. Incomplete responses and malformed verdicts
still deny the command. For Ollama-compatible servers that ignore the format
setting, one complete Markdown JSON fence is unwrapped before strict validation
(`ollama_markdown_unwrapped: true` in diagnostics). Prose, extra objects, duplicate
fields and missing fields are still rejected.

Use single quotes to send a literal remote command:

```sh
ssh -i ~/.ssh/id_ed25519 -p 2222 system_user@gateway.example.net 'uname -a'
```

Backticks would execute `uname -a` locally and send its output as the remote command.

## Recent review history

Every LLM review includes `recent_history` with the last five approved commands
and last five rejected commands for the same frontend/backend user pair. This is
automatic and independent of `log_commands` and the approval-cache size setting.
Lists are oldest-to-newest, contain completed decisions only, and exclude the
current request. All reviewers receive the same snapshot. Concurrent in-flight
requests are not included until their decisions have been recorded.

Example user payload sent to each reviewer:

```json
{
  "user": "alice",
  "command": "ls -la",
  "recent_history": {
    "approved": [{"command": "pwd", "source": "reviewers"}],
    "rejected": [{"command": "bash", "source": "static_policy"}]
  }
}
```

Each entry has command text and a decision source:

- `reviewers`: unanimous approval, or at least one valid rejecting verdict.
- `approval_cache`: an approved exact repeat that skipped the LLMs.
- `static_policy`: rejected before calling the LLMs.
- `whitelist`: approved by a command regex after passing static and blacklist checks.
- `blacklist`: rejected by a command regex before consulting the cache or LLMs.
- `review_error`: blocked because review failed, timed out, was cancelled or
  returned an invalid response; this is not itself a danger verdict.

Approvals record permission to execute, even if backend execution later fails.
Repeated requests occupy separate history entries. Cached approvals still skip
LLM calls, but update approved history for the next uncached review. History does
not invalidate previously cached approvals. `check` and `providers-test` are
standalone processes and send empty history lists.

History holds command text only in daemon memory and clears on restart. It is
never shared between user pairs. Past command text—including policy-rejected
commands—is sent to the configured providers on later reviews. History is not added to metrics. `verbose_review_logs` includes history in
Ollama replay commands; `log_commands` records incoming commands separately.
Oversized rejected commands are limited to 16,384 bytes in history and marked
`truncated: true` to bound memory and prompt size. Normal commands are preserved.

The system prompt identifies history as untrusted context and tells reviewers to
assess the current command independently, without treating earlier decisions as
instructions or automatic approval. It also distinguishes review errors from
safety rejections and permits history to inform analysis of related activity or
privilege escalation.
