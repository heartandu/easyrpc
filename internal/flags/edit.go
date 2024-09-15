package flags

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/pkg/editor"
)

// RegisterEditFlag registers the edit flag for a given command.
// The flag allows the user edit input with a text editor of choice.
func RegisterEditFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("edit", "e", false, "edit the request before printing")
}

// HandleEditFlag returns an editor.Editor instance if the edit flag is set. Otherwise it returns nil.
// The actual editor used is determined by configuration or, if not set, nano is used by default.
func HandleEditFlag(cmd *cobra.Command, fs afero.Fs, cfg *config.Config) (editor.Editor, error) {
	var e editor.Editor

	useEditor, err := cmd.Flags().GetBool("edit")
	if err != nil {
		return nil, fmt.Errorf("failed to get edit flag: %w", err)
	}

	if !useEditor {
		return e, nil
	}

	command := "nano"
	if e := cfg.Editor.Cmd; e != "" {
		command = e
	}

	return editor.NewCmdEditor(fs, command), nil
}
