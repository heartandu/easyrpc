package autocomplete

import (
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/pkg/fs"
)

// ProtoFileFlag represents a proto-file flag autocompletion functionality.
type ProtoFileFlag struct {
	cfgFunc func() (config.Config, error)
}

// NewProtoFileFlag returns a new instance of ProtoFileFlag.
func NewProtoFileFlag(cfgFunc func() (config.Config, error)) *ProtoFileFlag {
	return &ProtoFileFlag{
		cfgFunc: cfgFunc,
	}
}

// Complete provides autocomplete suggestions for proto files location.
// If import-path flag or configuration is provided, autocomplete will filter results by these directories.
func (f *ProtoFileFlag) Complete(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := f.cfgFunc()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	if len(cfg.Proto.ImportPaths) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	result := make([]string, 0, len(cfg.Proto.ImportPaths))

	for _, path := range cfg.Proto.ImportPaths {
		expandedPath, err := fs.ExpandHome(path)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		result = append(result, expandedPath)
	}

	return result, cobra.ShellCompDirectiveFilterDirs
}
