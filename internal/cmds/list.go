package cmds

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/client"
	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/internal/proto"
	"github.com/heartandu/easyrpc/pkg/usecase"
)

// List represents a command to list available RPC methods.
type List struct {
	fs  afero.Fs
	cfg *config.Config
}

// NewList creates a new List instance.
func NewList(fs afero.Fs, cfg *config.Config) *List {
	return &List{
		fs:  fs,
		cfg: cfg,
	}
}

// Run executes the List command.
func (l *List) Run(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("%w: %v", ErrUnexpectedArgs, args)
	}

	validator := config.NewValidator(config.WithValidateConn(l.cfg.Server.Reflection))
	if err := validator.Validate(l.cfg); err != nil {
		return errors.Join(ErrValidation, err)
	}

	ctx := context.Background()

	cc, err := client.New(l.fs, l.cfg)
	if err != nil {
		return fmt.Errorf("failed to create client connection: %w", err)
	}

	ds, err := proto.NewDescriptorSource(ctx, l.fs, l.cfg, cc)
	if err != nil {
		return fmt.Errorf("failed to create descriptor source: %w", err)
	}

	list := usecase.NewList(cmd.OutOrStdout(), ds)
	if err := list.Run(l.cfg.Request.Package, l.cfg.Request.Service); err != nil {
		return fmt.Errorf("failed to list methods: %w", err)
	}

	return nil
}
