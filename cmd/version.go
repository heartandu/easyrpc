package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.1"

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print application version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Fprintf(os.Stderr, "easyrpc %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
