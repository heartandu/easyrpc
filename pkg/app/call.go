package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/pkg/usecase"
)

func (a *App) registerCallCmd() {
	a.cmd.AddCommand(&cobra.Command{
		Use:   "call [method name]",
		Short: "Call a remote RPC",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return ErrMissingArgs
			}

			cc := usecase.NewCall(os.Stdout)
			if err := cc.MakeRPCCall(context.Background(), &a.cfg, args[0]); err != nil {
				return fmt.Errorf("call rpc failed: %w", err)
			}

			return nil
		},
	})
}
