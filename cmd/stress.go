package cmd

import (
	"log/slog"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/clog"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/spf13/cobra"
)

var (
	stressConfig = command.DefaultStressConfig()
	stressInput  = string(stressConfig.Input)
)

func init() {
	stressCmd.Flags().IntVar(&stressConfig.Connections, constant.KeywordFlagConnections, stressConfig.Connections, "Number of gRPC client connections to open")
	stressCmd.Flags().IntVar(&stressConfig.Concurrency, constant.KeywordFlagConcurrency, stressConfig.Concurrency, "Number of concurrent in-flight hash requests")
	stressCmd.Flags().Uint64Var(&stressConfig.Requests, constant.KeywordFlagRequests, stressConfig.Requests, "Total requests to send before stopping; 0 means run until duration expires")
	stressCmd.Flags().DurationVar(&stressConfig.Duration, constant.KeywordFlagDuration, stressConfig.Duration, "Maximum stress test duration; set requests for a request-bounded run")
	stressCmd.Flags().DurationVar(&stressConfig.Timeout, constant.KeywordFlagTimeout, stressConfig.Timeout, "Per-request timeout")
	stressCmd.Flags().DurationVar(&stressConfig.ConnectTimeout, constant.KeywordFlagConnectTimeout, stressConfig.ConnectTimeout, "Timeout for each client connection attempt")
	stressCmd.Flags().StringVar(&stressConfig.Profile, constant.KeywordFlagProfile, stressConfig.Profile, "Cryptographic profile to use for hash requests")
	stressCmd.Flags().StringVar(&stressInput, constant.KeywordFlagInput, stressInput, "Input payload to hash on every request")
}

var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Stress test the crypto broker server with concurrent hash requests.",
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		stressConfig.Input = []byte(stressInput)
		if err := stressConfig.Validate(); err != nil {
			slog.Error("Invalid stress test configuration", "error", err)
			panic(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := clog.SetupGlobalLogger(ctx)

		stressCommand, err := command.NewStress(logger)
		if err != nil {
			logger.Error("Failed to initialize stress command", "error", err)
			panic(err)
		}

		if err := stressCommand.Run(ctx, stressConfig); err != nil {
			logger.Error("Failed to run stress command", "error", err)
			panic(err)
		}
	},
}
