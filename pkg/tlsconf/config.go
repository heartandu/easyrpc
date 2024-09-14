package tlsconf

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

var (
	ErrMutualAuthParamsNotComplete = errors.New("cert and key are required for a mutual authentication")
	ErrAppendCACert                = errors.New("failed to append CA certificate")
)

// Config creates a TLS configuration based on the provided certificates and a key.
// Parameters:
// - cacert: path to the CA certificate file (optional),
// - cert: path to the client certificate file (optional),
// - key: path to the client certificate key file (optional).
func Config(cacert, cert, key string) (*tls.Config, error) {
	var tlsCfg tls.Config

	if cacert != "" {
		certBytes, err := os.ReadFile(cacert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(certBytes) {
			return nil, ErrAppendCACert
		}

		tlsCfg.RootCAs = cp
	}

	if cert != "" && key != "" {
		// Enable mutual authentication
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("failed to read the client certificate: %w", err)
		}

		tlsCfg.Certificates = append(tlsCfg.Certificates, certificate)
	} else if cert != "" || key != "" {
		return nil, ErrMutualAuthParamsNotComplete
	}

	return &tlsCfg, nil
}
