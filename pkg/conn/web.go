package conn

import (
	"context"
	"errors"
	"fmt"

	"github.com/heartandu/grpc-web-go-client/grpcweb"
	"google.golang.org/grpc"
)

var ErrNotAStreamRequest = errors.New("not a stream request")

type WebClient struct {
	cc *grpcweb.ClientConn
}

func NewWebClient(cc *grpcweb.ClientConn) *WebClient {
	return &WebClient{cc: cc}
}

func (c *WebClient) Invoke(ctx context.Context, method string, args, reply any, _ ...grpc.CallOption) error {
	if err := c.cc.Invoke(ctx, method, args, reply); err != nil {
		return fmt.Errorf("failed to call wrapped invoke: %w", err)
	}

	return nil
}

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
