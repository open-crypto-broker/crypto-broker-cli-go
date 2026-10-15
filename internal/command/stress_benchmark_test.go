package command

import (
	"os"
	"strconv"
	"testing"
)

const (
	stressEnvEnabled        = "STRESS_BENCHMARK_ENABLED"
	stressEnvConcurrent     = "STRESS_CONCURRENT"
	stressEnvCount          = "STRESS_COUNT"
	defaultStressConcurrent = 100
	defaultStressCount      = 100
)

// BenchmarkStressHashConcurrentConnections measures the same wave as the CLI.
// STRESS_COUNT is requests per connection; go test -count controls repeat runs.
func BenchmarkStressHashConcurrentConnections(b *testing.B) {
	if os.Getenv(stressEnvEnabled) != "true" {
		b.Skipf("set %s=true to run broker stress benchmark", stressEnvEnabled)
	}
	concurrent := stressEnvInt(b, stressEnvConcurrent, defaultStressConcurrent, 1, stressMaxConcurrentWorkers)
	num := stressEnvInt(b, stressEnvCount, defaultStressCount, 1, stressMaxRequestsPerWorker)
	b.ReportAllocs()
	var result StressResult
	for b.Loop() {
		var err error
		result, err = RunStressTest(b.Context(), concurrent, num)
		if err != nil {
			b.Fatalf("stress benchmark failed: %v", err)
		}
	}
	b.ReportMetric(float64(result.Connections), "connections")
	b.ReportMetric(float64(result.RequestsPerConnection), "requests/connection")
	b.ReportMetric(float64(result.Requests), "requests")
	b.ReportMetric(float64(result.Requests)*float64(b.N)/b.Elapsed().Seconds(), "requests/s")
	for code, count := range result.StatusCounts {
		b.ReportMetric(float64(count), "grpc_"+code)
	}
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
