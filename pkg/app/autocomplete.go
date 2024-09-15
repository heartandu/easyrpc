package app

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
)

func (a *App) methodAutocomplete(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	a.readConfig()

	ctx := context.Background()

	cc, err := a.clientConn()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	descSrc, err := a.descriptorSource(ctx, cc)
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
