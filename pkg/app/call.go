package app

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/pkg/usecase"
)

func (a *App) registerCallCmd() {
	cmd := &cobra.Command{
		Use:   "call [method name]",
		Short: "Call a remote RPC",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return ErrMissingArgs
			}

			data, err := cmd.Flags().GetString("data")
			if err != nil {
				return fmt.Errorf("failed to get data flag: %w", err)
			}

			cc := usecase.NewCall(os.Stdout)
			if err := cc.MakeRPCCall(context.Background(), &a.cfg, args[0], bytes.NewReader([]byte(data))); err != nil {
				return fmt.Errorf("call rpc failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("data", "d", "", "request data in json format")

	a.cmd.AddCommand(cmd)
}
