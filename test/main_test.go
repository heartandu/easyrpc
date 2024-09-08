package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

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
	insecureSocket  = ":50000"
	tlsSocket       = ":50001"
	protocol        = "tcp"
	insecureAddress = "localhost" + insecureSocket
	tlsAddress      = "localhost" + tlsSocket

	cacert  = "../internal/testdata/rootCA.crt"
	cert    = "../internal/testdata/localhost.crt"
	certKey = "../internal/testdata/localhost.key"

	importPath = "../internal/testdata"
	protoFile  = "test.proto"
)

func TestMain(m *testing.M) {
	os.Exit(runTest(m))
}

func runTest(m *testing.M) int {
	cfg, err := tlsconf.Config(cacert, cert, certKey)
	if err != nil {
		log.Printf("failed to get tls config: %v", err)
		return 1
	}

	insecureServer, err := serve(protocol, insecureSocket)
	if err != nil {
		log.Printf("failed to serve insecure server: %v", err)
		return 1
	}
	defer insecureServer.Stop()

	tlsServer, err := serve(protocol, tlsSocket, grpc.Creds(credentials.NewTLS(cfg)))
	if err != nil {
		log.Printf("failed to serve tls server: %v", err)
		return 1
	}
	defer tlsServer.Stop()

	return m.Run()
}

func serve(protocol, socket string, opts ...grpc.ServerOption) (*grpc.Server, error) {
	lis, err := net.Listen(protocol, socket)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer(opts...)
	testdata.RegisterEchoServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve error: %v", err)
		}
	}()

	return s, nil
}

type server struct {
	testdata.UnimplementedEchoServiceServer
}

func (*server) Echo(ctx context.Context, r *testdata.EchoRequest) (*testdata.EchoResponse, error) {
	const testKey = "test"

	msg := r.GetMsg()

	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get(testKey)) != 0 {
		msg += "\n" + md.Get(testKey)[0]
	}

	return &testdata.EchoResponse{Msg: msg}, nil
}

func (*server) Error(_ context.Context, r *testdata.ErrorRequest) (*testdata.ErrorResponse, error) {
	return nil, status.Error(codes.Internal, "internal error")
}
