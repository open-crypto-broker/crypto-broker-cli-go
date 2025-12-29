package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hashCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(benchmarkCmd)
}

var rootCmd = &cobra.Command{
	Use:   "go-client-cli",
	Short: "github.com/open-crypto-broker/crypto-broker-CLI-go for working with Crypto Broker",
}

func Execute() {
	rootCmd.Execute()
}
