package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/flags"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"

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

		// Initialize tracing
		ctx := cmd.Context()
		tracerProvider, err := otel.NewTracerProvider(ctx, "crypto-broker-cli-go", "0.0.0")
		if err != nil {
			slog.Error("Failed to initialize tracer provider", slog.String("error", err.Error()))
			os.Exit(1)
		}

		// Handle graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			slog.Info("Shutting down tracer provider")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown tracer provider", slog.String("error", err.Error()))
			}
			os.Exit(0)
		}()

		benchmarkCommand, err := command.NewBenchmark(cmd.Context(), logger, tracerProvider)
		if err != nil {
			log.Fatalf("Failed to initialize benchmark command: %v", err)
		}

		if err := benchmarkCommand.Run(cmd.Context(), flags.Loop); err != nil {
			log.Fatalf("Failed to run benchmark command: %v", err)
		}
	},
}
