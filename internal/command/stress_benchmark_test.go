package command

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"github.com/open-crypto-broker/crypto-broker-client-go/interceptor"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	stressEnvEnabled    = "STRESS_BENCHMARK_ENABLED"
	stressEnvConcurrent = "STRESS_CONCURRENT"
	stressEnvCount      = "STRESS_COUNT"

	defaultStressConcurrent    = 100
	defaultStressCount         = 100
	stressConnectionTimeout    = 10 * time.Second
	stressRequestTimeout       = 10 * time.Second
	stressMaxConcurrentWorkers = 100_000
	stressMaxRequestsPerWorker = 1_000_000
)

// BenchmarkStressHashConcurrentConnections opens one client connection per
// parallel worker and sends STRESS_COUNT hash requests on each connection.
//
// Run through Taskfile for a single stress wave:
//
//	task run-stress-tests CONCURRENT=1000 COUNT=100
func BenchmarkStressHashConcurrentConnections(b *testing.B) {
	if os.Getenv(stressEnvEnabled) != "true" {
		b.Skipf("set %s=true to run broker stress benchmark", stressEnvEnabled)
	}

	concurrent := stressEnvInt(b, stressEnvConcurrent, defaultStressConcurrent, 1, stressMaxConcurrentWorkers)
	requestsPerConnection := stressEnvInt(b, stressEnvCount, defaultStressCount, 1, stressMaxRequestsPerWorker)

	previousGOMAXPROCS := runtime.GOMAXPROCS(concurrent)
	defer runtime.GOMAXPROCS(previousGOMAXPROCS)

	b.SetParallelism(1)
	b.ReportAllocs()

	var totalRequests uint64
	var statusCounts [17]atomic.Uint64

	startedAt := time.Now()
	runStressHashWave(b, concurrent, requestsPerConnection, &totalRequests, &statusCounts)
	elapsed := time.Since(startedAt)

	b.ReportMetric(float64(concurrent), "connections")
	b.ReportMetric(float64(requestsPerConnection), "requests/connection")
	b.ReportMetric(float64(totalRequests), "requests")
	if elapsed > 0 {
		b.ReportMetric(float64(totalRequests)/elapsed.Seconds(), "requests/s")
	}

	for codeValue := range statusCounts {
		count := statusCounts[codeValue].Load()
		if count > 0 {
			b.ReportMetric(float64(count), fmt.Sprintf("grpc_%s", codes.Code(codeValue).String()))
		}
	}
}

func runStressHashWave(
	b *testing.B,
	concurrent int,
	requestsPerConnection int,
	totalRequests *uint64,
	statusCounts *[17]atomic.Uint64,
) {
	var workerSequence atomic.Int64
	var readyWorkers atomic.Int64
	var firstErr error
	var firstErrMu sync.Mutex

	ctx := context.Background()
	startRequests := make(chan struct{})

	recordErr := func(format string, args ...any) {
		firstErrMu.Lock()
		defer firstErrMu.Unlock()

		if firstErr == nil {
			firstErr = fmt.Errorf(format, args...)
		}
	}

	b.RunParallel(func(pb *testing.PB) {
		workerID := workerSequence.Add(1)

		connectCtx, cancelConnect := context.WithTimeout(ctx, stressConnectionTimeout)
		lib, err := newStressLibrary(connectCtx, workerID)
		cancelConnect()
		if err != nil {
			recordErr("worker %d: open client connection: %w", workerID, err)
		}
		defer func() {
			if lib != nil {
				if err := lib.Close(); err != nil {
					recordErr("worker %d: close client connection: %w", workerID, err)
				}
			}
		}()

		if readyWorkers.Add(1) == int64(concurrent) {
			close(startRequests)
		}
		<-startRequests

		if err != nil {
			for pb.Next() {
			}
			return
		}

		for i := 0; i < requestsPerConnection; i++ {
			requestCtx, cancelRequest := context.WithTimeout(ctx, stressRequestTimeout)
			_, requestErr := lib.HashData(requestCtx, cryptobrokerclientgo.HashDataPayload{
				Profile: "Default",
				Input:   []byte("stress-test"),
				Metadata: &cryptobrokerclientgo.Metadata{
					Id: uuid.New().String(),
				},
			})
			cancelRequest()

			statusCounts[stressStatusIndex(status.Code(requestErr))].Add(1)
			atomic.AddUint64(totalRequests, 1)
		}

		for pb.Next() {
		}
	})

	if firstErr != nil {
		b.Fatalf("stress benchmark failed: %v", firstErr)
	}
}

func newStressLibrary(ctx context.Context, workerID int64) (*cryptobrokerclientgo.Library, error) {
	return cryptobrokerclientgo.NewLibrary(ctx,
		interceptor.RetryConfig{
			MaxAttempts:          1,
			InitialBackoff:       "0s",
			BackoffMultiplier:    1,
			RetryableStatusCodes: nil,
		},
		interceptor.CircuitConfig{
			Name:                fmt.Sprintf("crypto-grpc-stress-%d", workerID),
			MaxRequests:         1,
			Interval:            "1h",
			Timeout:             "1s",
			ConsecutiveFailures: 1,
			FailureStatusCodes:  nil,
		},
	)
}

func stressEnvInt(b *testing.B, key string, fallback int, minValue int, maxValue int) int {
	rawValue := os.Getenv(key)
	if rawValue == "" {
		return fallback
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		b.Fatalf("%s must be an integer: %v", key, err)
	}

	if value < minValue || value > maxValue {
		b.Fatalf("%s must be between %d and %d", key, minValue, maxValue)
	}

	return value
}

func stressStatusIndex(code codes.Code) int {
	index := int(code)
	if index < 0 || index >= 17 {
		return int(codes.Unknown)
	}

	return index
}
