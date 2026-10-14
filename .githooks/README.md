# Pre-commit checks

The hook runs these checks in order and blocks the commit if any fails:

1. Gitleaks on the staged diff, with secrets redacted.
2. Grype on a temporary checkout of the staged tree, failing on High or Critical vulnerabilities.
3. `go test -mod=readonly -count=1 ./...` on that same staged tree.
4. `go vet -mod=readonly ./...` on that same staged tree.

The Gitleaks and Grype checks follow the staged-diff and staged-snapshot pattern.
Ignored credentials and build artifacts are excluded from the staged snapshot.
The directory scan covers the source dependencies; it does not scan the built
container's OS packages or installed Codex CLI.

Install `gitleaks`, `grype`, and Go (the version in `go.mod` or newer) on your PATH,
then activate the hook once per clone:

```sh
git config --local core.hooksPath .githooks
```

Stage your changes first. Run `.githooks/pre-commit` to check the current index
without committing, or run `git commit` to execute the checks automatically.
Unstaged changes do not affect the snapshot. The tests use local test servers and
fake provider responses; live LLM endpoints and provider credentials are not needed.

For a Go installation outside PATH, save its absolute executable path in this
clone, or set `GO` for one command:

```sh
git config --local ssh-ballpit.go /absolute/path/to/go
GO=/absolute/path/to/go git commit
```

See [the pre-commit guide](../docs/pre-commit.md) for this checkout's command.

Grype may need network access to update its vulnerability database. Go may need
network access to download modules or the requested toolchain. Missing tools,
failed downloads, scanner errors, test failures, and vet findings block the commit.
The temporary snapshot is removed when the hook exits.
