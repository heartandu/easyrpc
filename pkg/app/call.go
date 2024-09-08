package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/format"
	"github.com/heartandu/easyrpc/pkg/fqn"
	"github.com/heartandu/easyrpc/pkg/tlsconf"
	"github.com/heartandu/easyrpc/pkg/usecase"
)

func (a *App) registerCallCmd() {
	cmd := &cobra.Command{
		Use:     "call [method name]",
		Aliases: []string{"c"},
		Short:   "Call a remote RPC",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return ErrMissingArgs
			}

			input, err := handleDataFlag(cmd)
			if err != nil {
				return fmt.Errorf("failed to handle data flag: %w", err)
			}

			ctx := context.Background()

			creds, err := a.transportCredentials()
			if err != nil {
				return fmt.Errorf("failed to get transport credentials: %w", err)
			}

			clientConn, err := grpc.NewClient(
				a.cfg.Server.Address,
				grpc.WithTransportCredentials(creds),
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

			callCase := usecase.NewCall(a.cmd.OutOrStdout(), descSrc, clientConn, rp, rf)
			err = callCase.MakeRPCCall(
				context.Background(),
				fqn.FullyQualifiedMethodName(args[0], a.cfg.Request.Package, a.cfg.Request.Service),
				input,
			)
			if err != nil {
				return fmt.Errorf("call rpc failed: %w", err)
			}

			return nil
		},
	}

	registerDataFlag(cmd)

	a.cmd.AddCommand(cmd)
}

func (a *App) transportCredentials() (credentials.TransportCredentials, error) {
	if a.cfg.Server.TLS {
		conf, err := tlsconf.Config(a.cfg.Server.CACert, a.cfg.Server.Cert, a.cfg.Server.CertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to make tls config: %w", err)
		}

		return credentials.NewTLS(conf), nil
	}

	return insecure.NewCredentials(), nil
}
