# ssh-ballpit

[![CI](https://img.shields.io/github/check-runs/dlotterman/ssh-ballpit/main?nameFilter=CI&label=CI)](https://github.com/dlotterman/ssh-ballpit/actions/workflows/ci.yml)
[![Container build](https://img.shields.io/github/check-runs/dlotterman/ssh-ballpit/main?nameFilter=Container%20build&label=container%20build)](https://github.com/dlotterman/ssh-ballpit/actions/workflows/ci.yml)

SSH commands pass static checks and optional YAML blacklist/whitelist rules before
being forwarded to local sshd. Whitelisted commands skip LLM review; other commands
need approval from **every configured LLM**, or a cached unanimous approval.
Any rejection or failed review blocks execution. Existing sshd stays unchanged.
Supports individual commands, not interactive shells or ProxyJump.

## Motivation

The author's environment has four types of computers:

- Human workstations with broad access to the network.
- Ollama hosts with accelerators running local models.
- An isolated machine for third-party agents such as Codex and Claude, with
  internet access but limited access to the local network.
- Target machines: lab computers, Raspberry Pis, VMs and containers.

Sometimes an agent needs access to a target machine without a human approving
individual commands. `ssh-ballpit` is the resulting experiment: an SSH gateway
that asks configured LLMs to review commands before forwarding them to sshd.

The original goal was supervision by a locally hosted, open-weight model.
Additional providers can add independent reviews when the local model misses
something.

```mermaid
flowchart LR
    agent["Agent"] -->|SSH command| gateway["ssh-ballpit<br/>Authenticate and check command rules"]
    gateway <-->|Review when needed| reviewers["Configured LLM reviewers"]
    gateway -->|Approved| sshd["Local sshd<br/>Execute as mapped user"]
    gateway -->|Denied| reject["Reject command"]
```

The [connection flow](docs/connection-flow.md) describes authentication, command
decisions, execution and exit statuses in detail.

### Production

This is a personal experiment, shared for the sake of sharing. It is not ready
for production or dependable lab use.

You break it you buy it.

## Set up once

```sh
podman build -t localhost/ssh-ballpit .
mkdir -p config/providers
chmod 700 config
cp examples/server.yaml config/server.yaml
cp examples/providers/ollama.yaml config/providers/
```

Edit `config/server.yaml` with your SSH username and client key fingerprint.
Put these files in `config/`:

- `authorized_keys`: client public keys permitted to use the gateway.
- `backend_key`: private key already accepted by your local sshd; use mode `0600`.
- `known_hosts`: verified host key for your local sshd.

To enable command rules, copy and edit the examples:

```sh
cp examples/whitelist.yaml examples/blacklist.yaml config/
```

The whitelist includes `ls -al /home/.*` (regex equivalent of `ls -al /home/*`)
and the exact command `systemd-analyze --user verify`. The blacklist includes
exact `reboot`, plus `dnf` with any arguments. Blacklist wins; static checks always apply.
Missing files disable that list. See [command rules](docs/configuration.md#command-whitelist-and-blacklist)
for matching semantics and limits.

## Choose providers

Copy the providers you want from [`examples/providers`](examples/providers) into
`config/providers/`, then edit their YAML. **Every `.yaml` or `.yml` file there is
required to approve each command that reaches model review.** Remove a file to
disable that reviewer.

Ollama's endpoint URL lives in `ollama.yaml`. API providers have an `api_key` field.
For OpenAI subscription access, copy `openai-subscription.yaml` and register once:

```sh
podman run --rm -it --userns=keep-id -v "$PWD/config:/config:z" localhost/ssh-ballpit login-openai
```

Enter the printed one-time code at the URL on another device. Login state is
saved in `config/state/codex` and reused by the daemon and host binary.

## Run

```sh
podman run --rm --name ssh-ballpit --userns=keep-id --network=host -v "$PWD/config:/config:z" localhost/ssh-ballpit
```

```sh
ssh -T -p 2222 alice@127.0.0.1 'uname -a'
```

The gateway host key is generated once in `config/state/host_key`. Provider and
command-rule files are read at startup; restart after editing them. The default
listener is local only; change `listen` in `server.yaml` to accept remote clients.

These commands use rootless Podman on Linux. The single bind mount keeps
configuration, SSH keys and login state on your host. LLM review is a heuristic,
not a sandbox; direct connections to the existing sshd bypass review.

Optional [lab key discovery](docs/configuration.md#lab-key-discovery) and
[Prometheus metrics](docs/configuration.md#prometheus-metrics) are configured in `server.yaml`.

[Configuration and troubleshooting](docs/configuration.md)

## Development checks

Application source and tests live in [`cmd/ssh-ballpit`](cmd/ssh-ballpit), embedded
starter configs in [`examples`](examples), and deployment files in
[`packaging`](packaging). Build from the repository root:

```sh
go build -o bin/ssh-ballpit ./cmd/ssh-ballpit
go test ./...
go vet ./...
```

See [tests and checks](docs/tests.md) for the test groups, fixtures, and CI gates.

Enable the [pre-commit checks](docs/pre-commit.md) for Gitleaks, Grype, Go tests,
and vet:

```sh
git config --local core.hooksPath .githooks
```

GitHub Actions runs the same checks on every push. After they pass, a push to
`main` builds and publishes `dlotterman/ssh-ballpit:latest` and a commit-specific
`sha-<commit>` tag to Docker Hub when the push changes `Containerfile`,
`.containerignore`, `.dockerignore`, Go module files, `cmd/`, or `examples/`. Documentation and
other files outside the image inputs run CI without publishing a new image.
The build uses the repository's `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets.
The container badge shows a skipped job for documentation-only pushes.

## License & Contributing

[The Unlicense](LICENSE) applies only to the original ssh-ballpit code,
configuration, documentation and packaging in this repository. It does not
relicense third-party software imported by the application or bundled in its
container. Those components retain their own licenses, copyrights and notices.

The imported Go software includes **Apache-2.0, BSD-3-Clause and MIT** licenses;
the wider selected module graph also includes **ISC**. These are licenses of
different components or files, not a choice of license for the whole application.
The container additionally includes Codex, Node.js, npm and Debian packages.
Their license metadata and bundled notices include **Artistic-2.0, BSD-2-Clause,
GPL, LGPL, MPL-2.0 and Zlib**, among others. Codex's own Apache-2.0 license does
not cover all of its bundled native libraries; the SBOM lists those separately.

See [`sbom.md`](sbom.md) for exact versions, dependency scope, license evidence,
the runtime container inventory, and regeneration instructions. Retain applicable
upstream license and NOTICE files when redistributing third-party components.

This project is "vibe-coded". Instead of accepting contributions, the author
encourages you to fork it or build your own.
