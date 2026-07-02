package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"github.com/open-crypto-broker/crypto-broker-client-go/interceptor"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultStressProfile      = "Default"
	defaultStressInput        = "stress-test"
	defaultStressConnections  = 10
	defaultStressConcurrency  = 100
	defaultStressDuration     = 10 * time.Second
	defaultStressTimeout      = 5 * time.Second
	defaultStressConnectLimit = 5 * time.Second

	maxStressConnections = 1_000
	maxStressConcurrency = 100_000
)

var latencyBucketBounds = [...]time.Duration{
	time.Millisecond,
	2 * time.Millisecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
	20 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	200 * time.Millisecond,
	500 * time.Millisecond,
	time.Second,
	2 * time.Second,
	5 * time.Second,
	10 * time.Second,
}

// StressConfig controls a bounded client-side stress run against the broker.
type StressConfig struct {
	Connections    int
	Concurrency    int
	Requests       uint64
	Duration       time.Duration
	Timeout        time.Duration
	ConnectTimeout time.Duration
	Profile        string
	Input          []byte
}

// DefaultStressConfig returns conservative defaults that are useful locally but
// still explicit enough to exercise concurrent request handling.
func DefaultStressConfig() StressConfig {
	return StressConfig{
		Connections:    defaultStressConnections,
		Concurrency:    defaultStressConcurrency,
		Duration:       defaultStressDuration,
		Timeout:        defaultStressTimeout,
		ConnectTimeout: defaultStressConnectLimit,
		Profile:        defaultStressProfile,
		Input:          []byte(defaultStressInput),
	}
}

// Validate rejects misleading or unsafe stress-test settings before any
// connections are opened.
func (config StressConfig) Validate() error {
	if config.Connections < 1 || config.Connections > maxStressConnections {
		return fmt.Errorf("'connections' must be between 1 and %d", maxStressConnections)
	}

	if config.Concurrency < 1 || config.Concurrency > maxStressConcurrency {
		return fmt.Errorf("'concurrency' must be between 1 and %d", maxStressConcurrency)
	}

	if config.Connections > config.Concurrency {
		return fmt.Errorf("'connections' must not exceed 'concurrency'")
	}

	if config.Requests == 0 && config.Duration <= 0 {
		return fmt.Errorf("'duration' must be positive when 'requests' is not set")
	}

	if config.Duration < 0 {
		return fmt.Errorf("'duration' must not be negative")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("'timeout' must be positive")
	}

	if config.ConnectTimeout <= 0 {
		return fmt.Errorf("'connect-timeout' must be positive")
	}

	if config.Profile == "" {
		return fmt.Errorf("'profile' must not be empty")
	}

	return nil
}

// Stress creates many client connections and in-flight Hash requests to stress
// the broker server's gRPC request handling.
type Stress struct {
	logger *slog.Logger
}

// NewStress initializes the stress command.
func NewStress(logger *slog.Logger) (*Stress, error) {
	return &Stress{logger: logger}, nil
}

// Run executes the stress test and logs an aggregate summary.
func (command *Stress) Run(ctx context.Context, config StressConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	command.logger.Info("Opening stress test client connections",
		"connections", config.Connections,
		"concurrency", config.Concurrency,
	)

	libraries, err := command.openLibraries(ctx, config)
	if err != nil {
		return err
	}
	defer closeLibraries(command.logger, libraries)

	runCtx := ctx
	if config.Duration > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, config.Duration)
		defer cancel()
	}

	calls := make([]stressCall, len(libraries))
	for i, library := range libraries {
		lib := library
		calls[i] = func(ctx context.Context) error {
			_, err := lib.HashData(ctx, cryptobrokerclientgo.HashDataPayload{
				Profile: config.Profile,
				Input:   config.Input,
				Metadata: &cryptobrokerclientgo.Metadata{
					Id: uuid.New().String(),
				},
			})

			return err
		}
	}

	command.logger.Info("Running stress test",
		"duration", config.Duration.String(),
		"request_limit", config.Requests,
		"profile", config.Profile,
		"input_size", len(config.Input),
	)

	result := runStress(runCtx, ctx, config, calls)
	command.logSummary(config, result)

	return nil
}

func (command *Stress) openLibraries(ctx context.Context, config StressConfig) ([]*cryptobrokerclientgo.Library, error) {
	libraries := make([]*cryptobrokerclientgo.Library, 0, config.Connections)

	for i := 0; i < config.Connections; i++ {
		connectCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
		lib, err := cryptobrokerclientgo.NewLibrary(connectCtx,
			// Stress runs should expose raw server pressure signals instead of
			// hiding them behind client retries or an open circuit breaker.
			interceptor.RetryConfig{
				MaxAttempts:          1,
				InitialBackoff:       "0s",
				BackoffMultiplier:    1,
				RetryableStatusCodes: nil,
			},
			interceptor.CircuitConfig{
				Name:                fmt.Sprintf("crypto-grpc-stress-%d", i),
				MaxRequests:         1,
				Interval:            "1h",
				Timeout:             "1s",
				ConsecutiveFailures: 1,
				FailureStatusCodes:  nil,
			},
		)
		cancel()
		if err != nil {
			closeLibraries(command.logger, libraries)
			return nil, fmt.Errorf("open stress connection %d: %w", i+1, err)
		}

		libraries = append(libraries, lib)
	}

	return libraries, nil
}

