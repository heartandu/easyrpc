package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/heartandu/easyrpc/internal/testdata/proto"
	"github.com/heartandu/easyrpc/internal/testdata/proto/echo"
	"github.com/heartandu/easyrpc/internal/testdata/proto/types"
)

type packagelessServer struct {
	proto.UnimplementedEchoServiceServer
	proto.UnimplementedTimeServiceServer
}

func (s *packagelessServer) Now(ctx context.Context, _ *proto.NowRequest) (*proto.NowResponse, error) {
	return &proto.NowResponse{Timestamp: time.Now().Unix()}, nil
}

func (s *packagelessServer) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return &proto.EchoResponse{Msg: req.GetMsg()}, nil
}

type server struct {
	echo.UnimplementedEchoServiceServer
	types.UnimplementedTypesServiceServer
}

func (s *server) Echo(ctx context.Context, r *echo.EchoRequest) (*echo.EchoResponse, error) {
	msg := r.GetMsg()

	if testVal := s.getTestMDKey(ctx); testVal != "" {
		msg += "\n" + testVal
	}

	return &echo.EchoResponse{Msg: msg}, nil
}

func (*server) Error(_ context.Context, r *echo.ErrorRequest) (*echo.ErrorResponse, error) {
	return nil, status.Error(codes.Internal, "internal error")
}

func (*server) ScalarTypes(_ context.Context, r *types.ScalarTypes) (*types.ScalarTypes, error) {
	return r, nil
}

func (*server) EnumTypes(_ context.Context, r *types.EnumTypes) (*types.EnumTypes, error) {
	return r, nil
}

func (*server) Maps(_ context.Context, r *types.Maps) (*types.Maps, error) {
	return r, nil
}

func (*server) Oneof(_ context.Context, r *types.Oneof) (*types.Oneof, error) {
	return r, nil
}

func (*server) Imported(_ context.Context, r *types.Imported) (*types.Imported, error) {
	return r, nil
}

func (*server) Recursive(_ context.Context, r *types.Recursive) (*types.Recursive, error) {
	return r, nil
}

func (*server) Optional(_ context.Context, r *types.Optional) (*types.Optional, error) {
	return r, nil
}

func (*server) Repeated(_ context.Context, r *types.Repeated) (*types.Repeated, error) {
	return r, nil
}

func (s *server) ClientStream(
	stream grpc.ClientStreamingServer[echo.ClientStreamRequest, echo.ClientStreamResponse],
) error {
	resp := &echo.ClientStreamResponse{}

	for {
		r, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to receive message: %w", err)
		}

		resp.Msgs = append(resp.Msgs, r.GetMsg())
	}

	if testVal := s.getTestMDKey(stream.Context()); testVal != "" {
		resp.Msgs = append(resp.Msgs, testVal)
	}

	if err := stream.SendAndClose(resp); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *server) ServerStream(
	r *echo.ServerStreamRequest,
	stream grpc.ServerStreamingServer[echo.ServerStreamResponse],
) error {
	for _, msg := range r.GetMsgs() {
		if err := stream.Send(&echo.ServerStreamResponse{Msg: msg}); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	if testVal := s.getTestMDKey(stream.Context()); testVal != "" {
		if err := stream.Send(&echo.ServerStreamResponse{Msg: testVal}); err != nil {
			return fmt.Errorf("failed to send md message: %w", err)
		}
	}

	return nil
}

func (s *server) BidiStream(
	stream grpc.BidiStreamingServer[echo.BidiStreamRequest, echo.BidiStreamResponse],
) error {
	var responses []*echo.BidiStreamResponse

	for {
		r, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to receive message: %w", err)
		}

		responses = append(responses, &echo.BidiStreamResponse{Msg: r.GetMsg()})
	}

	if testVal := s.getTestMDKey(stream.Context()); testVal != "" {
		responses = append(responses, &echo.BidiStreamResponse{Msg: testVal})
	}

	for _, resp := range responses {
		if err := stream.Send(resp); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	return nil
}

func (*server) getTestMDKey(ctx context.Context) string {
	const testMDKey = "test"

	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get(testMDKey)) != 0 {
		return md.Get(testMDKey)[0]
	}

	return ""
}
