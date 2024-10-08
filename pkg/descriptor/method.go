package descriptor

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

var ErrInvalidFQN = errors.New("invalid fully qualified method name")

// Method is an interface representing a gRPC method.
type Method interface {
	// RequestMessage returns the request proto message.
	RequestMessage() proto.Message
	// ResponseMessage returns the response proto message.
	ResponseMessage() proto.Message
	// StreamDesc returns the stream descriptor.
	StreamDesc() *grpc.StreamDesc
	// String returns the fully qualified method name.
	String() (string, error)
	// IsStreamingClient returns true if the method is a streaming client.
	IsStreamingClient() bool
	// IsStreamingServer returns true if the method is a streaming server.
	IsStreamingServer() bool
}

// methodWrapper is a struct implementing the Method interface.
type methodWrapper struct {
	rpc protoreflect.MethodDescriptor
}

// NewMethod creates a new Method instance with the given protoreflect.MethodDescriptor.
func NewMethod(rpc protoreflect.MethodDescriptor) Method {
	return &methodWrapper{rpc: rpc}
}

// RequestMessage parses the request message using the given RequestParser.
func (m *methodWrapper) RequestMessage() proto.Message {
	return dynamicpb.NewMessage(m.rpc.Input())
}

// ResponseMessage returns a new response message instance.
func (m *methodWrapper) ResponseMessage() proto.Message {
	return dynamicpb.NewMessage(m.rpc.Output())
}

// StreamDesc returns the stream descriptor for the method.
func (m *methodWrapper) StreamDesc() *grpc.StreamDesc {
	return &grpc.StreamDesc{
		StreamName:    string(m.rpc.Name()),
		ServerStreams: m.rpc.IsStreamingServer(),
		ClientStreams: m.rpc.IsStreamingClient(),
	}
}

// String returns the fully qualified method name.
func (m *methodWrapper) String() (string, error) {
	const minParts = 2

	parts := strings.Split(string(m.rpc.FullName()), ".")
	if len(parts) < minParts {
		return "", ErrInvalidFQN
	}

	return fmt.Sprintf("/%s/%s", strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]), nil
}

// IsStreamingClient returns true if the method is a streaming client.
func (m *methodWrapper) IsStreamingClient() bool {
	return m.rpc.IsStreamingClient()
}

// IsStreamingServer returns true if the method is a streaming server.
func (m *methodWrapper) IsStreamingServer() bool {
	return m.rpc.IsStreamingServer()
}
