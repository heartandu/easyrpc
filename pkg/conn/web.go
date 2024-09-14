package conn

import (
	"context"
	"errors"
	"fmt"

	"github.com/heartandu/grpc-web-go-client/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var ErrNotAStreamRequest = errors.New("not a stream request")

type WebClient struct {
	cc *grpcweb.ClientConn
}

func NewWebClient(cc *grpcweb.ClientConn) *WebClient {
	return &WebClient{
		cc: cc,
	}
}

func (c *WebClient) Invoke(ctx context.Context, method string, args any, reply any) error {
	return c.cc.Invoke(ctx, method, args, reply)
}

func (c *WebClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string) (ClientStream, error) {
	var stream grpcweb.Stream
	var err error

	switch {
	case desc.ClientStreams && desc.ServerStreams:
		stream, err = c.cc.NewBidiStream(ctx, desc, method)
		if err != nil {
			return nil, fmt.Errorf("failed to create bidi stream: %w", err)
		}
	case desc.ClientStreams:
		stream, err = c.cc.NewClientStream(ctx, desc, method)
		if err != nil {
			return nil, fmt.Errorf("failed to create client stream: %w", err)
		}
	case desc.ServerStreams:
		stream, err = c.cc.NewServerStream(ctx, desc, method)
		if err != nil {
			return nil, fmt.Errorf("failed to create server stream: %w", err)
		}
	default:
		return nil, ErrNotAStreamRequest
	}

	return &webStream{stream}, nil
}

type webStream struct {
	stream grpcweb.Stream
}

func (s *webStream) Header() (metadata.MD, error) {
	return s.stream.Header()
}

func (s *webStream) Trailer() metadata.MD {
	return s.stream.Trailer()
}

func (s *webStream) CloseSend() error {
	return s.stream.CloseSend()
}

func (s *webStream) Context() context.Context {
	return s.stream.Context()
}

func (s *webStream) SendMsg(m any) error {
	return s.stream.SendMsg(m)
}

func (s *webStream) RecvMsg(m any) error {
	return s.stream.RecvMsg(m)
}
