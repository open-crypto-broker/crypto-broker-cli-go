package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// gitSHA and gitTag are set at build time using ldflags
var (
	gitSHA = "unknown"
	gitTag = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of the CLI.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Git Tag: %s\n", gitTag)
		fmt.Printf("Git SHA: %s\n", gitSHA)
	},
}
