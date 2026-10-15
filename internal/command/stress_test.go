package command

import (
	"context"
	"errors"
	"testing"
)

func TestRunStressTestValidation(t *testing.T) {
	for _, tt := range []struct{ concurrent, num int }{
		{0, 1}, {-1, 1}, {stressMaxConcurrentWorkers + 1, 1},
		{1, 0}, {1, -1}, {1, stressMaxRequestsPerWorker + 1},
	} {
		if _, err := RunStressTest(t.Context(), tt.concurrent, tt.num); err == nil {
			t.Errorf("RunStressTest(%d, %d) accepted invalid input", tt.concurrent, tt.num)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := RunStressTest(ctx, 1, 1); !errors.Is(err, context.Canceled) {
		t.Errorf("canceled run returned %v", err)
	}
}
