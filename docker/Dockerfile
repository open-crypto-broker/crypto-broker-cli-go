## This Dockerfile builds the Crypto Broker CLI Go test application
## This is a simple CLI application that can be used to test the Crypto Broker Server. It is not intended for production use.

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.23

## Build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN  go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bin/go-client-cli main.go

## Deploy stage
FROM scratch

WORKDIR /app

COPY --from=build /app/bin/go-client-cli ./go-client-cli

ENTRYPOINT ["./go-client-cli"]
