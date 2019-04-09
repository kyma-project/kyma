package connectorservice

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
)

type CertificateCredentials struct {
	ClientKey  *rsa.PrivateKey
	ClientCert *x509.Certificate
	CACert     *x509.Certificate
}

type MutualTLSClientProvider interface {
	CreateClient(credentials CertificateCredentials) MutualTLSConnectorClient
}

type mutualTLSClientProvider struct {
	csrProvider certificates.CSRProvider
}

func NewMutualTLSClientProvider(csrProvider certificates.CSRProvider) MutualTLSClientProvider {
	return &mutualTLSClientProvider{
		csrProvider: csrProvider,
	}
}

func (cp *mutualTLSClientProvider) CreateClient(credentials CertificateCredentials) MutualTLSConnectorClient {

	certs := []tls.Certificate{
		{
			PrivateKey:  credentials.ClientKey,
			Certificate: [][]byte{credentials.ClientCert.Raw, credentials.CACert.Raw},
		},
	}

	tlsConfig := &tls.Config{
		Certificates: certs,
	}

	return NewMutualTLSConnectorClient(tlsConfig, cp.csrProvider, credentials.ClientCert.Subject)
}
