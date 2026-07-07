package cmd

import (
	"context"
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
	signCertificateCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	signCertificateCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
	signCertificateCmd.Flags().StringVarP(&flags.Encoding, constant.KeywordFlagEncoding, "", constant.EncodingPEM,
		fmt.Sprintf("Specify encoding to be used (%s, %s)", constant.EncodingPEM, constant.EncodingDER))
	signCertificateCmd.Flags().StringVarP(&flags.Subject, constant.KeywordFlagSubject, "", "", "Specify custom subject to be used for certificate generation")
	signCertificateCmd.Flags().StringVarP(&flags.FilePathCSR, constant.KeywordFlagFilePathCSR, "", "", "Specify relative path to CSR file")
	signCertificateCmd.Flags().StringVarP(&flags.FilePathCACert, constant.KeywordFlagFilePathCACert, "", "", "Specify relative path to CA certificate file")
	signCertificateCmd.Flags().StringVarP(&flags.FilePathSigningKey, constant.KeywordFlagFilePathSigningKey, "", "", "Specify relative path to signing key file")

	var err error
	err = errors.Join(err, signCertificateCmd.MarkFlagRequired(constant.KeywordFlagFilePathCSR))
	err = errors.Join(err, signCertificateCmd.MarkFlagRequired(constant.KeywordFlagFilePathCACert))
	err = errors.Join(err, signCertificateCmd.MarkFlagRequired(constant.KeywordFlagFilePathSigningKey))
	if err != nil {
		panic(err)
	}

	signCertificateCmd.MarkFlagsRequiredTogether(constant.KeywordFlagFilePathCSR, constant.KeywordFlagFilePathCACert, constant.KeywordFlagFilePathSigningKey)
}

var signCertificateCmd = &cobra.Command{
	Use:   "sign-certificate",
	Short: "Sign certificate sends certificate signing request to crypto broker.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := flags.ValidateFlagEncoding(flags.Encoding); err != nil {
			slog.Error("Invalid encoding flag value", "error", err)
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

		signCertificateCommand, err := command.NewSignCertificate(ctx, lib, logger, tracerProvider)
		if err != nil {
			shutdownTracer()
			logger.Error("Failed to initialize sign certificate command", "error", err)
			panic(err)
		}

		if err := signCertificateCommand.Run(ctx,
			flags.FilePathCSR, flags.FilePathCACert, flags.FilePathSigningKey, flags.Profile, flags.Encoding, flags.Subject, flags.Loop); err != nil {
			shutdownTracer()
			logger.Error("Failed to run sign certificate command", "error", err)
			panic(err)
		}
	},
}
