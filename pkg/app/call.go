package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/format"
	"github.com/heartandu/easyrpc/pkg/fqn"
	"github.com/heartandu/easyrpc/pkg/tlsconf"
	"github.com/heartandu/easyrpc/pkg/usecase"
)

func (a *App) registerCallCmd() {
	cmd := &cobra.Command{
		Use:               "call [method name]",
		Aliases:           []string{"c"},
		Short:             "Call a remote RPC",
		ValidArgsFunction: a.callAutocomplete,
		RunE:              a.runCall,
	}

	// TODO: consider using - instead of @ for stdin.
	registerDataFlag(cmd)

	a.cmd.AddCommand(cmd)
}

func (a *App) callAutocomplete(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	a.readConfig()

	ctx := context.Background()

	clientConn, err := a.clientConn()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	descSrc, err := a.descriptorSource(ctx, clientConn)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	methods, err := descSrc.ListMethods()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	result := make([]string, 0)

	isCaseInsensitive := toComplete == strings.ToLower(toComplete)

	completionToCompare := toComplete
	if isCaseInsensitive {
		completionToCompare = strings.ToLower(completionToCompare)
	}

	for _, method := range methods {
		methodToCompare := method
		if isCaseInsensitive {
			methodToCompare = strings.ToLower(methodToCompare)
		}

		if strings.Contains(methodToCompare, completionToCompare) {
			result = append(result, method)
		}
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}

func (a *App) runCall(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return ErrMissingArgs
	}

	input, err := handleDataFlag(a.fs, cmd)
	if err != nil {
		return fmt.Errorf("failed to handle data flag: %w", err)
	}

	ctx := context.Background()

	clientConn, err := a.clientConn()
	if err != nil {
		return fmt.Errorf("failed to create grpc client connection: %w", err)
	}

	descSrc, err := a.descriptorSource(ctx, clientConn)
	if err != nil {
		return fmt.Errorf("failed to create descriptor source: %w", err)
	}

	rp := format.JSONRequestParser(input, protojson.UnmarshalOptions{})
	rf := format.JSONResponseFormatter(protojson.MarshalOptions{Multiline: true})

	callCase := usecase.NewCall(a.cmd.OutOrStdout(), descSrc, clientConn, rp, rf, metadata.New(a.cfg.Request.Metadata))

	err = callCase.MakeRPCCall(
		context.Background(),
		fqn.FullyQualifiedMethodName(args[0], a.cfg.Request.Package, a.cfg.Request.Service),
		input,
	)
	if err != nil {
		return fmt.Errorf("call rpc failed: %w", err)
	}

	return nil
}

func (a *App) clientConn() (*grpc.ClientConn, error) {
	creds, err := a.transportCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get transport credentials: %w", err)
	}

	clientConn, err := grpc.NewClient(
		a.cfg.Server.Address,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return clientConn, nil
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

func (a *App) descriptorSource(ctx context.Context, clientConn *grpc.ClientConn) (descriptor.Source, error) {
	var (
		descSrc descriptor.Source
		err     error
	)

	if a.cfg.Server.Reflection {
		descSrc, err = descriptor.ReflectionSource(ctx, clientConn)
	} else {
		descSrc, err = descriptor.ProtoFilesSource(ctx, a.cfg.Proto.ImportPaths, a.cfg.Proto.ProtoFiles)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create descriptor source: %w", err)
	}

	return descSrc, nil
}
