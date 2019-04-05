package connectorservice

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
)

type MutualTLSClientProvider interface {
	CreateClient(key *rsa.PrivateKey, clientCert *x509.Certificate) MutualTLSConnectorClient
}

type mutualTLSClientProvider struct {
	csrProvider certificates.CSRProvider
}

func NewMutualTLSClientProvider(csrProvider certificates.CSRProvider) MutualTLSClientProvider {
	return &mutualTLSClientProvider{
		csrProvider: csrProvider,
	}
}

func (cp *mutualTLSClientProvider) CreateClient(key *rsa.PrivateKey, clientCert *x509.Certificate) MutualTLSConnectorClient {
	certs := []tls.Certificate{
		{
			PrivateKey:  key,
			Certificate: [][]byte{clientCert.Raw},
		},
	}

	tlsConfig := &tls.Config{
		Certificates: certs,
	}

	return NewMutualTLSConnectorClient(tlsConfig, cp.csrProvider, clientCert.Subject)
}
