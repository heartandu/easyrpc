package usecase

import (
	"context"
	"errors"
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
func (c *Call) MakeRPCCall(ctx context.Context, methodName string) error {
	rpc, err := c.findMethod(methodName)
	if err != nil {
		return fmt.Errorf("failed to find method %q: %w", methodName, err)
	}

	switch {
	case rpc.IsStreamingClient() && rpc.IsStreamingServer():
		return ErrNotImplemented // TODO: implement me
	case rpc.IsStreamingServer():
		return ErrNotImplemented // TODO: implement me
	case rpc.IsStreamingClient():
		return c.clientStreamCall(ctx, rpc)
	default:
		return c.unaryCall(ctx, rpc)
	}
}

func (c *Call) clientStreamCall(ctx context.Context, rpc protoreflect.MethodDescriptor) error {
	resp := dynamicpb.NewMessage(rpc.Output())

	method, err := requestMethod(rpc)
	if err != nil {
		return fmt.Errorf("failed to convert method name: %w", err)
	}

	desc := &grpc.StreamDesc{
		StreamName:    string(rpc.Name()),
		ServerStreams: rpc.IsStreamingServer(),
		ClientStreams: rpc.IsStreamingClient(),
	}

	stream, err := c.cc.NewStream(ctx, desc, method)
	if err != nil {
		return fmt.Errorf("failed to create client stream: %w", err)
	}

	for {
		req, err := c.makeRequest(rpc)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to make request: %w", err)
		}

		if err := stream.SendMsg(req); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	if err := stream.CloseSend(); err != nil {
		return fmt.Errorf("failed to close client stream: %w", err)
	}

	if err := stream.RecvMsg(resp); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to receive message: %w", err)
	}

	if err := c.printResponse(resp); err != nil {
		return fmt.Errorf("failed to print response: %w", err)
	}

	return nil
}

func (c *Call) unaryCall(ctx context.Context, rpc protoreflect.MethodDescriptor) error {
	resp := dynamicpb.NewMessage(rpc.Output())

	req, err := c.makeRequest(rpc)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to make request: %w", err)
	}

	reqStr, err := requestMethod(rpc)
	if err != nil {
		return fmt.Errorf("failed to convert method name: %w", err)
	}

	if err = c.cc.Invoke(metadata.NewOutgoingContext(ctx, c.md), reqStr, req, resp); err != nil {
		return fmt.Errorf("failed to invoke rpc: %w", err)
	}

	if err := c.printResponse(resp); err != nil {
		return fmt.Errorf("failed to print response: %w", err)
	}

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

func (c *Call) makeRequest(rpc protoreflect.MethodDescriptor) (proto.Message, error) {
	msg := dynamicpb.NewMessage(rpc.Input())

	if err := c.rp.Next(msg); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	return msg, nil
}

func (c *Call) printResponse(resp *dynamicpb.Message) error {
	formattedResp, err := c.rf.Format(resp)
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	fmt.Fprintf(c.output, "%v\n", formattedResp)

	return nil
}

func requestMethod(rpc protoreflect.MethodDescriptor) (string, error) {
	const minParts = 2

	parts := strings.Split(string(rpc.FullName()), ".")
	if len(parts) < minParts {
		return "", ErrInvalidFQN
	}

	return fmt.Sprintf("/%s/%s", strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]), nil
}
