## Build stage
FROM golang:alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN  go mod download

COPY . .

RUN go build -o bin/go-client-cli main.go

## Deploy stage
FROM alpine

WORKDIR /app

COPY --from=build /app/bin/go-client-cli ./go-client-cli

ENTRYPOINT ["./go-client-cli"]