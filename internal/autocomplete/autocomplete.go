package autocomplete

import (
	"context"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/client"
	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/internal/proto"
)

// MethodArg represents a method argument autocompletion functionality.
type MethodArg struct {
	fs  afero.Fs
	cfg *config.Config
}

// NewMethodArg creates a new MethodArg instance.
func NewMethodArg(fs afero.Fs, cfg *config.Config) *MethodArg {
	return &MethodArg{
		fs:  fs,
		cfg: cfg,
	}
}

// Complete provides autocomplete suggestions for methods based on the input toComplete.
func (a *MethodArg) Complete(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()

	cc, err := client.New(a.fs, a.cfg)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	descSrc, err := proto.NewDescriptorSource(ctx, a.fs, a.cfg, cc)
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
