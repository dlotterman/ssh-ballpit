# Command whitelist and blacklist

Command rules are optional YAML lists beside `server.yaml` (or the legacy JSON
server configuration). A blacklist match denies a command. An eligible whitelist
match skips LLM review and sends the original command text to the backend. Rules
apply to all configured users; they are not a sandbox.

## Set up

Copy and edit the starter files in the same directory as your server configuration:

```sh
cp examples/whitelist.yaml examples/blacklist.yaml config/
```

The daemon loads `whitelist.yaml` and `blacklist.yaml` independently. Each file
contains one `patterns` list of Go regular expressions. Missing files disable
that list; use `patterns: []` for an explicitly empty list. `configure` creates
both starter files alongside its server and provider files and refuses to
overwrite existing files.

Invalid YAML, unknown fields, duplicate keys, multiple documents, non-string or
empty patterns, invalid regexes, and unreadable files stop startup. Files are
limited to 1 MiB and read once at startup; restart after changes. Both `serve` and `check` accept repeatable `--whitelist-file` and
`--blacklist-file` flags. Flag paths resolve from the process working directory.
Each flag adds patterns to the corresponding default list. A missing explicitly
named file is an error. Every blacklist takes precedence over every whitelist.

```sh
ssh-ballpit serve --config config/server.yaml \
  --whitelist-file rules/systemd.yaml --whitelist-file rules/storage.yaml \
  --blacklist-file rules/site.yaml --blacklist-file rules/maintenance.yaml
```

### Collect a suggested whitelist

Pass `serve --suggest-whitelist DIR` with an existing private directory outside
`providers/`. Each start creates a fresh mode-0600 `suggested-whitelist-*.yaml`
and logs its path. After a model-approved command (including a cache hit) runs
and returns an exit status, the daemon records it once if it qualifies for
simple-command whitelisting. A nonzero backend exit still counts as execution.
Existing whitelist matches, denials and backend failures produce no suggestions.

The file contains exact command patterns with regex metacharacters escaped.
Command arguments can contain secrets even when command logging is off.
Inspect the file before using it with `--whitelist-file` on a later start.
Promoting a suggestion makes it available to **all** mapped users. Collection
stops adding entries that would exceed the 1 MiB rule-file limit and logs a
write warning. Recording errors leave the prior file and SSH result intact.

Example `whitelist.yaml`:

```yaml
patterns:
  - 'ls -al /home/.*'
  - 'systemd-analyze --user verify'
```

Example `blacklist.yaml`:

```yaml
patterns:
  - 'reboot'
  - 'dnf(?:\s+(?s:.*))?'
```

The source files are [`examples/whitelist.yaml`](../examples/whitelist.yaml) and
[`examples/blacklist.yaml`](../examples/blacklist.yaml).

For a broader opt-in set of unprivileged system diagnostics, see
[`examples/debug_whitelist.yaml`](../examples/debug_whitelist.yaml). Copy only
the patterns you want into `config/whitelist.yaml`; the extra example file is
not loaded by the daemon. It includes bounded `ss -N NAME` and `ip -n NAME`
forms for named network namespaces. Those commands may need permissions that
the backend account does not have. `ss -p` can likewise omit processes owned
by other users. The example uses `df -hT` for filesystem capacity and `du -cs`
for directory totals because `df -cs` is not a valid `df` invocation.

### Extended debug example

[`debug_whitelist_extended.yaml`](../examples/debug_whitelist_extended.yaml) includes
the short debug rules along with system and user unit dependencies, selected
`systemctl show` properties, current and previous boot journals, SMART identity,
USB power, XFS error settings, device-mapper relationships and GPU counters.
The last section provides commented templates for checks tied to local paths or
services. Edit those paths and names before enabling a template. The file is
opt-in and applies to every mapped user. Unit definitions, command lines and
journals can contain sensitive values.

From the checkout, preview a decision using the same rule file you plan to serve:

```sh
ssh-ballpit check --config config/server.yaml --user alice \
  --whitelist-file examples/debug_whitelist_extended.yaml \
  --command 'journalctl -k -b --no-pager -o short-iso -n 50'
```

A whitelist match reports `"source":"whitelist"`. To enable the file, pass the
same flag to `serve`. For a container, copy the file into the mounted config
directory and use its container path, such as `/config/debug_whitelist_extended.yaml`.

## Matching

Patterns match the **whole command**, are case-sensitive, and are automatically
anchored even when they contain alternation or multiline flags. They use Go regex
syntax, not shell glob syntax: `.*` matches any suffix (except newlines by default),
so `ls -al /home/.*` expresses the requested `ls -al /home/*` wildcard, including
nested paths. A bare `*` repeats the preceding regex item. Use YAML single quotes
to preserve backslashes. Lookaround and backreferences are unsupported.

`systemd-analyze --user verify` is one exact command: extra arguments or whitespace
do not match. `reboot` is also exact; to block it with any arguments, use
`reboot(?:\s+(?s:.*))?`. The example `dnf` rule blocks bare `dnf` and all arguments,
including tabs and newlines.

The decision order is:

1. Existing static shell policy. No rule can override a static rejection.
2. Blacklist. A match denies with exit 126, even if also whitelisted or cached.
3. Whitelist. A match on one eligible simple command skips the cache and LLMs.
4. Exact approval cache, followed by unanimous model review on a miss.

Whitelist matching uses the original command text. To keep wildcard suffixes from
approving additional shell operations, bypass requires one foreground simple
command without assignments, redirections, pipelines, command lists, comments,
shell control structures or expansions other than pathname globs. Plain quoted
arguments are allowed. Thus `ls -al /home/alice; touch /tmp/file` and
`ls -al /home/alice > /tmp/file` still need normal review, while command
substitutions remain statically denied. Commands outside the whitelist are not
automatically denied; they keep the normal cache/model path.

Blacklist matching checks the original text and each parsed simple invocation
inside compound shell commands. For these invocations it also checks a version
with one space between argument words, preserving each word's original spelling.
This catches `uptime; reboot`, `MODE=test reboot`, and whitespace variations.
Matching does not execute shell expansions, resolve aliases, inspect scripts,
or identify every wrapper program; regex rules do not provide a sandbox.

## Observe a decision

Local decisions are recorded in history with source `whitelist` or `blacklist`.
Whitelist approvals are not added to the model approval cache. Logs include the
source and command hash, and startup logs include rule counts; patterns and raw
commands are not logged by default. Blacklist denials increment the
`commands_total{outcome="blacklist_denied"}` metric.

`check --user alice --command 'systemd-analyze --user verify'` applies the same
rules without executing anything; a local decision includes `source`, `safe`, and
an empty `reviewers` list in its JSON output. `providers-test` always exercises
providers, regardless of command rules. At least one reviewer is still required
for unmatched commands. The [connection flow](connection-flow.md) shows where
these decisions occur in an SSH session.
