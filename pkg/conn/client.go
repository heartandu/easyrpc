package conn

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	Invoke(ctx context.Context, method string, args any, reply any) error
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string) (ClientStream, error)
}

type ClientStream interface {
	Header() (metadata.MD, error)
	Trailer() metadata.MD
	CloseSend() error
	Context() context.Context
	SendMsg(m any) error
	RecvMsg(m any) error
}

type ServerStream interface {
	SetHeader(md metadata.MD) error
	SendHeader(md metadata.MD) error
	SetTrailer(md metadata.MD)
	Context() context.Context
	SendMsg(m any) error
	RecvMsg(m any) error
}
