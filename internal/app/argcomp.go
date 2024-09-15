package app

import "github.com/spf13/cobra"

type argsFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// loadedConfigCompletion wraps an argument completion function,
// and reads configuration before passing control to the wrapped function.
func (a *App) loadedConfigCompletion(f argsFunc) argsFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		a.readConfig()

		return f(cmd, args, toComplete)
	}
}
