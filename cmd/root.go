package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hashDataCmd)
	rootCmd.AddCommand(signCertificateCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(benchmarkCmd)
	rootCmd.AddCommand(newStressTestCommand())
	rootCmd.AddCommand(encryptDataCmd)
	rootCmd.AddCommand(decryptDataCmd)
	rootCmd.AddCommand(signDataCmd)
	rootCmd.AddCommand(verifyDataCmd)
	rootCmd.AddCommand(fakeEndpointCmd)
	rootCmd.AddCommand(versionCmd)
}

var rootCmd = &cobra.Command{
	Use:   "go-client-cli",
	Short: "CLI for working with Crypto Broker",
}

func Execute() {
	cmd, err := rootCmd.ExecuteC()
	if err != nil && cmd.Name() == "stress-test" {
		os.Exit(1)
	}
}
