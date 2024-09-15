package cmds

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/internal/client"
	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/internal/flags"
	"github.com/heartandu/easyrpc/internal/proto"
	"github.com/heartandu/easyrpc/pkg/format"
	"github.com/heartandu/easyrpc/pkg/fqn"
	"github.com/heartandu/easyrpc/pkg/usecase"
)

// Call represents a command to make an RPC call.
type Call struct {
	fs  afero.Fs
	cfg *config.Config
}

// NewCall creates a new Call command.
func NewCall(fs afero.Fs, cfg *config.Config) *Call {
	return &Call{
		fs:  fs,
		cfg: cfg,
	}
}

// Run executes the Call command.
func (c *Call) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return ErrMissingArgs
	}

	input, err := flags.HandleDataFlag(cmd, c.fs)
	if err != nil {
		return fmt.Errorf("failed to handle data flag: %w", err)
	}
	defer input.Close()

	ctx := context.Background()

	cc, err := client.New(c.fs, c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create client connection: %w", err)
	}

	descSrc, err := proto.NewDescriptorSource(ctx, c.fs, c.cfg, cc)
	if err != nil {
		return fmt.Errorf("failed to create descriptor source: %w", err)
	}

	mp := format.JSONMessageParser(input, protojson.UnmarshalOptions{})
	mf := format.JSONMessageFormatter(protojson.MarshalOptions{Multiline: true, EmitUnpopulated: true})

	call := usecase.NewCall(cmd.OutOrStdout(), descSrc, cc, mp, mf, metadata.New(c.cfg.Request.Metadata))

	err = call.MakeRPCCall(ctx, fqn.FullyQualifiedMethodName(args[0], c.cfg.Request.Package, c.cfg.Request.Service))
	if err != nil {
		return fmt.Errorf("call rpc failed: %w", err)
	}

	return nil
}
