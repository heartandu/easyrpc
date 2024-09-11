package conn

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GrpcClient struct {
	cc grpc.ClientConnInterface
}

func NewGrpcClient(cc grpc.ClientConnInterface) *GrpcClient {
	return &GrpcClient{
		cc: cc,
	}
}

func (c *GrpcClient) Invoke(ctx context.Context, method string, args any, reply any) error {
	return c.cc.Invoke(ctx, method, args, reply)
}

func (c *GrpcClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string) (ClientStream, error) {
	stream, err := c.cc.NewStream(ctx, desc, method)
	if err != nil {
		return nil, fmt.Errorf("failed to call wrapped NewStream: %w", err)
	}

	return newGrpcStream(stream), nil
}

type grpcStream struct {
	stream grpc.ClientStream
}

func newGrpcStream(stream grpc.ClientStream) *grpcStream {
	return &grpcStream{stream: stream}
}

func (s *grpcStream) Header() (metadata.MD, error) {
	return s.stream.Header()
}

func (s *grpcStream) Trailer() metadata.MD {
	return s.stream.Trailer()
}

func (s *grpcStream) CloseSend() error {
	return s.stream.CloseSend()
}

func (s *grpcStream) Context() context.Context {
	return s.stream.Context()
}

func (s *grpcStream) SendMsg(m any) error {
	return s.stream.SendMsg(m)
}

func (s *grpcStream) RecvMsg(m any) error {
	return s.stream.RecvMsg(m)
}
