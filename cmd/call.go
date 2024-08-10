package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// callCmd represents the call command.
var callCmd = &cobra.Command{
	Use:   "call",
	Short: "Call a remote RPC",
	Run: func(_ *cobra.Command, _ []string) {
		log.Println("call called")
	},
}

func init() {
	rootCmd.AddCommand(callCmd)
}
