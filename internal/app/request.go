package app

import (
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/autocomplete"
	"github.com/heartandu/easyrpc/internal/cmds"
	"github.com/heartandu/easyrpc/internal/flags"
)

func (a *App) registerRequestCmd() {
	requestCmd := cmds.NewRequest(a.fs, &a.cfg)
	methodArgComp := autocomplete.NewMethodArg(a.fs, &a.cfg)

	cmd := &cobra.Command{
		Use:               "request [method]",
		Aliases:           []string{"r"},
		Short:             "Prepare a request for a method",
		ValidArgsFunction: a.loadedConfigCompletion(methodArgComp.Complete),
		RunE:              requestCmd.Run,
	}

	flags.RegisterEditFlag(cmd)
	flags.RegisterOutputFlag(cmd)

	a.cmd.AddCommand(cmd)
}
