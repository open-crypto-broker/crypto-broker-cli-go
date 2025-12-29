package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	"github.com/spf13/cobra"
)

var tracerProvider *otel.TracerProvider

func init() {
	rootCmd.AddCommand(hashCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(benchmarkCmd)

	// Initialize tracing
	ctx := context.Background()
	var err error
	tracerProvider, err = otel.NewTracerProvider(ctx, "", "")
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
}

var rootCmd = &cobra.Command{
	Use:   "go-client-cli",
	Short: "github.com/open-crypto-broker/crypto-broker-CLI-go for working with Crypto Broker",
}

func Execute() {
	rootCmd.Execute()
}
