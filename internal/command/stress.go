package command

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	stressConnectionTimeout    = 10 * time.Second
	stressRequestTimeout       = 10 * time.Second
	stressMaxConcurrentWorkers = 100_000
	stressMaxRequestsPerWorker = 1_000_000
)

// StressResult reports one wave of concurrent hash requests.
type StressResult struct {
	Connections           int
	RequestsPerConnection int
	Requests              uint64
	StatusCounts          map[string]uint64
}

// RunStressTest opens one connection per worker and sends concurrent requests on each.
// RPC errors are counted by status; connection and cleanup errors fail the run.
func RunStressTest(ctx context.Context, concurrent, requestsPerConnection int) (StressResult, error) {
	result := StressResult{
		Connections:           concurrent,
		RequestsPerConnection: requestsPerConnection,
		StatusCounts:          make(map[string]uint64),
	}
	if concurrent < 1 || concurrent > stressMaxConcurrentWorkers {
		return result, fmt.Errorf("concurrent must be between 1 and %d", stressMaxConcurrentWorkers)
	}
	if requestsPerConnection < 1 || requestsPerConnection > stressMaxRequestsPerWorker {
		return result, fmt.Errorf("num must be between 1 and %d", stressMaxRequestsPerWorker)
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}

	var readyWorkers atomic.Int64
	var statusCounts [17]atomic.Uint64
	var firstErr error
	var firstErrMu sync.Mutex
	var workers sync.WaitGroup
	startRequests := make(chan struct{})

	recordErr := func(workerID int, operation string, err error) {
		firstErrMu.Lock()
		defer firstErrMu.Unlock()
		if firstErr == nil {
			firstErr = fmt.Errorf("worker %d: %s client connection: %w", workerID, operation, err)
		}
	}

	for workerID := range concurrent {
		workers.Go(func() {
			connectCtx, cancelConnect := context.WithTimeout(ctx, stressConnectionTimeout)
			lib, err := cryptobrokerclientgo.NewLibrary(connectCtx)
			cancelConnect()
			if err != nil {
				recordErr(workerID, "open", err)
			} else {
				defer func() {
					if err := lib.Close(); err != nil {
						recordErr(workerID, "close", err)
					}
				}()
			}

			if readyWorkers.Add(1) == int64(concurrent) {
				close(startRequests)
			}
			<-startRequests
			if err != nil || ctx.Err() != nil {
				return
			}

			var requests sync.WaitGroup
			for range requestsPerConnection {
				requests.Go(func() {
					requestCtx, cancelRequest := context.WithTimeout(ctx, stressRequestTimeout)
					defer cancelRequest()
					_, err := lib.HashData(requestCtx, cryptobrokerclientgo.HashDataPayload{
						Profile: "Default",
						Input:   []byte("stress-test"),
						Metadata: &cryptobrokerclientgo.Metadata{
							Id: uuid.New().String(),
						},
					})
					code := status.Code(err)
					if code > codes.Unauthenticated {
						code = codes.Unknown
					}
					statusCounts[code].Add(1)
				})
			}
			requests.Wait()
		})
	}
	workers.Wait()

	for code := range statusCounts {
		if count := statusCounts[code].Load(); count > 0 {
			result.StatusCounts[codes.Code(code).String()] = count
			result.Requests += count
		}
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, firstErr
}
