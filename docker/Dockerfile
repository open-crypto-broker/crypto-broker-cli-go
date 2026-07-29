## This Dockerfile builds the Crypto Broker CLI Go test application
## This is a simple CLI application that can be used to test the Crypto Broker Server. It is not intended for production use.

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.24

# Runtime user/group id. Must match the uid that owns the crypto-broker server's
# Unix socket (default 1000). Override at build time with --build-arg APP_UID=...
# / APP_GID=... to line up with a server built for a different uid.
ARG APP_UID=1000
ARG APP_GID=1000

## Build stage
# --platform=$BUILDPLATFORM builds on the builder's native architecture and
# cross-compiles to the requested TARGETOS/TARGETARCH (both injected by buildx).
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN  go mod download

COPY . .

# -trimpath makes the build reproducible; -s -w shrink the binary.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o bin/go-client-cli main.go

## Deploy stage
FROM scratch

## Bring the runtime uid/gid build args into this stage's scope.
ARG APP_UID
ARG APP_GID

WORKDIR /app

COPY --from=build /app/bin/go-client-cli ./go-client-cli

# Run as non-root so the CLI can connect to the crypto-broker Unix socket. The
# server creates the socket with mode 0600 owned by APP_UID, so only a process
# running as that same uid (the socket owner) can connect to it.
USER ${APP_UID}:${APP_GID}

ENTRYPOINT ["./go-client-cli"]
