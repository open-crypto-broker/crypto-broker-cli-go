package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/flags"

	"github.com/spf13/cobra"
)

func init() {
	signCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	signCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in miliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := flags.ValidateFlagEncoding(flags.Encoding); err != nil {
			return err
		}

		return flags.ValidateFlagLoop(flags.Loop)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(os.Stdout, "CLIENT: ", log.Ldate|log.Lmicroseconds)
		signCommand, err := command.NewSign(cmd.Context(), logger)
		if err != nil {
			return err
		}

		return signCommand.Run(cmd.Context(),
			flags.FilePathCSR, flags.FilePathCACert, flags.FilePathSigningKey, flags.Profile, flags.Encoding, flags.Subject, flags.Loop)
	},
}
