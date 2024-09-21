package app

import (
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/autocomplete"
	"github.com/heartandu/easyrpc/internal/cmds"
	"github.com/heartandu/easyrpc/internal/flags"
)

func (a *App) registerCallCmd() {
	callCmd := cmds.NewCall(a.fs, &a.cfg)
	methodArgComp := autocomplete.NewProtoComp(a.fs, a.viper)

	cmd := &cobra.Command{
		Use:               "call [method]",
		Aliases:           []string{"c"},
		Short:             "Call a remote RPC",
		ValidArgsFunction: methodArgComp.CompleteMethod,
		RunE:              callCmd.Run,
	}

	flags.RegisterDataFlag(cmd)

	a.cmd.AddCommand(cmd)
}
