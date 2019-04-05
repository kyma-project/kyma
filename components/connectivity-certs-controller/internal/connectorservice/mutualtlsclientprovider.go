package connectorservice

import (
	"crypto/tls"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
)

type MutualTLSClientProvider interface {
	CreateClient() (MutualTLSConnectorClient, error)
}

type mutualTLSClientProvider struct {
	certificateProvider certificates.Provider
	csrProvider         certificates.CSRProvider
}

func NewMutualTLSClientProvider(csrProvider certificates.CSRProvider, certProvider certificates.Provider) MutualTLSClientProvider {
	return &mutualTLSClientProvider{
		csrProvider:         csrProvider,
		certificateProvider: certProvider,
	}
}

func (cp *mutualTLSClientProvider) CreateClient() (MutualTLSConnectorClient, error) {
	key, certificate, err := cp.certificateProvider.GetClientCredentials()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read client certificate and key")
	}

	certs := []tls.Certificate{
		{
			PrivateKey:  key,
			Certificate: [][]byte{certificate.Raw},
		},
	}

	tlsConfig := &tls.Config{
		Certificates: certs,
	}

	return NewMutualTLSConnectorClient(tlsConfig, cp.csrProvider, certificate.Subject), nil
}
