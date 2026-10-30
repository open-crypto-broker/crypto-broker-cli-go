package command

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/env"
)

func setBenchmarkParallelism(b *testing.B) {
	multiplicator, err := benchmarkParallelismMultiplicator()
	if err != nil {
		b.Fatal(err)
	}

	b.SetParallelism(multiplicator)
	b.ReportMetric(
		float64(runtime.GOMAXPROCS(0))*float64(multiplicator),
		"parallel_workers",
	)
}

func benchmarkParallelismMultiplicator() (int, error) {
	rawValue := os.Getenv(env.BENCHMARK_PARALLELISM_MULTIPLICATOR)
	if rawValue == "" {
		return 1, nil
	}

	multiplicator, err := strconv.Atoi(rawValue)
	if err != nil || multiplicator < 1 {
		return 0, fmt.Errorf(
			"%s must be a positive integer, got %q",
			env.BENCHMARK_PARALLELISM_MULTIPLICATOR,
			rawValue,
		)
	}

	return multiplicator, nil
}

func TestBenchmarkParallelismMultiplicator(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "unset uses default", want: 1},
		{name: "positive value", value: "4", want: 4},
		{name: "zero is invalid", value: "0", wantErr: true},
		{name: "negative value is invalid", value: "-1", wantErr: true},
		{name: "non-integer value is invalid", value: "invalid", wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(env.BENCHMARK_PARALLELISM_MULTIPLICATOR, testCase.value)

			got, err := benchmarkParallelismMultiplicator()
			if (err != nil) != testCase.wantErr {
				t.Fatalf("benchmarkParallelismMultiplicator() error = %v, wantErr %t", err, testCase.wantErr)
			}
			if got != testCase.want {
				t.Fatalf("benchmarkParallelismMultiplicator() = %d, want %d", got, testCase.want)
			}
		})
	}
}
