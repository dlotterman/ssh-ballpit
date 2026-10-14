FROM docker.io/library/golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY examples ./examples
RUN go test ./... && go vet ./... && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ssh-ballpit ./cmd/ssh-ballpit

FROM docker.io/library/node:22-bookworm-slim
ARG CODEX_VERSION=0.162.1
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global @openai/codex@${CODEX_VERSION} \
    && npm cache clean --force \
    && codex --version
COPY --from=build /out/ssh-ballpit /usr/local/bin/ssh-ballpit
RUN mkdir -p /etc/ssh-ballpit /var/lib/ssh-ballpit /workspace \
    && chown node:node /var/lib/ssh-ballpit /workspace
ENV SSHBALLPIT_CONFIG=/config/server.yaml
# Podman --userns=keep-id selects the host UID; do not pin an image USER.
WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/ssh-ballpit"]
CMD ["serve"]
