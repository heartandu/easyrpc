package cmds

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/internal/client"
	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/internal/flags"
	"github.com/heartandu/easyrpc/internal/proto"
	"github.com/heartandu/easyrpc/pkg/format"
	"github.com/heartandu/easyrpc/pkg/fqn"
	"github.com/heartandu/easyrpc/pkg/usecase"
)

// Request represents a command to build formatted request string or file.
type Request struct {
	fs  afero.Fs
	cfg *config.Config
}

// NewRequest creates a new Request instance.
func NewRequest(fs afero.Fs, cfg *config.Config) *Request {
	return &Request{
		fs:  fs,
		cfg: cfg,
	}
}

// Run executes the Request command.
func (r *Request) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return ErrMissingArgs
	}

	if err := r.validateConfig(); err != nil {
		return errors.Join(ErrValidation, err)
	}

	ctx := context.Background()

	cc, err := client.New(r.fs, r.cfg)
	if err != nil {
		return fmt.Errorf("failed to create client connection: %w", err)
	}

	ds, err := proto.NewDescriptorSource(ctx, r.fs, r.cfg, cc)
	if err != nil {
		return fmt.Errorf("failed to create descriptor source: %w", err)
	}

	e, err := flags.HandleEditFlag(cmd, r.fs, r.cfg)
	if err != nil {
		return fmt.Errorf("failed to handle edit flag: %w", err)
	}

	out, err := flags.HandleOutputFlag(cmd, r.fs)
	if err != nil {
		return fmt.Errorf("failed to handle output flag: %w", err)
	}
	defer out.Close()

	mf := format.JSONMessageFormatter(protojson.MarshalOptions{Multiline: true, EmitUnpopulated: true})
	request := usecase.NewRequest(out, e, r.fs, ds, mf)

	err = request.Prepare(fqn.FullyQualifiedMethodName(args[0], r.cfg.Request.Package, r.cfg.Request.Service))
	if err != nil {
		return fmt.Errorf("request preparing failed: %w", err)
	}

	return nil
}

func (r *Request) validateConfig() error {
	var err error

	if len(r.cfg.Proto.ProtoFiles) == 0 && !r.cfg.Server.Reflection {
		err = errors.Join(err, ErrNoSource)
	}

	if r.cfg.Server.Reflection {
		if r.cfg.Server.Address == "" {
			err = errors.Join(err, ErrEmptyAddress)
		}

		if r.cfg.TLS.Cert == "" && r.cfg.TLS.Key != "" || r.cfg.TLS.Cert != "" && r.cfg.TLS.Key == "" {
			err = errors.Join(err, ErrMissingCertOrKey)
		}
	}

	return err
}
