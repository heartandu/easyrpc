package app

import (
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/cmds"
)

func (a *App) registerListCmd() {
	listCmd := cmds.NewList(a.fs, &a.cfg)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available RPCs",
		RunE:    listCmd.Run,
	}

	a.cmd.AddCommand(cmd)
}
