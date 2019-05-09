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

type EstablishedConnectionClientProvider interface {
	CreateClient(credentials CertificateCredentials) EstablishedConnectionClient
}

type establishedConnectionClientProvider struct {
	csrProvider certificates.CSRProvider
}

func NewEstablishedConnectionClientProvider(csrProvider certificates.CSRProvider) EstablishedConnectionClientProvider {
	return &establishedConnectionClientProvider{
		csrProvider: csrProvider,
	}
}

func (cp *establishedConnectionClientProvider) CreateClient(credentials CertificateCredentials) EstablishedConnectionClient {

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

	return NewEstablishedConnectionClient(tlsConfig, cp.csrProvider, credentials.ClientCert.Subject)
}
