package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/heartandu/easyrpc/pkg/descriptor"
	"github.com/heartandu/easyrpc/pkg/format"
)

// Call represents a use case for making RPC calls.
type Call struct {
	output io.Writer
	ds     descriptor.Source
	cc     grpc.ClientConnInterface
	rp     format.RequestParser
	rf     format.ResponseFormatter
	md     metadata.MD
}

// NewCall returns a new instance of Call.
func NewCall(
	output io.Writer,
	descSrc descriptor.Source,
	clientConn grpc.ClientConnInterface,
	reqParser format.RequestParser,
	respFormatter format.ResponseFormatter,
	md metadata.MD,
) *Call {
	return &Call{
		output: output,
		ds:     descSrc,
		cc:     clientConn,
		rp:     reqParser,
		rf:     respFormatter,
		md:     md,
	}
}

// MakeRPCCall makes an RPC call using the provided configuration and method name.
func (c *Call) MakeRPCCall(ctx context.Context, methodName string, rawRequest io.ReadCloser) error {
	rpc, err := c.findMethod(methodName)
	if err != nil {
		return fmt.Errorf("failed to find method %q: %w", methodName, err)
	}

	// TODO: Implement this.
	if rpc.IsStreamingClient() || rpc.IsStreamingServer() {
		return ErrNotImplemented
	}

	resp := dynamicpb.NewMessage(rpc.Output())

	req, err := c.makeRequest(rpc, rawRequest)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	reqStr, err := toRequestString(rpc)
	if err != nil {
		return fmt.Errorf("failed to convert method name:: %w", err)
	}

	if err = c.cc.Invoke(metadata.NewOutgoingContext(ctx, c.md), reqStr, req, resp); err != nil {
		return fmt.Errorf("failed to invoke rpc: %w", err)
	}

	formattedResp, err := c.rf.Format(resp)
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	fmt.Fprintf(c.output, "%v\n", formattedResp)

	return nil
}

func (c *Call) findMethod(methodName string) (protoreflect.MethodDescriptor, error) {
	fd, err := c.ds.FindSymbol(methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbol: %w", err)
	}

	if rpc, ok := fd.(protoreflect.MethodDescriptor); ok {
		return rpc, nil
	}

	return nil, ErrNotAMethod
}

func (c *Call) makeRequest(rpc protoreflect.MethodDescriptor, rawRequest io.ReadCloser) (proto.Message, error) {
	defer rawRequest.Close()

	msg := dynamicpb.NewMessage(rpc.Input())

	if err := c.rp.Parse(msg); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return msg, nil
}

func toRequestString(rpc protoreflect.MethodDescriptor) (string, error) {
	const minParts = 2

	parts := strings.Split(string(rpc.FullName()), ".")
	if len(parts) < minParts {
		return "", ErrInvalidFQN
	}

	return fmt.Sprintf("/%s/%s", strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]), nil
}