func (command *Stress) logSummary(config StressConfig, result StressResult) {
	duration := result.Duration()
	requestsPerSecond := 0.0
	if duration > 0 {
		requestsPerSecond = float64(result.Total) / duration.Seconds()
	}

	command.logger.Info("Stress test summary",
		"connections", config.Connections,
		"concurrency", config.Concurrency,
		"total_requests", result.Total,
		"successful_requests", result.Successful,
		"failed_requests", result.Failed,
		"requests_per_second", requestsPerSecond,
		"duration", duration.String(),
		"latency_avg", result.AverageLatency().String(),
		"latency_min", result.MinLatency.String(),
		"latency_max", result.MaxLatency.String(),
		"latency_p95_upper_bound", result.PercentileUpperBound(95).String(),
		"latency_p99_upper_bound", result.PercentileUpperBound(99).String(),
		"latency_buckets", result.NonEmptyLatencyBuckets(),
		"status_codes", result.StatusCodesByName(),
	)
}

type stressCall func(context.Context) error

type stressCallResult struct {
	code    codes.Code
	latency time.Duration
}

// StressResult contains aggregate stress-test measurements.
type StressResult struct {
	StartedAt  time.Time
	FinishedAt time.Time

	Total      uint64
	Successful uint64
	Failed     uint64

	StatusCodes map[codes.Code]uint64

	TotalLatency time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration

	LatencyBuckets [len(latencyBucketBounds) + 1]uint64
}

// LatencyBucket contains the number of requests observed in one latency range.
type LatencyBucket struct {
	UpperBound string `json:"upperBound"`
	Count      uint64 `json:"count"`
}

func runStress(runCtx context.Context, requestParentCtx context.Context, config StressConfig, calls []stressCall) StressResult {
	results := make(chan stressCallResult, config.Concurrency*2)
	result := StressResult{
		StartedAt:   time.Now(),
		StatusCodes: make(map[codes.Code]uint64),
	}

	var wg sync.WaitGroup
	var nextRequest uint64

	for workerID := 0; workerID < config.Concurrency; workerID++ {
		call := calls[workerID%len(calls)]
		wg.Add(1)

		go func() {
			defer wg.Done()

			for shouldStartRequest(runCtx, config.Requests, &nextRequest) {
				requestCtx, cancel := context.WithTimeout(requestParentCtx, config.Timeout)
				startedAt := time.Now()
				err := call(requestCtx)
				latency := time.Since(startedAt)
				cancel()

				results <- stressCallResult{
					code:    statusCode(err),
					latency: latency,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for callResult := range results {
		result.record(callResult)
	}

	result.FinishedAt = time.Now()

	return result
}

func shouldStartRequest(ctx context.Context, requests uint64, nextRequest *uint64) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	if requests == 0 {
		return true
	}

	return atomic.AddUint64(nextRequest, 1) <= requests
}

func (result *StressResult) record(callResult stressCallResult) {
	result.Total++
	result.StatusCodes[callResult.code]++
	result.TotalLatency += callResult.latency

	if callResult.code == codes.OK {
		result.Successful++
	} else {
		result.Failed++
	}

	if result.MinLatency == 0 || callResult.latency < result.MinLatency {
		result.MinLatency = callResult.latency
	}

	if callResult.latency > result.MaxLatency {
		result.MaxLatency = callResult.latency
	}

	bucketIndex := len(latencyBucketBounds)
	for i, bound := range latencyBucketBounds {
		if callResult.latency <= bound {
			bucketIndex = i
			break
		}
	}
	result.LatencyBuckets[bucketIndex]++
}

// Duration returns the wall-clock duration covered by the aggregate result.
func (result StressResult) Duration() time.Duration {
	if result.FinishedAt.IsZero() {
		return 0
	}

	return result.FinishedAt.Sub(result.StartedAt)
}

// AverageLatency returns the mean request latency.
func (result StressResult) AverageLatency() time.Duration {
	if result.Total == 0 {
		return 0
	}

	return result.TotalLatency / time.Duration(result.Total)
}

// PercentileUpperBound returns the upper bound of the fixed latency bucket
// containing the requested percentile.
func (result StressResult) PercentileUpperBound(percentile int) time.Duration {
	if result.Total == 0 {
		return 0
	}

	target := uint64(math.Ceil(float64(result.Total) * float64(percentile) / 100))
	var seen uint64

	for i, count := range result.LatencyBuckets {
		seen += count
		if seen >= target {
			if i >= len(latencyBucketBounds) {
				return result.MaxLatency
			}

			return latencyBucketBounds[i]
		}
	}

	return result.MaxLatency
}

// StatusCodesByName returns status-code counts with stable, readable keys.
func (result StressResult) StatusCodesByName() map[string]uint64 {
	codesByName := make(map[string]uint64, len(result.StatusCodes))
	for code, count := range result.StatusCodes {
		codesByName[code.String()] = count
	}

	return codesByName
}

// NonEmptyLatencyBuckets returns observed latency buckets in ascending order.
func (result StressResult) NonEmptyLatencyBuckets() []LatencyBucket {
	buckets := make([]LatencyBucket, 0, len(result.LatencyBuckets))
	for i, count := range result.LatencyBuckets {
		if count == 0 {
			continue
		}

		if i >= len(latencyBucketBounds) {
			buckets = append(buckets, LatencyBucket{
				UpperBound: fmt.Sprintf(">%s", latencyBucketBounds[len(latencyBucketBounds)-1]),
				Count:      count,
			})
			continue
		}

		buckets = append(buckets, LatencyBucket{
			UpperBound: fmt.Sprintf("<=%s", latencyBucketBounds[i]),
			Count:      count,
		})
	}

	return buckets
}

func statusCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	if errors.Is(err, context.Canceled) {
		return codes.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return codes.DeadlineExceeded
	}

	return status.Code(err)
}

func closeLibraries(logger *slog.Logger, libraries []*cryptobrokerclientgo.Library) {
	for _, lib := range libraries {
		if err := lib.Close(); err != nil {
			logger.Warn("Failed to close stress test client connection", "error", err)
		}
	}
}
