package conn

import (
	"context"
	"errors"
	"fmt"

	"github.com/ktr0731/grpc-web-go-client/grpcweb"
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
	switch {
	case desc.ClientStreams && desc.ServerStreams:
		stream, err := c.cc.NewBidiStream(desc, method)
		if err != nil {
			return nil, fmt.Errorf("failed to create bidi stream: %w", err)
		}

		return &webBidiStream{ctx: ctx, stream: stream}, nil
	case desc.ClientStreams:
		stream, err := c.cc.NewClientStream(desc, method)
		if err != nil {
			return nil, fmt.Errorf("failed to create client stream: %w", err)
		}

		return &webClientStream{ctx: ctx, stream: stream}, nil
	case desc.ServerStreams:
		stream, err := c.cc.NewServerStream(desc, method)
		if err != nil {
			return nil, fmt.Errorf("failed to create server stream: %w", err)
		}

		return &webServerStream{ctx: ctx, stream: stream}, nil
	default:
		return nil, ErrNotAStreamRequest
	}
}

type webClientStream struct {
	ctx    context.Context
	stream grpcweb.ClientStream
}

func (s *webClientStream) Header() (metadata.MD, error) {
	return s.stream.Header()
}

func (s *webClientStream) Trailer() metadata.MD {
	return s.stream.Trailer()
}

func (s *webClientStream) CloseSend() error {
	return nil
}

func (s *webClientStream) Context() context.Context {
	return s.ctx
}

func (s *webClientStream) SendMsg(m any) error {
	return s.stream.Send(s.ctx, m)
}

func (s *webClientStream) RecvMsg(m any) error {
	return s.stream.CloseAndReceive(s.ctx, m)
}

type webServerStream struct {
	ctx    context.Context
	stream grpcweb.ServerStream
}

func (s *webServerStream) Header() (metadata.MD, error) {
	return s.stream.Header()
}

func (s *webServerStream) Trailer() metadata.MD {
	return s.stream.Trailer()
}

func (s *webServerStream) CloseSend() error {
	return nil
}

func (s *webServerStream) Context() context.Context {
	return s.ctx
}

func (s *webServerStream) SendMsg(m any) error {
	return s.stream.Send(s.ctx, m)
}

func (s *webServerStream) RecvMsg(m any) error {
	return s.stream.Receive(s.ctx, m)
}

type webBidiStream struct {
	ctx    context.Context
	stream grpcweb.BidiStream
}

func (s *webBidiStream) Header() (metadata.MD, error) {
	return s.stream.Header()
}

func (s *webBidiStream) Trailer() metadata.MD {
	return s.stream.Trailer()
}

func (s *webBidiStream) CloseSend() error {
	return s.stream.CloseSend()
}

func (s *webBidiStream) Context() context.Context {
	return s.ctx
}

func (s *webBidiStream) SendMsg(m any) error {
	return s.stream.Send(s.ctx, m)
}

func (s *webBidiStream) RecvMsg(m any) error {
	return s.stream.Receive(s.ctx, m)
}
