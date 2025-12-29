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
	otelio "go.opentelemetry.io/otel"

	"github.com/spf13/cobra"
)

func init() {
	hashCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	hashCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var hashCmd = &cobra.Command{
	Use:   "hash SLICE_OF_BYTES_TO_BE_HASHED",
	Short: "Hash sends hashing request to crypto broker.",
	Args:  cobra.ExactArgs(1),
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

		// Debug: Log propagator configuration
		if propagator := otelio.GetTextMapPropagator(); propagator != nil {
			log.Printf("CLI DEBUG: TextMapPropagator type: %T", propagator)
		} else {
			log.Printf("CLI DEBUG: No TextMapPropagator configured")
		}

		// Ensure tracer provider is shut down when command completes
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown tracer provider", slog.String("error", err.Error()))
			}
		}()

		// Handle graceful shutdown on signals
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			slog.Info("Received signal, shutting down tracer provider")
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown tracer provider", slog.String("error", err.Error()))
			}
			os.Exit(0)
		}()

		hashCommand, err := command.NewHash(ctx, logger, tracerProvider)
		if err != nil {
			log.Fatalf("Failed to initialize hash command: %v", err)
		}

		if err := hashCommand.Run(ctx, []byte(args[0]), flags.Profile, flags.Loop); err != nil {
			log.Fatalf("Failed to run hash command: %v", err)
		}
	},
}
