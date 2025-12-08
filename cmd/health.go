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
	healthCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in miliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health checks the broker server status.",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return flags.ValidateFlagLoop(flags.Loop)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(os.Stdout, "CLIENT: ", log.Ldate|log.Lmicroseconds)

		healthCommand, err := command.NewHealth(cmd.Context(), logger)
		if err != nil {
			return err
		}

		return healthCommand.Run(cmd.Context(), flags.Loop)
	},
}
