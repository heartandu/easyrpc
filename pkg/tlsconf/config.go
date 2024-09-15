package tlsconf

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/spf13/afero"
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
func Config(fs afero.Fs, cacert, cert, key string) (*tls.Config, error) {
	var tlsCfg tls.Config

	if cacert != "" {
		certBytes, err := afero.ReadFile(fs, cacert)
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
		certificate, err := readX509KeyPair(fs, cert, key)
		if err != nil {
			return nil, fmt.Errorf("failed to read x509 key pair: %w", err)
		}

		tlsCfg.Certificates = append(tlsCfg.Certificates, certificate)
	} else if cert != "" || key != "" {
		return nil, ErrMutualAuthParamsNotComplete
	}

	return &tlsCfg, nil
}

func readX509KeyPair(fs afero.Fs, cert, key string) (tls.Certificate, error) {
	certPEMBlock, err := afero.ReadFile(fs, cert)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	keyPEMBlock, err := afero.ReadFile(fs, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read key: %w", err)
	}

	certificate, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read the client certificate: %w", err)
	}

	return certificate, nil
}
