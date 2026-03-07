package autocomplete

import (
	"context"
	"iter"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/heartandu/easyrpc/internal/client"
	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/internal/proto"
	"github.com/heartandu/easyrpc/pkg/fqn"
)

// ProtoComp represents a protobuf symbol autocompletion functionality.
type ProtoComp struct {
	fs      afero.Fs
	cfgFunc func() (config.Config, error)
}

// NewProtoComp creates a new ProtoComp instance.
func NewProtoComp(fs afero.Fs, cfgFunc func() (config.Config, error)) *ProtoComp {
	return &ProtoComp{
		fs:      fs,
		cfgFunc: cfgFunc,
	}
}

// CompleteMethod provides autocomplete suggestions for methods.
func (c *ProtoComp) CompleteMethod(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := c.cfgFunc()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	methods, err := c.symbols(cmd.Context(), &cfg, toComplete, func(fqmn fqn.FQMN) string {
		b := fqmn.PartsBuilder()

		// If package name has been set, and doesn't match
		// the received one, filter the method out.
		// If they match, omit the package name from the result.
		// Otherwise, if the method contains a package name, use it
		// to form a fully qualified name.
		// WithPackage call also forces service name to be included for correctness.
		if cfg.Request.Package != "" {
			if fqmn.PackageName != cfg.Request.Package {
				return ""
			}
		} else if fqmn.PackageName != "" {
			b.WithPackage()
		}

		// If service name has been set, and doesn't match
		// the received one, filter the method out.
		// If they match, omit the service name from the result.
		// Otherwise, use it to form a fully qualified name.
		if cfg.Request.Service != "" {
			if fqmn.Service != cfg.Request.Service {
				return ""
			}
		} else {
			b.WithService()
		}

		return b.String()
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return methods, cobra.ShellCompDirectiveNoFileComp
}

// CompletePackage provides autocomplete suggestions for package names.
func (c *ProtoComp) CompletePackage(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := c.cfgFunc()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	packages, err := c.symbols(cmd.Context(), &cfg, toComplete, func(fqmn fqn.FQMN) string {
		return fqmn.PackageName
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return packages, cobra.ShellCompDirectiveNoFileComp
}

// CompleteService provides autocomplete suggestions for service names.
func (c *ProtoComp) CompleteService(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := c.cfgFunc()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	services, err := c.symbols(cmd.Context(), &cfg, toComplete, func(fqmn fqn.FQMN) string {
		// If package name has been set, and doesn't match
		// the received one, filter the method out.
		if cfg.Request.Package != "" && fqmn.PackageName != cfg.Request.Package {
			return ""
		}

		return fqmn.Service
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return services, cobra.ShellCompDirectiveNoFileComp
}

func (c *ProtoComp) symbols(
	ctx context.Context,
	cfg *config.Config,
	toComplete string,
	filterMapFunc func(fqmn fqn.FQMN) string,
) ([]string, error) {
	cc, err := client.New(c.fs, cfg)
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in autocomplete.
	}

	descSrc, err := proto.NewDescriptorSource(ctx, c.fs, cfg, cc)
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in autocomplete.
	}

	methods, err := descSrc.ListMethods()
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in autocomplete.
	}

	encounteredSymbols := map[string]struct{}{}
	result := make([]string, 0)

	completionToCompare := strings.ToLower(toComplete)
	for symbol := range filterMapIter(methods, filterMapFunc) {
		symbolToCompare := strings.ToLower(symbol)

		if strings.Contains(symbolToCompare, completionToCompare) {
			if _, ok := encounteredSymbols[symbol]; !ok {
				result = append(result, symbol)
			}

			encounteredSymbols[symbol] = struct{}{}
		}
	}

	return result, nil
}

func filterMapIter(s []string, f func(fqmn fqn.FQMN) string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, s := range s {
			s = f(fqn.ParseFQMN(s))
			if s == "" {
				continue
			}

			if cont := yield(s); !cont {
				return
			}
		}
	}
}
