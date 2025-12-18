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
	benchmarkCmd.Flags().IntVarP(&flags.Loop, constant.KeywordFlagLoop, "", constant.NoLoopFlagValue,
		fmt.Sprintf("Specify delay for loop in milliseconds (%d-%d)", constant.MinLoopFlagValue, constant.MaxLoopFlagValue))
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark runs server-side cryptographic benchmarks.",
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := flags.ValidateFlagLoop(flags.Loop); err != nil {
			log.Fatalf("Invalid loop flag value: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "CLIENT: ", log.Ldate|log.Lmicroseconds)

		benchmarkCommand, err := command.NewBenchmark(cmd.Context(), logger)
		if err != nil {
			log.Fatalf("Failed to initialize benchmark command: %v", err)
		}

		if err := benchmarkCommand.Run(cmd.Context(), flags.Loop); err != nil {
			log.Fatalf("Failed to run benchmark command: %v", err)
		}
	},
}
