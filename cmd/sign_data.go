package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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
	signDataCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	signDataCmd.Flags().StringVarP(&flags.FilePathKey, constant.KeywordFlagFilePathKey, "", "", "Path to the PEM private key used for signing")
	signDataCmd.Flags().StringVarP(&flags.SignatureFormat, constant.KeywordFlagSignatureFormat, "", constant.EncodingRAW, "Signature format: raw, der, or pem")
	signDataCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue, fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
	if err := errors.Join(signDataCmd.MarkFlagRequired(constant.KeywordFlagFilePathKey)); err != nil {
		panic(err)
	}
}

var signDataCmd = &cobra.Command{
	Use: "sign-data input", Short: "Sign data through Crypto Broker.", Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if !validateSignatureFormat(flags.SignatureFormat) {
			panic(fmt.Errorf("unsupported signature format %q", flags.SignatureFormat))
		}

		if err := flags.ValidateFlagLoop(flags.Loop); err != nil {
			slog.Error("Invalid loop flag value", "error", err)

			panic(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := clog.SetupGlobalLogger(ctx)
		tracerProvider, err := otel.NewTracerProvider(ctx, logger)
		if err != nil {
			panic(err)
		}

		defer func() { shutdownTracerProvider(logger, tracerProvider) }()
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			panic(err)
		}

		command, err := command.NewSignData(ctx, lib, logger, tracerProvider)
		if err != nil {
			panic(err)
		}

		if err := command.Run(ctx, []byte(args[0]), flags.Profile, flags.FilePathKey, flags.SignatureFormat, flags.Loop); err != nil {
			panic(err)
		}
	},
}

// validateSignatureFormat checks whether the provided signature format is valid
func validateSignatureFormat(format string) bool {
	switch strings.ToLower(format) {
	case constant.EncodingRAW, constant.EncodingDER, constant.EncodingPEM:
		return true
	default:
		return false
	}
}

func shutdownTracerProvider(logger *slog.Logger, tracerProvider *otel.TracerProvider) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
		logger.Warn("Failed to shutdown tracer provider", "error", err)
	}
}
