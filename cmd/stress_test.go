package cmd

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestStressTestRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{"--count=0"}, {"--count=-1"}, {"--count=invalid"},
		{"--concurrent=0"}, {"--concurrent=100001"},
		{"--num=0"}, {"--num=1000001"}, {"unexpected"},
	} {
		cmd := newStressTestCommand()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Errorf("accepted invalid arguments: %s", strings.Join(args, " "))
		}
	}
}

func TestStressTestExitStatus(t *testing.T) {
	if os.Getenv("STRESS_TEST_EXIT_CHECK") == "true" {
		rootCmd.SetArgs(strings.Fields(os.Getenv("STRESS_TEST_ARGS")))
		Execute()
		return
	}
	for _, tt := range []struct {
		args string
		exit int
	}{
		{"stress-test --count=0", 1},
		{"stress-test --count=invalid", 1},
		{"stress-test --unexpected", 1},
		{"stress-test --help", 0},
		{"version --unexpected", 0},
		{"unknown-command", 0},
	} {
		t.Run(tt.args, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestStressTestExitStatus$")
			cmd.Env = append(os.Environ(), "STRESS_TEST_EXIT_CHECK=true", "STRESS_TEST_ARGS="+tt.args)
			output, err := cmd.CombinedOutput()
			if cmd.ProcessState == nil {
				t.Fatal(err)
			}
			if got := cmd.ProcessState.ExitCode(); got != tt.exit {
				t.Fatalf("exit status %d, want %d: %s", got, tt.exit, output)
			}
		})
	}
}

type stressConnectionKey struct{}

type stressConnections struct {
	opened atomic.Int64
	closed atomic.Int64
}

func (s *stressConnections) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, stressConnectionKey{}, s.opened.Add(1))
}

func (s *stressConnections) HandleConn(_ context.Context, event stats.ConnStats) {
	if _, ok := event.(*stats.ConnEnd); ok {
		s.closed.Add(1)
	}
}

func (*stressConnections) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (*stressConnections) HandleRPC(context.Context, stats.RPCStats) {}

func TestStressTestConnectionsRequestsAndRuns(t *testing.T) {
	if os.Getenv("STRESS_TEST_INTEGRATION") != "true" {
		t.Skip("set STRESS_TEST_INTEGRATION=true with the broker socket unused")
	}
	// The client library has a fixed socket path; never replace an existing socket.
	if err := os.MkdirAll("/tmp/open-crypto-broker", 0700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", "/tmp/open-crypto-broker/crypto-broker-server.sock")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	var connections stressConnections
	var received atomic.Int64
	var mu sync.Mutex
	requestsPerConnection := make(map[int64]int)
	ready := []chan struct{}{make(chan struct{}), make(chan struct{})}
	server := grpc.NewServer(grpc.StatsHandler(&connections), grpc.UnknownServiceHandler(func(_ any, stream grpc.ServerStream) error {
		method, _ := grpc.MethodFromServerStream(stream)
		if method != "/CryptoBroker.CryptoGrpc/HashData" {
			return status.Errorf(codes.Unimplemented, "unexpected method %s", method)
		}
		if err := stream.RecvMsg(&emptypb.Empty{}); err != nil {
			return err
		}
		mu.Lock()
		requestsPerConnection[stream.Context().Value(stressConnectionKey{}).(int64)]++
		mu.Unlock()
		n := received.Add(1)
		gate := ready[(n-1)/6]
		if n%6 == 0 {
			close(gate)
		}
		select {
		case <-gate:
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
		if n == 7 {
			return status.Error(codes.InvalidArgument, "test RPC failure")
		}
		return stream.SendMsg(&emptypb.Empty{})
	}))
	t.Cleanup(server.Stop)
	go func() { _ = server.Serve(listener) }()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var output bytes.Buffer
	cmd := newStressTestCommand()
	cmd.SetOut(&output)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--concurrent=2", "--num=3", "--count=2"})
	if err := cmd.ExecuteContext(ctx); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two benchmark summaries: %s", output.String())
	}
	for run, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "BenchmarkStressHashConcurrentConnections-") || fields[1] != "1" || len(fields)%2 != 0 {
			t.Fatalf("invalid benchmark summary: %s", line)
		}
		metrics := make(map[string]float64)
		for i := 2; i < len(fields); i += 2 {
			value, err := strconv.ParseFloat(fields[i], 64)
			if err != nil {
				t.Fatal(err)
			}
			metrics[fields[i+1]] = value
		}
		if metrics["connections"] != 2 || metrics["requests/connection"] != 3 || metrics["requests"] != 6 || metrics["requests/s"] <= 0 || metrics["grpc_OK"] != float64(6-run) || metrics["grpc_InvalidArgument"] != float64(run) || metrics["ns/op"] <= 0 || metrics["B/op"] <= 0 || metrics["allocs/op"] <= 0 {
			t.Fatalf("unexpected metrics for run %d: %s", run, line)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requestsPerConnection) != 4 {
		t.Fatalf("got %d connections, want 4", len(requestsPerConnection))
	}
	for connection, requests := range requestsPerConnection {
		if requests != 3 {
			t.Errorf("connection %d received %d requests, want 3", connection, requests)
		}
	}
	for connections.closed.Load() != 4 {
		select {
		case <-ctx.Done():
			t.Fatalf("closed %d connections, want 4", connections.closed.Load())
		case <-time.After(time.Millisecond):
		}
	}
}
