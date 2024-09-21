package autocomplete

import (
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/config"
)

// ProtoFileArg represents a proto-file flag autocompletion functionality.
type ProtoFileArg struct {
	cfg *config.Config
}

// NewProtoFileArg returns a new instance of ProtoFileArg.
func NewProtoFileArg(cfg *config.Config) *ProtoFileArg {
	return &ProtoFileArg{
		cfg: cfg,
	}
}

// Complete provides autocomplete suggestions for proto files location.
// If import-path flag or configuration is provided, autocomplete will filter results by these directories.
func (a *ProtoFileArg) Complete(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(a.cfg.Proto.ImportPaths) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	return a.cfg.Proto.ImportPaths, cobra.ShellCompDirectiveFilterDirs
}
