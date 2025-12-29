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
	healthCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health checks the broker server status.",
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

		// Ensure tracer provider is shut down when command completes
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown tracer provider", slog.String("error", err.Error()))
			}
			os.Exit(0)
		}()

		healthCommand, err := command.NewHealth(cmd.Context(), logger, tracerProvider)
		if err != nil {
			log.Fatalf("Failed to initialize health command: %v", err)
		}

		if err := healthCommand.Run(cmd.Context(), flags.Loop); err != nil {
			log.Fatalf("Failed to run health command: %v", err)
		}
	},
}
