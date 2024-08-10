package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/heartandu/easyrpc/pkg/config"
)

// Call represents a use case for making RPC calls.
type Call struct {
	output io.Writer
}

// NewCall returns a new instance of Call.
func NewCall(output io.Writer) *Call {
	return &Call{
		output: output,
	}
}

// MakeRPCCall makes an RPC call using the provided configuration and method name.
func (c *Call) MakeRPCCall(_ context.Context, cfg *config.Config, methodName string) error {
	fmt.Fprintf(c.output, "calling %s on address %s\n", methodName, cfg.Server.Address)

	// Perform the RPC call here

	return nil
}
