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
)

const delim = "."

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
	_ *cobra.Command,
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

	methods, err := c.symbols(&cfg, toComplete, func(pkg, svc, method string) string {
		const methodPartsCount = 3

		result := make([]string, 0, methodPartsCount)

		// If package name has been set, and doesn't match
		// the received one, filter the method out.
		// If they match, omit the package name from the result.
		// Otherwise, use it to form a fully qualified name.
		if cfg.Request.Package != "" {
			if pkg != cfg.Request.Package {
				return ""
			}
		} else {
			result = append(result, pkg)
		}

		// If service name has been set, and doesn't match
		// the received one, filter the method out.
		// If they match, omit the service name from the result.
		// Otherwise, use it to form a fully qualified name.
		if cfg.Request.Service != "" {
			// Service name may as well be a fully qualivied service name,
			// so we also need to handle this special case.
			if !servicesMatch(cfg.Request.Service, pkg, svc) {
				return ""
			}

			// If we only specified the service name, we need to append it to obtain the correct method name.
			// Otherwise, we would render "example.Method" from the method name "example.Service.Method",
			// which would be invalid.
			if len(result) > 0 {
				result = append(result, svc)
			}
		} else {
			result = append(result, svc)
		}

		result = append(result, method)

		return strings.Join(result, delim)
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

	cfg, err := c.cfgFunc()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	packages, err := c.symbols(&cfg, toComplete, func(pkg, _, _ string) string {
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

	cfg, err := c.cfgFunc()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	services, err := c.symbols(&cfg, toComplete, func(pkg, svc, _ string) string {
		const servicePartsCount = 2

		parts := make([]string, 0, servicePartsCount)

		// If package name has been set, and doesn't match
		// the received one, filter the method out.
		// If they match, omit the package name from the result.
		// Otherwise, use it to form a fully qualified name.
		if cfg.Request.Package != "" {
			if pkg != cfg.Request.Package {
				return ""
			}
		} else {
			parts = append(parts, pkg)
		}

		parts = append(parts, svc)

		return strings.Join(parts, delim)
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return services, cobra.ShellCompDirectiveNoFileComp
}

func (c *ProtoComp) symbols(
	cfg *config.Config,
	toComplete string,
	filterMapFunc func(pkg, svc, method string) string,
) ([]string, error) {
	ctx := context.Background()

	cc, err := client.New(c.fs, cfg)
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in authocomplete.
	}

	descSrc, err := proto.NewDescriptorSource(ctx, c.fs, cfg, cc)
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in authocomplete.
	}

	methods, err := descSrc.ListMethods()
	if err != nil {
		return nil, err //nolint:wrapcheck // Error wrapping is unnecessary in authocomplete.
	}

	encounteredSymbols := map[string]struct{}{}
	result := make([]string, 0)

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
			if _, ok := encounteredSymbols[symbol]; !ok {
				result = append(result, symbol)
			}

			encounteredSymbols[symbol] = struct{}{}
		}
	}

	return result, nil
}

func filterMapIter(s []string, f func(pkg, svc, method string) string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, s := range s {
			const minParts = 3

			parts := strings.Split(s, delim)
			if len(parts) < minParts {
				continue
			}

			s = f(strings.Join(parts[:len(parts)-2], delim), parts[len(parts)-2], parts[len(parts)-1])
			if s == "" {
				continue
			}

			if cont := yield(s); !cont {
				return
			}
		}
	}
}

func servicesMatch(cfgService, pkg, svc string) bool {
	svcParts := strings.Split(cfgService, delim)
	if len(svcParts) > 1 {
		if pkg != strings.Join(svcParts[:len(svcParts)-1], delim) {
			return false
		}
	}

	if svc != svcParts[len(svcParts)-1] {
		return false
	}

	return true
}
