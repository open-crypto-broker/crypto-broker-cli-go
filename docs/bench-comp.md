# Benchmark comparison

This guide describes steps required to compare two e2e benchmarks regression using `benchstat` tool.

## High level overview

1. Checkout `crypto-broker-server` and `crypto-broker-cli-go` to desired commits - baseline.
2. Collect results of benchmarks regression as "benchmarks_baseline.txt".
3. Checkout `crypto-broker-server` and `crypto-broker-cli-go` to desired commits - newer version.
4. Collect results of benchmarks regression as "benchmarks_new.txt".
5. Compare data using `benchstat` program and interpret with your knowledge or AI assistance

## Detailed steps

### 1. Checkout to baseline version of programs

Turn on console / terminal and checkout to `crypto-broker-server`. Use following command to checkout to baseline version of program (use `git log --oneline` to quickly examine recent commits)

```shell
git checkout <--VERSION--> 
```

where `VERSION` is desired git SHA or TAG.

Repeat the same for `crypto-broker-cli-go`.

### 2. Collect baseline benchmark results

Change directory to `crypto-broker-server`, set up environment variables of server and turn it on. It can be done with

```shell
task run
```

Your terminal session will be occupied with server, therefore please turn on new terminal session and keep the session with `crypto-broker-server` running.

With new shell session, please change directory to `crypto-broker-cli-go` and invoke

```shell
task run-benchmarks
```

This will prove, that `crypto-broker-server` is compatible with `crypto-broker-cli-go` and server cache will warm up.

Next, invoke following to collect baseline data

```shell
OTEL_TRACES_SAMPLER=always_off go test ./... -benchmem -run=^$ -bench ^Benchmark -count=10 > benchmarks_baseline.txt
```

### 3. Checkout to newer version of programs

In the begging, please turn off server in your console. Next, please repeat steps from point 1. That time please checkout to your newer version that you want to compare.

### 4. Collect newer benchmark results

Please repeat the steps from point 2. Last command should be slightly different and look like this

```shell
OTEL_TRACES_SAMPLER=always_off go test ./... -benchmem -run=^$ -bench ^Benchmark -count=10 > benchmarks_new.txt
```

### 5. Compare data with benchstat

As prerequisite, please install `benchstat` with `go install golang.org/x/perf/cmd/benchstat@latest`. It will be downloaded to `$GOPATH/bin` directory. If you have `$GOPATH/bin` in your `$PATH` it should be globally available in terminal.

Change directory to `crypto-broker-cli-go` where you have

- benchmarks_new.txt
- benchmarks_baseline.txt

files. Run following in terminal

```shell
benchstat benchmarks_baseline.txt benchmarks_new.txt
```

To interpret results, please refer to [x/perf/cmd/benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) or ask AI for interpretation.