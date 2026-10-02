package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/clog"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/flags"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"github.com/spf13/cobra"
)

func init() {
	verifyDataCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	verifyDataCmd.Flags().StringVarP(&flags.FilePathKey, constant.KeywordFlagFilePathKey, "", "", "Path to the PEM public key used for verification")
	verifyDataCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue, fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
	if err := errors.Join(verifyDataCmd.MarkFlagRequired(constant.KeywordFlagFilePathKey)); err != nil {
		panic(err)
	}
}

var verifyDataCmd = &cobra.Command{
	Use: "verify-data input hex-signature", Short: "Verify data through Crypto Broker.", Args: cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
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

		command, err := command.NewVerifyData(ctx, lib, logger, tracerProvider)
		if err != nil {
			panic(err)
		}

		signature, err := hex.DecodeString(args[1])
		if err != nil {
			panic(fmt.Errorf("signature must be hexadecimal: %w", err))
		}

		if err := command.Run(ctx, []byte(args[0]), signature, flags.Profile, flags.FilePathKey, flags.Loop); err != nil {
			panic(err)
		}
	},
}
