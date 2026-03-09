package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/clog"
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
			slog.Error("Invalid loop flag value", "error", err)
			panic(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := clog.SetupGlobalLogger(ctx)

		tracerProvider, err := otel.NewTracerProvider(ctx, logger, "crypto-broker-cli-go", "0.0.0")
		if err != nil {
			logger.Error("Failed to initialize tracer provider", "error", err)
			panic(err)
		}

		// Shutdown function that ensures proper cleanup
		shutdownTracer := func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				logger.Warn("Failed to shutdown tracer provider", "error", err)
			}
		}
		defer shutdownTracer()

		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			shutdownTracer()
			logger.Error("Failed to initialize library", "error", err)
			panic(err)
		}

		benchmarkCommand, err := command.NewBenchmark(ctx, lib, logger, tracerProvider)
		if err != nil {
			shutdownTracer()
			logger.Error("Failed to initialize benchmark command", "error", err)
			panic(err)
		}

		if err := benchmarkCommand.Run(ctx, flags.Loop); err != nil {
			shutdownTracer()
			logger.Error("Failed to run benchmark command", "error", err)
			panic(err)
		}
	},
}
