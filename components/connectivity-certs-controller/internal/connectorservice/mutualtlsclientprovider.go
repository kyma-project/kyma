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
	CACerts    []*x509.Certificate
}

type MutualTLSClientProvider interface {
	CreateClient(credentials CertificateCredentials) MutualTLSClient
}

type mutualTLSClientProvider struct {
	csrProvider certificates.CSRProvider
}

func NewMutualTLSClientProvider(csrProvider certificates.CSRProvider) MutualTLSClientProvider {
	return &mutualTLSClientProvider{
		csrProvider: csrProvider,
	}
}

func (cp *mutualTLSClientProvider) CreateClient(credentials CertificateCredentials) MutualTLSClient {

	rawCerts := [][]byte{credentials.ClientCert.Raw}

	for _, cert := range credentials.CACerts {
		rawCerts = append(rawCerts, cert.Raw)
	}

	certs := []tls.Certificate{
		{
			PrivateKey:  credentials.ClientKey,
			Certificate: rawCerts,
		},
	}

	tlsConfig := &tls.Config{
		Certificates: certs,
	}

	return NewMutualTLSConnectorClient(tlsConfig, cp.csrProvider, credentials.ClientCert.Subject)
}
