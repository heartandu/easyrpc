package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.1"

func (a *App) registerVersionCmd() {
	a.cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print application version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintf(os.Stderr, "easyrpc %s\n", version)
		},
	})
}
