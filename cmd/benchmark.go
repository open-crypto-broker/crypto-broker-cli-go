package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/flags"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"github.com/spf13/cobra"
)

func init() {
	benchmarkCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark runs server-side cryptographic benchmarks.",
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := flags.ValidateFlagLoop(flags.Loop); err != nil {
			log.Fatalf("Invalid loop flag value: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "CLIENT: ", log.Ldate|log.Lmicroseconds)

		ctx := cmd.Context()
		tracerProvider, err := otel.NewTracerProvider(ctx, "crypto-broker-cli-go", "0.0.0")
		if err != nil {
			log.Fatalf("Failed to initialize tracer provider: %v", err)
		}

		// Shutdown function that ensures proper cleanup
		shutdownTracer := func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				log.Printf("Warning: Failed to shutdown tracer provider: %v", err)
			}
		}
		defer shutdownTracer()

		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			shutdownTracer()
			log.Fatalf("Failed to initialize library: %v", err)
		}

		benchmarkCommand, err := command.NewBenchmark(ctx, lib, logger, tracerProvider)
		if err != nil {
			shutdownTracer()
			log.Fatalf("Failed to initialize benchmark command: %v", err)
		}

		if err := benchmarkCommand.Run(ctx, flags.Loop); err != nil {
			shutdownTracer()
			log.Fatalf("Failed to run benchmark command: %v", err)
		}
	},
}
