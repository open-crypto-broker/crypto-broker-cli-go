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
	signCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	signCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
	signCmd.Flags().StringVarP(&flags.Encoding, constant.KeywordFlagEncoding, "", constant.EncodingPEM,
		fmt.Sprintf("Specify encoding to be used (%s, %s)", constant.EncodingPEM, constant.EncodingB64))
	signCmd.Flags().StringVarP(&flags.Subject, constant.KeywordFlagSubject, "", "", "Specify custom subject to be used for certificate generation")
	signCmd.Flags().StringVarP(&flags.FilePathCSR, constant.KeywordFlagFilePathCSR, "", "", "Specify relative path to CSR file")
	signCmd.Flags().StringVarP(&flags.FilePathCACert, constant.KeywordFlagFilePathCACert, "", "", "Specify relative path to CA certificate file")
	signCmd.Flags().StringVarP(&flags.FilePathSigningKey, constant.KeywordFlagFilePathSigningKey, "", "", "Specify relative path to signing key file")

	signCmd.MarkFlagRequired(constant.KeywordFlagFilePathCSR)
	signCmd.MarkFlagRequired(constant.KeywordFlagFilePathCACert)
	signCmd.MarkFlagRequired(constant.KeywordFlagFilePathSigningKey)
	signCmd.MarkFlagsRequiredTogether(constant.KeywordFlagFilePathCSR, constant.KeywordFlagFilePathCACert, constant.KeywordFlagFilePathSigningKey)
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign sends certificate signing request to crypto broker.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := flags.ValidateFlagEncoding(flags.Encoding); err != nil {
			log.Fatalf("Invalid encoding flag value: %v", err)
		}

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

		signCommand, err := command.NewSign(cmd.Context(), logger, tracerProvider)
		if err != nil {
			log.Fatalf("Failed to initialize sign command: %v", err)
		}

		if err := signCommand.Run(cmd.Context(),
			flags.FilePathCSR, flags.FilePathCACert, flags.FilePathSigningKey, flags.Profile, flags.Encoding, flags.Subject, flags.Loop); err != nil {
			log.Fatalf("Failed to run sign command: %v", err)
		}
	},
}
