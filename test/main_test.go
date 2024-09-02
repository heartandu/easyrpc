package test

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/heartandu/easyrpc/internal/testdata"
)

const (
	socket   = "/tmp/test.sock"
	protocol = "unix"
	address  = protocol + "://" + socket
)

func TestMain(m *testing.M) {
	os.Exit(runTest(m))
}

func runTest(m *testing.M) int {
	lis, err := net.Listen(protocol, socket)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	defer s.Stop()

	testdata.RegisterEchoServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve error: %v", err)
		}
	}()

	return m.Run()
}

type server struct {
	testdata.UnimplementedEchoServiceServer
}

func (*server) Echo(_ context.Context, r *testdata.EchoRequest) (*testdata.EchoResponse, error) {
	return &testdata.EchoResponse{Msg: r.GetMsg()}, nil
}
