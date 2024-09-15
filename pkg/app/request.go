package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/heartandu/easyrpc/pkg/editor"
	"github.com/heartandu/easyrpc/pkg/format"
	"github.com/heartandu/easyrpc/pkg/fqn"
	"github.com/heartandu/easyrpc/pkg/usecase"
)

func (a *App) registerRequestCmd() {
	cmd := &cobra.Command{
		Use:               "request [method]",
		Aliases:           []string{"r"},
		Short:             "Prepare a request for a method",
		ValidArgsFunction: a.methodAutocomplete,
		RunE:              a.runRequest,
	}

	cmd.Flags().BoolP("edit", "e", false, "edit the request before printing")
	cmd.Flags().StringP("output", "o", "", "output file to write to")

	a.cmd.AddCommand(cmd)
}

func (a *App) runRequest(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return ErrMissingArgs
	}

	ctx := context.Background()

	cc, err := a.clientConn()
	if err != nil {
		return fmt.Errorf("failed to create client connection: %w", err)
	}

	ds, err := a.descriptorSource(ctx, cc)
	if err != nil {
		return fmt.Errorf("failed to create descriptor source: %w", err)
	}

	var e editor.Editor

	useEditor, err := cmd.Flags().GetBool("edit")
	if err != nil {
		return fmt.Errorf("failed to get edit flag: %w", err)
	}

	if useEditor {
		e = a.editor()
	}

	out := cmd.OutOrStdout()

	outFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("failed to get output flag: %w", err)
	}

	if outFile != "" {
		f, err := a.fs.Create(outFile) //nolint:govet // Handled error variable shadowing is not that bad.
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer f.Close()

		out = f
	}

	mf := format.JSONMessageFormatter(protojson.MarshalOptions{Multiline: true, EmitUnpopulated: true})
	request := usecase.NewRequest(out, e, a.fs, ds, mf)

	err = request.Prepare(fqn.FullyQualifiedMethodName(args[0], a.cfg.Request.Package, a.cfg.Request.Service))
	if err != nil {
		return fmt.Errorf("request print failed: %w", err)
	}

	return nil
}
