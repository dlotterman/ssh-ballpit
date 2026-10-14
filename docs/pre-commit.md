# Pre-commit checks

Run these commands from the repository root. Activate the hook once per clone:

```sh
git config --local core.hooksPath .githooks
```

The hook needs Gitleaks, Grype, and Go 1.26 or newer. In this checkout, Go is
installed outside `PATH`. Save its absolute path in this clone's local Git
configuration so a normal `git commit` can find it:

```sh
git config --local ssh-ballpit.go /tmp/ssh-ballpit-toolchain/go/bin/go
```

On another machine, replace that path with its Go executable. The `/tmp` path
is temporary; update this setting if the toolchain is removed. If `go` is on
`PATH`, the setting is unnecessary. A `GO=/absolute/path/to/go` environment
variable overrides the saved setting for one command.

Before the first commit, set your GitHub noreply address if you do not want your
personal email address in commit metadata:

```sh
git config --local user.email 'YOUR_GITHUB_NOREPLY_ADDRESS'
```

Stage the files you intend to commit, then check that staged snapshot and commit:

```sh
git add -- docs/pre-commit.md
.githooks/pre-commit
git commit -m "Describe the change"
```

The manual hook command checks the staged files without creating a commit. Git
automatically runs the hook again on `git commit` or `git commit .`; you do not
need to invoke it separately unless you want to check before committing. The
example `git add` stages only this document; stage your other intended changes
too. Use `git status --short` to review the staged file list first.

The hook runs Gitleaks, Grype, Go tests, and vet against the staged files. See
[the hook details](../.githooks/README.md) for the exact checks and requirements.
