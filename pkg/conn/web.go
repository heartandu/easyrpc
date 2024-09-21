package conn

import (
	"context"
	"fmt"

	"github.com/heartandu/grpc-web-go-client/grpcweb"
	"google.golang.org/grpc"
)

// WebClient is an adapter for a gRPC-Web client.
type WebClient struct {
	cc *grpcweb.ClientConn
}

// NewWebClient creates a new WebClient.
func NewWebClient(cc *grpcweb.ClientConn) *WebClient {
	return &WebClient{cc: cc}
}

// Invoke makes a unary gRPC call to the server.
func (c *WebClient) Invoke(ctx context.Context, method string, args, reply any, _ ...grpc.CallOption) error {
	if err := c.cc.Invoke(ctx, method, args, reply); err != nil {
		return fmt.Errorf("failed to call wrapped invoke: %w", err)
	}

	return nil
}

// NewStream creates a new client stream for making streaming gRPC calls to the server.
func (c *WebClient) NewStream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	_ ...grpc.CallOption,
) (grpc.ClientStream, error) {
	stream, err := c.cc.NewStream(ctx, desc, method)
	if err != nil {
		return nil, fmt.Errorf("failed to create new stream: %w", err)
	}

	return stream, nil
}
