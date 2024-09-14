package test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/heartandu/easyrpc/internal/testdata"
	"github.com/heartandu/easyrpc/pkg/tlsconf"
)

const (
	insecureSocket    = ":50000"
	tlsSocket         = ":50001"
	insecureWebSocket = ":50002"
	tlsWebSocket      = ":50003"
	protocol          = "tcp"

	cacert = "../internal/testdata/rootCA.crt"
	cert   = "../internal/testdata/localhost.crt"
	key    = "../internal/testdata/localhost.key"

	importPath = "../internal/testdata"
	protoFile  = "test.proto"
)

func TestMain(m *testing.M) {
	os.Exit(runTest(m))
}

func runTest(m *testing.M) int {
	cfg, err := tlsconf.Config(cacert, cert, key)
	if err != nil {
		log.Printf("failed to get tls config: %v", err)
		return 1
	}

	insecureServer := newServer()
	defer insecureServer.Stop()

	tlsServer := newServer(grpc.Creds(credentials.NewTLS(cfg)))
	defer tlsServer.Stop()

	if err := serve(insecureServer, protocol, insecureSocket); err != nil {
		log.Printf("failed to serve insecure server: %v", err)
		return 1
	}

	if err := serve(tlsServer, protocol, tlsSocket); err != nil {
		log.Printf("failed to serve tls server: %v", err)
		return 1
	}

	if err := serveWeb(grpcweb.WrapServer(insecureServer), protocol, insecureWebSocket, nil); err != nil {
		log.Printf("failed to serve insecure web server: %v", err)
		return 1
	}

	if err := serveWeb(grpcweb.WrapServer(insecureServer), protocol, tlsWebSocket, cfg); err != nil {
		log.Printf("failed to serve tls web server: %v", err)
		return 1
	}

	return m.Run()
}

func newServer(opts ...grpc.ServerOption) *grpc.Server {
	s := grpc.NewServer(opts...)
	testdata.RegisterEchoServiceServer(s, &server{})
	reflection.Register(s)

	return s
}

func serve(s *grpc.Server, protocol, socket string) error {
	lis, err := net.Listen(protocol, socket)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve error: %v", err)
		}
	}()

	return nil
}

func serveWeb(s *grpcweb.WrappedGrpcServer, protocol, socket string, tlsCfg *tls.Config) error {
	srv := http.Server{
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if s.IsGrpcWebSocketRequest(req) {
				s.HandleGrpcWebsocketRequest(resp, req)
				return
			}

			if s.IsGrpcWebRequest(req) {
				s.HandleGrpcWebRequest(resp, req)
				return
			}

			// Fall back to other servers.
			http.DefaultServeMux.ServeHTTP(resp, req)
		}),
		TLSConfig: tlsCfg,
	}

	lis, err := net.Listen(protocol, socket)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if tlsCfg != nil {
			if err := srv.ServeTLS(lis, "", ""); err != nil {
				log.Fatalf("serve tls error: %v", err)
			}
		} else {
			if err := srv.Serve(lis); err != nil {
				log.Fatalf("serve error: %v", err)
			}
		}
	}()

	return nil
}

type server struct {
	testdata.UnimplementedEchoServiceServer
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

func address(socket string) string {
	return "localhost" + socket
}
