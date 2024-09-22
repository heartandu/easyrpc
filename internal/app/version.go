package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) registerVersionCmd() {
	a.cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print application version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintf(a.cmd.OutOrStdout(), "easyrpc %s\n", a.version)
		},
	})
}
