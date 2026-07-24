package cmd

import (
	"context"
	"encoding/hex"
	"errors"
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
	decryptDataCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	decryptDataCmd.Flags().StringVarP(&flags.KeySource, constant.KeywordFlagKeySource, "", constant.KeySourceRaw, "Specify key source: raw or key-id")
	decryptDataCmd.Flags().StringVarP(&flags.Key, constant.KeywordFlagKey, "", "", "Specify raw key or managed key ID")
	decryptDataCmd.Flags().StringVarP(&flags.Nonce, constant.KeywordFlagNonce, "", "", "Specify AES-GCM nonce as hexadecimal")
	decryptDataCmd.Flags().StringVarP(&flags.AAD, constant.KeywordFlagAAD, "", "", "Specify additional authenticated data as hexadecimal")
	decryptDataCmd.Flags().StringVarP(&flags.Tag, constant.KeywordFlagTag, "", "", "Specify AES-GCM authentication tag as hexadecimal")
	decryptDataCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
	err := errors.Join(
		decryptDataCmd.MarkFlagRequired(constant.KeywordFlagKey),
		decryptDataCmd.MarkFlagRequired(constant.KeywordFlagNonce),
		decryptDataCmd.MarkFlagRequired(constant.KeywordFlagTag),
	)
	if err != nil {
		panic(err)
	}
}

var decryptDataCmd = &cobra.Command{
	Use:   "decrypt-data SLICE_OF_BYTES_TO_BE_DECRYPTED",
	Short: "Decrypt data through Crypto Broker.",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := flags.ValidateFlagKeySource(flags.KeySource); err != nil {
			slog.Error("Invalid key source flag value", "error", err)
			panic(err)
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

		decryptDataCommand, err := command.NewDecryptData(ctx, lib, logger, tracerProvider)
		if err != nil {
			shutdownTracer()
			logger.Error("Failed to initialize decrypt-data command", "error", err)
			panic(err)
		}

		ciphertext, err := hex.DecodeString(args[0])
		if err != nil {
			logger.Error("Ciphertext must be hexadecimal", "error", err)
			panic(err)
		}
		if err := decryptDataCommand.Run(ctx, ciphertext, flags.Profile, flags.KeySource, flags.Key, flags.Nonce, flags.AAD, flags.Tag, flags.Loop); err != nil {
			shutdownTracer()
			logger.Error("Failed to run decrypt-data command", "error", err)
			panic(err)
		}
	},
}
