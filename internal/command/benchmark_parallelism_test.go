package command

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/env"
)

func setTestConcurrency(b *testing.B) {
	multiplicator, err := testConcurrencyMultiplicator()
	if err != nil {
		b.Fatal(err)
	}

	b.SetParallelism(multiplicator)
	b.ReportMetric(
		float64(runtime.GOMAXPROCS(0))*float64(multiplicator),
		"concurrent_workers",
	)
}

func testConcurrencyMultiplicator() (int, error) {
	rawValue := os.Getenv(env.TEST_CONCURRENCY_MULTIPLICATOR)
	if rawValue == "" {
		return 1, nil
	}

	multiplicator, err := strconv.Atoi(rawValue)
	if err != nil || multiplicator < 1 {
		return 0, fmt.Errorf(
			"%s must be a positive integer, got %q",
			env.TEST_CONCURRENCY_MULTIPLICATOR,
			rawValue,
		)
	}

	return multiplicator, nil
}

func TestTestConcurrencyMultiplicator(t *testing.T) {
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
			t.Setenv(env.TEST_CONCURRENCY_MULTIPLICATOR, testCase.value)

			got, err := testConcurrencyMultiplicator()
			if (err != nil) != testCase.wantErr {
				t.Fatalf("testConcurrencyMultiplicator() error = %v, wantErr %t", err, testCase.wantErr)
			}
			if got != testCase.want {
				t.Fatalf("testConcurrencyMultiplicator() = %d, want %d", got, testCase.want)
			}
		})
	}
}
