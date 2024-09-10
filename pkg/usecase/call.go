package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
func (c *Call) MakeRPCCall(ctx context.Context, methodName string) error {
	m, err := c.findMethod(methodName)
	if err != nil {
		return fmt.Errorf("failed to find method %q: %w", methodName, err)
	}

	ctx = metadata.NewOutgoingContext(ctx, c.md)

	if m.IsStreamingClient() || m.IsStreamingServer() {
		return c.streamCall(ctx, m)
	}

	return c.unaryCall(ctx, m)
}

func (c *Call) streamCall(ctx context.Context, m descriptor.Method) error {
	method, err := m.String()
	if err != nil {
		return fmt.Errorf("failed to get method name: %w", err)
	}

	stream, err := c.cc.NewStream(ctx, m.StreamDesc(), method)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	if err := c.streamRequestMessages(stream, m); err != nil {
		return fmt.Errorf("failed to stream request messages: %w", err)
	}

	if err := stream.CloseSend(); err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}

	if err := c.streamResponseMessages(stream, m); err != nil {
		return fmt.Errorf("failed to stream response messages: %w", err)
	}

	return nil
}

func (c *Call) unaryCall(ctx context.Context, m descriptor.Method) error {
	resp := m.ResponseMessage()

	req, err := m.RequestMessage(c.rp)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to make request: %w", err)
	}

	method, err := m.String()
	if err != nil {
		return fmt.Errorf("failed to convert method name: %w", err)
	}

	if err = c.cc.Invoke(ctx, method, req, resp); err != nil {
		return fmt.Errorf("failed to invoke rpc: %w", err)
	}

	if err := c.printResponse(resp); err != nil {
		return fmt.Errorf("failed to print response: %w", err)
	}

	return nil
}

func (c *Call) streamRequestMessages(stream grpc.ClientStream, m descriptor.Method) error {
	for {
		req, err := m.RequestMessage(c.rp)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("failed to make request: %w", err)
		}

		if err := stream.SendMsg(req); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}
}

func (c *Call) streamResponseMessages(stream grpc.ClientStream, m descriptor.Method) error {
	for {
		resp := m.ResponseMessage()
		if err := stream.RecvMsg(resp); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("failed to receive message: %w", err)
		}

		if err := c.printResponse(resp); err != nil {
			return fmt.Errorf("failed to print response: %w", err)
		}
	}
}

func (c *Call) findMethod(methodName string) (descriptor.Method, error) {
	fd, err := c.ds.FindSymbol(methodName)
	if err != nil {
		return nil, fmt.Errorf("failed to find symbol: %w", err)
	}

	if rpc, ok := fd.(protoreflect.MethodDescriptor); ok {
		return descriptor.NewMethod(rpc), nil
	}

	return nil, ErrNotAMethod
}

func (c *Call) printResponse(resp *dynamicpb.Message) error {
	formattedResp, err := c.rf.Format(resp)
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	fmt.Fprintf(c.output, "%v\n", formattedResp)

	return nil
}
