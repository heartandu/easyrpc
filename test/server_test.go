package test

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/heartandu/easyrpc/internal/testdata"
)

type server struct {
	testdata.UnimplementedEchoServiceServer
	testdata.UnimplementedTypesServiceServer
}

func (s *server) Echo(ctx context.Context, r *testdata.EchoRequest) (*testdata.EchoResponse, error) {
	msg := r.GetMsg()

	if testVal := s.getTestMDKey(ctx); testVal != "" {
		msg += "\n" + testVal
	}

	return &testdata.EchoResponse{Msg: msg}, nil
}

func (*server) Error(_ context.Context, r *testdata.ErrorRequest) (*testdata.ErrorResponse, error) {
	return nil, status.Error(codes.Internal, "internal error")
}

func (*server) ScalarTypes(_ context.Context, r *testdata.ScalarTypes) (*testdata.ScalarTypes, error) {
	return r, nil
}

func (*server) EnumTypes(_ context.Context, r *testdata.EnumTypes) (*testdata.EnumTypes, error) {
	return r, nil
}

func (*server) Maps(_ context.Context, r *testdata.Maps) (*testdata.Maps, error) {
	return r, nil
}

func (*server) Oneof(_ context.Context, r *testdata.Oneof) (*testdata.Oneof, error) {
	return r, nil
}

func (*server) Nested(_ context.Context, r *testdata.Nested) (*testdata.Nested, error) {
	return r, nil
}

func (*server) Recursive(_ context.Context, r *testdata.Recursive) (*testdata.Recursive, error) {
	return r, nil
}

func (*server) Optional(_ context.Context, r *testdata.Optional) (*testdata.Optional, error) {
	return r, nil
}

func (*server) Repeated(_ context.Context, r *testdata.Repeated) (*testdata.Repeated, error) {
	return r, nil
}

func (s *server) ClientStream(
	stream grpc.ClientStreamingServer[testdata.ClientStreamRequest, testdata.ClientStreamResponse],
) error {
	resp := &testdata.ClientStreamResponse{}

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
	r *testdata.ServerStreamRequest,
	stream grpc.ServerStreamingServer[testdata.ServerStreamResponse],
) error {
	for _, msg := range r.GetMsgs() {
		if err := stream.Send(&testdata.ServerStreamResponse{Msg: msg}); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	if testVal := s.getTestMDKey(stream.Context()); testVal != "" {
		if err := stream.Send(&testdata.ServerStreamResponse{Msg: testVal}); err != nil {
			return fmt.Errorf("failed to send md message: %w", err)
		}
	}

	return nil
}

func (s *server) BidiStream(
	stream grpc.BidiStreamingServer[testdata.BidiStreamRequest, testdata.BidiStreamResponse],
) error {
	var responses []*testdata.BidiStreamResponse

	for {
		r, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to receive message: %w", err)
		}

		responses = append(responses, &testdata.BidiStreamResponse{Msg: r.GetMsg()})
	}

	if testVal := s.getTestMDKey(stream.Context()); testVal != "" {
		responses = append(responses, &testdata.BidiStreamResponse{Msg: testVal})
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
