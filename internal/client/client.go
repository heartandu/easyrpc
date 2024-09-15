package client

import (
	"crypto/tls"
	"fmt"

	"github.com/heartandu/grpc-web-go-client/grpcweb"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/pkg/conn"
	"github.com/heartandu/easyrpc/pkg/tlsconf"
)

// New creates a new gRPC client connection based on the provided configuration.
// It checks the configuration to determine whether to establish a gRPC or gRPC-Web connection.
func New(fs afero.Fs, cfg *config.Config) (grpc.ClientConnInterface, error) {
	if cfg.Server.Web {
		return clientWebConn(fs, cfg)
	}

	return clientGRPCConn(fs, cfg)
}

// clientGRPCConn creates a new gRPC client connection.
// It handles the creation of the gRPC connection with or without TLS.
func clientGRPCConn(fs afero.Fs, cfg *config.Config) (*grpc.ClientConn, error) {
	creds := insecure.NewCredentials()

	if cfg.TLS.Enabled {
		conf, err := tlsConfig(fs, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get tls config: %w", err)
		}

		creds = credentials.NewTLS(conf)
	}

	clientConn, err := grpc.NewClient(cfg.Server.Address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return clientConn, nil
}

// clientWebConn creates a new gRPC-Web client connection.
// It handles the creation of the gRPC-Web connection with or without TLS.
func clientWebConn(fs afero.Fs, cfg *config.Config) (*conn.WebClient, error) {
	credsOpt := grpcweb.WithInsecure()

	if cfg.TLS.Enabled {
		conf, err := tlsConfig(fs, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get tls config: %w", err)
		}

		credsOpt = grpcweb.WithTLSConfig(conf)
	}

	cc, err := grpcweb.NewClient(cfg.Server.Address, credsOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to dial context: %w", err)
	}

	return conn.NewWebClient(cc), nil
}

// tlsConfig creates a TLS configuration.
// It reads the TLS certificates and keys from the file system and constructs a TLS configuration.
func tlsConfig(fs afero.Fs, cfg *config.Config) (*tls.Config, error) {
	conf, err := tlsconf.Config(fs, cfg.TLS.CACert, cfg.TLS.Cert, cfg.TLS.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to make tls config: %w", err)
	}

	return conf, nil
}
