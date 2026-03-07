package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/heartandu/easyrpc/internal/testdata/proto"
	"github.com/heartandu/easyrpc/internal/testdata/proto/echo"
	"github.com/heartandu/easyrpc/internal/testdata/proto/types"
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

	importPath = "../internal/testdata/proto"
	protoFile  = "echo/echo.proto"
)

func TestMain(m *testing.M) {
	code, err := runTest(m)
	if err != nil {
		log.Printf("error: %s", err)
	}

	os.Exit(code)
}

func runTest(m *testing.M) (int, error) {
	fs := afero.NewOsFs()

	cfg, err := tlsconf.Config(fs, cacert, cert, key)
	if err != nil {
		return 1, fmt.Errorf("failed to get tls config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	insecureServer := newServer()
	defer insecureServer.Stop()

	tlsServer := newServer(grpc.Creds(credentials.NewTLS(cfg)))
	defer tlsServer.Stop()

	if err := serve(ctx, insecureServer, protocol, insecureSocket); err != nil {
		return 1, fmt.Errorf("failed toerve insecure server: %w", err)
	}

	if err := serve(ctx, tlsServer, protocol, tlsSocket); err != nil {
		return 1, fmt.Errorf("failed to serve tls server: %w", err)
	}

	if err := serveWeb(ctx, grpcweb.WrapServer(insecureServer), protocol, insecureWebSocket, nil); err != nil {
		return 1, fmt.Errorf("failed to serve insecure web server: %w", err)
	}

	if err := serveWeb(ctx, grpcweb.WrapServer(insecureServer), protocol, tlsWebSocket, cfg); err != nil {
		return 1, fmt.Errorf("failed to serve tls web server: %w", err)
	}

	return m.Run(), nil
}

func newServer(opts ...grpc.ServerOption) *grpc.Server {
	s := grpc.NewServer(opts...)
	serverImpl := &server{}
	packagelessServerImpl := &packagelessServer{}

	echo.RegisterEchoServiceServer(s, serverImpl)
	types.RegisterTypesServiceServer(s, serverImpl)
	proto.RegisterTimeServiceServer(s, packagelessServerImpl)
	proto.RegisterEchoServiceServer(s, packagelessServerImpl)
	reflection.Register(s)

	return s
}

func serve(ctx context.Context, s *grpc.Server, protocol, socket string) error {
	lis, err := (&net.ListenConfig{}).Listen(ctx, protocol, socket)
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

func serveWeb(ctx context.Context, s *grpcweb.WrappedGrpcServer, protocol, socket string, tlsCfg *tls.Config) error {
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

	lis, err := (&net.ListenConfig{}).Listen(ctx, protocol, socket)
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
