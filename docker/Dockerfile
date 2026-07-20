## This Dockerfile builds the Crypto Broker CLI Go test application
## This is a simple CLI application that can be used to test the Crypto Broker Server. It is not intended for production use.

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.23

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

WORKDIR /app

COPY --from=build /app/bin/go-client-cli ./go-client-cli

# Run as non-root uid 1000 so the CLI can connect to the crypto-broker Unix
# socket, which the server sidecar (also uid 1000) owns. A different uid could
# not connect to the default 0755 socket.
USER 1000:1000

ENTRYPOINT ["./go-client-cli"]
