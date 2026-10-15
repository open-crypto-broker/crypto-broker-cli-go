package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
	"github.com/spf13/cobra"
)

func newStressTestCommand() *cobra.Command {
	var concurrent, num, count int
	cmd := &cobra.Command{
		Use:   "stress-test",
		Short: "Run concurrent client-side hash requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if count < 1 {
				return fmt.Errorf("count must be at least 1")
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			for range count {
				var before, after runtime.MemStats
				runtime.ReadMemStats(&before)
				startedAt := time.Now()
				result, err := command.RunStressTest(ctx, concurrent, num)
				elapsed := time.Since(startedAt)
				runtime.ReadMemStats(&after)
				if err != nil {
					return err
				}
				// Measure one wave; benchmark calibration would change COUNT.
				benchmark := testing.BenchmarkResult{
					N: 1, T: elapsed,
					MemAllocs: after.Mallocs - before.Mallocs,
					MemBytes:  after.TotalAlloc - before.TotalAlloc,
					Extra: map[string]float64{
						"connections":         float64(result.Connections),
						"requests/connection": float64(result.RequestsPerConnection),
						"requests":            float64(result.Requests),
						"requests/s":          float64(result.Requests) / elapsed.Seconds(),
					},
				}
				for code, count := range result.StatusCounts {
					benchmark.Extra["grpc_"+code] = float64(count)
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "BenchmarkStressHashConcurrentConnections-%d\t%s\t%s\n", runtime.GOMAXPROCS(0), benchmark.String(), benchmark.MemString()); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&concurrent, "concurrent", 100, "Number of concurrent connections (1-100000)")
	cmd.Flags().IntVar(&num, "num", 100, "Concurrent hash requests per connection (1-1000000)")
	cmd.Flags().IntVar(&count, "count", 1, "Number of runs (at least 1)")
	return cmd
}
