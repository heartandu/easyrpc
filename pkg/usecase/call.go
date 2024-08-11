package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

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
func (c *Call) MakeRPCCall(ctx context.Context, cfg *config.Config, methodName string, rawRequest io.ReadCloser) error {
	fd, err := findDescriptor(ctx, cfg.Proto.ImportPaths, cfg.Proto.ProtoFiles, methodName)
	if err != nil {
		return fmt.Errorf("compile proto failed: %w", err)
	}

	rpc, ok := fd.(protoreflect.MethodDescriptor)
	if !ok {
		return ErrNotAMethod
	}

	if rpc.IsStreamingClient() || rpc.IsStreamingServer() {
		return ErrNotImplemented
	}

	resp := dynamicpb.NewMessage(rpc.Output())

	req, err := makeRequest(rpc, rawRequest)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	client, err := grpc.NewClient(cfg.Server.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	parts := strings.Split(string(rpc.FullName()), ".")
	if len(parts) < 2 {
		return ErrInvalidFQN
	}

	fqrn := fmt.Sprintf("/%s/%s", strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1])

	if err := client.Invoke(ctx, fqrn, req, resp); err != nil {
		return fmt.Errorf("failed to invoke rpc: %w", err)
	}

	fmt.Fprintf(c.output, "resp: %v\n", resp)

	return nil
}

func findDescriptor(ctx context.Context, importPaths, files []string, fqn string) (protoreflect.Descriptor, error) {
	comp := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: importPaths,
		}),
	}

	fds, err := comp.Compile(ctx, files...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto files: %w", err)
	}

	for _, fd := range fds {
		if desc := fd.FindDescriptorByName(protoreflect.FullName(fqn)); desc != nil {
			return desc, nil
		}
	}

	return nil, ErrDescriptorNotFound
}

func makeRequest(rpc protoreflect.MethodDescriptor, rawRequest io.ReadCloser) (*dynamicpb.Message, error) {
	defer rawRequest.Close()

	req := dynamicpb.NewMessage(rpc.Input())

	var rawJSON json.RawMessage
	if err := json.NewDecoder(rawRequest).Decode(&rawJSON); err != nil {
		if errors.Is(err, io.EOF) {
			return req, nil
		}

		return nil, fmt.Errorf("failed to decode raw request: %w", err)
	}

	if err := protojson.Unmarshal(rawJSON, req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proto: %w", err)
	}

	return req, nil
}
