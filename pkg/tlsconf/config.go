package tlsconf

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

var (
	ErrMutualAuthParamsNotComplete = errors.New("cert and certkey are required to authenticate mutually")
	ErrAppendCerts                 = errors.New("failed to append the client certificate")
)

func Config(cacert, cert, certKey string) (*tls.Config, error) {
	var tlsCfg tls.Config

	if cacert != "" {
		certBytes, err := os.ReadFile(cacert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(certBytes) {
			return nil, ErrAppendCerts
		}

		tlsCfg.RootCAs = cp
	}

	if cert != "" && certKey != "" {
		// Enable mutual authentication
		certificate, err := tls.LoadX509KeyPair(cert, certKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read the client certificate: %w", err)
		}

		tlsCfg.Certificates = append(tlsCfg.Certificates, certificate)
	} else if cert != "" || certKey != "" {
		return nil, ErrMutualAuthParamsNotComplete
	}

	return &tlsCfg, nil
}
