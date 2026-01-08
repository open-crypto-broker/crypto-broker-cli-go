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

		healthCommand, err := command.NewHealth(ctx, lib, logger, tracerProvider)
		if err != nil {
			shutdownTracer()
			log.Fatalf("Failed to initialize health command: %v", err)
		}

		if err := healthCommand.Run(ctx, flags.Loop); err != nil {
			shutdownTracer()
			log.Fatalf("Failed to run health command: %v", err)
		}
	},
}
