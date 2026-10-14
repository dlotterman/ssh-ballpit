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
limited to 1 MiB and read once at startup; restart after changes. There are no
extra server settings.

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
