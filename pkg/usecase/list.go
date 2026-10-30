package usecase

import (
	"fmt"
	"io"
	"sort"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/fqn"
)

// List represents a use case for listing available RPC methods.
type List struct {
	out io.Writer
	ds  descriptor.Source
}

// NewList returns a new instance of List.
func NewList(out io.Writer, ds descriptor.Source) *List {
	return &List{
		out: out,
		ds:  ds,
	}
}

// Run retrieves and prints the list of available RPC methods to the output.
// Methods are filtered and formatted according to the provided package and service.
func (l *List) Run(pkg, svc string) error {
	methods, err := l.ds.ListMethods()
	if err != nil {
		return fmt.Errorf("failed to list methods: %w", err)
	}

	seen := make(map[string]struct{}, len(methods))
	result := make([]string, 0, len(methods))

	for _, method := range methods {
		formatted, ok := fqn.ParseFQMN(method).FilterAndFormat(pkg, svc)
		if !ok {
			continue
		}

		if _, exists := seen[formatted]; exists {
			continue
		}

		seen[formatted] = struct{}{}
		result = append(result, formatted)
	}

	sort.Strings(result)

	for _, method := range result {
		if _, err := fmt.Fprintln(l.out, method); err != nil {
			return fmt.Errorf("failed to write method: %w", err)
		}
	}

	return nil
}
