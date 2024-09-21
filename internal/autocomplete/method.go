package autocomplete

import (
	"context"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/heartandu/easyrpc/internal/client"
	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/internal/proto"
)

// ProtoComp represents a protobuf symbol autocompletion functionality.
type ProtoComp struct {
	fs    afero.Fs
	viper *viper.Viper

	cfg config.Config
}

// NewProtoComp creates a new ProtoComp instance.
func NewProtoComp(fs afero.Fs, v *viper.Viper) *ProtoComp {
	return &ProtoComp{
		fs:    fs,
		viper: v,
	}
}

// CompleteMethod provides autocomplete suggestions for methods.
func (c *ProtoComp) CompleteMethod(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if err := c.viper.Unmarshal(&c.cfg); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	methods, err := c.symbols(toComplete, func(pkg, svc, method string) string {
		const methodPartsCount = 3

		result := make([]string, 0, methodPartsCount)

		if c.cfg.Request.Package != "" {
			if pkg != c.cfg.Request.Package {
				return ""
			}
		} else {
			result = append(result, pkg)
		}

		if c.cfg.Request.Service != "" {
			if svc != c.cfg.Request.Service {
				return ""
			}
		} else {
			result = append(result, svc)
		}

		result = append(result, method)

		return strings.Join(result, ".")
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return methods, cobra.ShellCompDirectiveNoFileComp
}

// CompletePackage provides autocomplete suggestions for package names.
func (c *ProtoComp) CompletePackage(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if err := c.viper.Unmarshal(&c.cfg); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	packages, err := c.symbols(toComplete, func(pkg, _, _ string) string {
		return pkg
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return packages, cobra.ShellCompDirectiveNoFileComp
}

// CompleteService provides autocomplete suggestions for service names.
func (c *ProtoComp) CompleteService(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if err := c.viper.Unmarshal(&c.cfg); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	services, err := c.symbols(toComplete, func(pkg, svc, _ string) string {
		if c.cfg.Request.Package != "" && pkg != c.cfg.Request.Package {
			return ""
		}

		return svc
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return services, cobra.ShellCompDirectiveNoFileComp
}

func (c *ProtoComp) symbols(toComplete string, filterMapFunc func(pkg, svc, method string) string) ([]string, error) {
	ctx := context.Background()

	cc, err := client.New(c.fs, &c.cfg)
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in authocomplete.
	}

	descSrc, err := proto.NewDescriptorSource(ctx, c.fs, &c.cfg, cc)
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in authocomplete.
	}

	methods, err := descSrc.ListMethods()
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in authocomplete.
	}

	result := map[string]struct{}{}

	isCaseInsensitive := toComplete == strings.ToLower(toComplete)

	completionToCompare := toComplete
	if isCaseInsensitive {
		completionToCompare = strings.ToLower(completionToCompare)
	}

	for symbol := range filterMapIter(methods, filterMapFunc) {
		symbolToCompare := symbol

		if isCaseInsensitive {
			symbolToCompare = strings.ToLower(symbolToCompare)
		}

		if strings.Contains(symbolToCompare, completionToCompare) {
			result[symbol] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(result)), nil
}

func filterMapIter(s []string, f func(pkg, svc, method string) string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, s := range s {
			const (
				minParts = 3
				delim    = "."
			)

			if f != nil {
				parts := strings.Split(s, delim)
				if len(parts) < minParts {
					continue
				}

				s = f(strings.Join(parts[:len(parts)-2], delim), parts[len(parts)-2], parts[len(parts)-1])
				if s == "" {
					continue
				}
			}

			if cont := yield(s); !cont {
				return
			}
		}
	}
}
