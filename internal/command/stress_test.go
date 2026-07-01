package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStressConfigValidate(t *testing.T) {
	valid := DefaultStressConfig()

	tests := []struct {
		name    string
		config  StressConfig
		wantErr bool
	}{
		{
			name:    "accepts default configuration",
			config:  valid,
			wantErr: false,
		},
		{
			name: "rejects zero connections",
			config: func() StressConfig {
				config := valid
				config.Connections = 0
				return config
			}(),
			wantErr: true,
		},
		{
			name: "rejects connections above concurrency",
			config: func() StressConfig {
				config := valid
				config.Connections = 2
				config.Concurrency = 1
				return config
			}(),
			wantErr: true,
		},
		{
			name: "rejects unbounded run",
			config: func() StressConfig {
				config := valid
				config.Requests = 0
				config.Duration = 0
				return config
			}(),
			wantErr: true,
		},
		{
			name: "accepts request-bounded run without duration",
			config: func() StressConfig {
				config := valid
				config.Requests = 1
				config.Duration = 0
				return config
			}(),
			wantErr: false,
		},
		{
			name: "rejects empty profile",
			config: func() StressConfig {
				config := valid
				config.Profile = ""
				return config
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunStressHonorsRequestLimit(t *testing.T) {
	config := StressConfig{
		Connections:    2,
		Concurrency:    8,
		Requests:       25,
		Timeout:        time.Second,
		ConnectTimeout: time.Second,
		Profile:        defaultStressProfile,
		Input:          []byte(defaultStressInput),
	}

	result := runStress(context.Background(), context.Background(), config, []stressCall{
		func(context.Context) error { return nil },
		func(context.Context) error { return nil },
	})

	if result.Total != config.Requests {
		t.Fatalf("Total = %d, want %d", result.Total, config.Requests)
	}

	if result.Successful != config.Requests {
		t.Fatalf("Successful = %d, want %d", result.Successful, config.Requests)
	}

	if result.Failed != 0 {
		t.Fatalf("Failed = %d, want 0", result.Failed)
	}

	if got := result.StatusCodes[codes.OK]; got != config.Requests {
		t.Fatalf("StatusCodes[OK] = %d, want %d", got, config.Requests)
	}
}

func TestRunStressRecordsStatusCodes(t *testing.T) {
	config := StressConfig{
		Connections:    1,
		Concurrency:    3,
		Requests:       6,
		Timeout:        time.Second,
		ConnectTimeout: time.Second,
		Profile:        defaultStressProfile,
		Input:          []byte(defaultStressInput),
	}

	result := runStress(context.Background(), context.Background(), config, []stressCall{
		func(context.Context) error {
			return status.Error(codes.ResourceExhausted, "too many streams")
		},
	})

	if result.Total != config.Requests {
		t.Fatalf("Total = %d, want %d", result.Total, config.Requests)
	}

	if result.Successful != 0 {
		t.Fatalf("Successful = %d, want 0", result.Successful)
	}

	if result.Failed != config.Requests {
		t.Fatalf("Failed = %d, want %d", result.Failed, config.Requests)
	}

	if got := result.StatusCodes[codes.ResourceExhausted]; got != config.Requests {
		t.Fatalf("StatusCodes[ResourceExhausted] = %d, want %d", got, config.Requests)
	}
}

func TestRunStressDurationStopsSchedulingWithoutCancelingInflightRequests(t *testing.T) {
	config := StressConfig{
		Connections:    1,
		Concurrency:    1,
		Timeout:        time.Second,
		ConnectTimeout: time.Second,
		Profile:        defaultStressProfile,
		Input:          []byte(defaultStressInput),
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	callStarted := make(chan struct{})
	releaseCall := make(chan struct{})

	var calls int
	resultCh := make(chan StressResult, 1)
	go func() {
		resultCh <- runStress(runCtx, context.Background(), config, []stressCall{
			func(ctx context.Context) error {
				calls++
				close(callStarted)
				<-releaseCall
				return ctx.Err()
			},
		})
	}()

	<-callStarted
	cancelRun()
	close(releaseCall)

	result := <-resultCh
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}

	if result.StatusCodes[codes.OK] != 1 {
		t.Fatalf("StatusCodes[OK] = %d, want 1", result.StatusCodes[codes.OK])
	}
}

func TestStressResultLatencySummary(t *testing.T) {
	result := StressResult{StatusCodes: make(map[codes.Code]uint64)}
	result.record(stressCallResult{code: codes.OK, latency: time.Millisecond})
	result.record(stressCallResult{code: codes.OK, latency: 10 * time.Millisecond})
	result.record(stressCallResult{code: codes.Unavailable, latency: 100 * time.Millisecond})

	if got, want := result.AverageLatency(), 37*time.Millisecond; got != want {
		t.Fatalf("AverageLatency() = %s, want %s", got, want)
	}

	if got, want := result.PercentileUpperBound(95), 100*time.Millisecond; got != want {
		t.Fatalf("PercentileUpperBound(95) = %s, want %s", got, want)
	}

	statusCodes := result.StatusCodesByName()
	if got := statusCodes[codes.OK.String()]; got != 2 {
		t.Fatalf("StatusCodesByName()[OK] = %d, want 2", got)
	}

	latencyBuckets := result.LatencyBucketsByUpperBound()
	if got := latencyBuckets["<=1ms"]; got != 1 {
		t.Fatalf("LatencyBucketsByUpperBound()[<=1ms] = %d, want 1", got)
	}
}

func TestStatusCodeMapsContextErrors(t *testing.T) {
	if got := statusCode(context.Canceled); got != codes.Canceled {
		t.Fatalf("statusCode(context.Canceled) = %s, want %s", got, codes.Canceled)
	}

	if got := statusCode(context.DeadlineExceeded); got != codes.DeadlineExceeded {
		t.Fatalf("statusCode(context.DeadlineExceeded) = %s, want %s", got, codes.DeadlineExceeded)
	}

	if got := statusCode(errors.New("plain error")); got != codes.Unknown {
		t.Fatalf("statusCode(plain error) = %s, want %s", got, codes.Unknown)
	}
}
