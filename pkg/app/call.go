package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/format"
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

			input, err := handleDataFlag(cmd)
			if err != nil {
				return fmt.Errorf("failed to handle data flag: %w", err)
			}

			ctx := context.Background()

			clientConn, err := grpc.NewClient(
				a.cfg.Server.Address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("failed to create grpc client connection: %w", err)
			}

			var descSrc descriptor.Source
			if a.cfg.Server.Reflection {
				descSrc, err = descriptor.ReflectionSource(ctx, clientConn)
			} else {
				descSrc, err = descriptor.ProtoFilesSource(ctx, a.cfg.Proto.ImportPaths, a.cfg.Proto.ProtoFiles)
			}

			if err != nil {
				return fmt.Errorf("failed to create descriptor source: %w", err)
			}

			rp := format.JSONRequestParser(input, protojson.UnmarshalOptions{})
			rf := format.JSONResponseFormatter(protojson.MarshalOptions{Multiline: true})

			callCase := usecase.NewCall(os.Stdout, descSrc, clientConn, rp, rf)
			if err := callCase.MakeRPCCall(context.Background(), args[0], input); err != nil {
				return fmt.Errorf("call rpc failed: %w", err)
			}

			return nil
		},
	}

	registerDataFlag(cmd)

	a.cmd.AddCommand(cmd)
}
