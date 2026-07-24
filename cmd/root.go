package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hashDataCmd)
	rootCmd.AddCommand(signCertificateCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(benchmarkCmd)
	rootCmd.AddCommand(encryptDataCmd)
	rootCmd.AddCommand(decryptDataCmd)
	rootCmd.AddCommand(fakeEndpointCmd)
	rootCmd.AddCommand(versionCmd)
}

var rootCmd = &cobra.Command{
	Use:   "go-client-cli",
	Short: "CLI for working with Crypto Broker",
}

func Execute() {
	_ = rootCmd.Execute()
}
