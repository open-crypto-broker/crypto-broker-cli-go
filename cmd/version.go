package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/command"
)

// gitSHA and gitTag are set at build time using ldflags
var (
	gitSHA = "unknown"
	gitTag = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of the CLI and its Go client library.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		versionCmd, err := command.NewVersion()
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "failed to initialize version command: %v\n", err)

			return
		}

		out, err := versionCmd.Run(gitTag, gitSHA)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "failed to compute version payload: %v\n", err)

			return
		}

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "failed to marshal version JSON: %v\n", err)

			return
		}

		fmt.Println(string(b))
	},
}
