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
	hashCmd.Flags().StringVarP(&flags.Profile, constant.KeywordFlagProfile, "", "Default", "Specify profile to be used")
	hashCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in miliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var hashCmd = &cobra.Command{
	Use:   "hash SLICE_OF_BYTES_TO_BE_HASHED",
	Short: "Hash sends hashing request to crypto broker.",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return flags.ValidateFlagLoop(flags.Loop)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.New(os.Stdout, "CLIENT: ", log.Ldate|log.Lmicroseconds)

		hashCommand, err := command.NewHash(cmd.Context(), logger)
		if err != nil {
			return err
		}

		return hashCommand.Run(cmd.Context(), []byte(args[0]), flags.Profile, flags.Loop)
	},
}
