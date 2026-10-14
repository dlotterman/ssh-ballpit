# CI/CD and builds

Run build commands from the repository root. Application source and tests live
in [`cmd/ssh-ballpit/`](../cmd/ssh-ballpit/), embedded starter configuration in
[`examples/`](../examples/), and deployment files in [`packaging/`](../packaging/).
The required Go version is in [`go.mod`](../go.mod).

## Build locally

Build a host binary:

```sh
go build -o bin/ssh-ballpit ./cmd/ssh-ballpit
```

`bin/` is ignored by Git. To build a local image for development, use Podman:

```sh
podman build -f Containerfile -t localhost/ssh-ballpit .
```

Docker can build the same recipe with
`docker build -f Containerfile -t localhost/ssh-ballpit .`. The repository has
both `.containerignore` and `.dockerignore` so each engine sends only the Go
module, source, examples, and build recipe as context. The Containerfile runs
Go tests and vet before compiling the gateway binary, then packages it with
Codex in the runtime image.

The [README setup](../README.md#set-up-once) pulls the published image as
`docker.io/dlotterman/ssh-ballpit:latest`. To use a local build with those
commands, replace that image name with `localhost/ssh-ballpit`.

## Check changes before committing

Run the Go checks directly:

```sh
go test -mod=readonly -count=1 ./...
go vet -mod=readonly ./...
```

The [test guide](tests.md) explains what the tests protect and when to run the
race detector. The [pre-commit guide](pre-commit.md) covers Gitleaks, Grype, Go
tests, and vet on the staged tree. Activate the hook once per clone:

```sh
git config --local core.hooksPath .githooks
```

## GitHub Actions

The [CI and container workflow](../.github/workflows/ci.yml) runs on every push.
Its CI job scans pushed commits with Gitleaks, scans a snapshot of the committed
source with Grype (failing on High or Critical findings), then runs Go tests and
vet. These checks run even for documentation-only pushes.

After CI passes, a push to `main` builds and pushes a container only when the
push changes `Containerfile`, `.containerignore`, `.dockerignore`, `go.mod`,
`go.sum`, `cmd/`, or `examples/`. Other files, including README and docs, do not
trigger a container publish. The build uses Docker Buildx for `linux/amd64` and
`linux/arm64` and publishes these tags to Docker Hub:

```text
dlotterman/ssh-ballpit:latest
dlotterman/ssh-ballpit:sha-<full commit SHA>
```

The publish job uses the repository's `DOCKERHUB_USERNAME` and
`DOCKERHUB_TOKEN` secrets. It never runs after a failed CI job. The badges in
the README show the CI and container job states; a documentation-only push
skips the container job.
