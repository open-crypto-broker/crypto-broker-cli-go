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
	fakeEndpointCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "",
		constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var fakeEndpointCmd = &cobra.Command{
	Use:   "fake-endpoint",
	Short: "Fake endpoint sends fake endpoint request to crypto broker.",
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

		tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
		if err != nil {
			logger.Error("Failed to initialize tracer provider", "error", err)
			panic(err)
		}

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

		fakeEndpointCommand, err := command.NewFakeEndpoint(ctx, lib, logger, tracerProvider)
		if err != nil {
			shutdownTracer()
			logger.Error("Failed to initialize fake endpoint command", "error", err)
			panic(err)
		}

		if err := fakeEndpointCommand.Run(ctx, flags.Loop); err != nil {
			shutdownTracer()
			logger.Error("Failed to run fake endpoint command", "error", err)
			panic(err)
		}
	},
}
