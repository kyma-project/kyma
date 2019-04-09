package connectorservice

import (
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates/mocks"

	"github.com/stretchr/testify/require"
)

func TestMutualTLSClientProvider_CreateClient(t *testing.T) {

	t.Run("should create Mutual TLS Connector Client", func(t *testing.T) {
		// given
		credentials := CertificateCredentials{
			ClientKey:  &rsa.PrivateKey{},
			ClientCert: &x509.Certificate{Subject: subject},
			CACert:     &x509.Certificate{},
		}

		csrProvider := &mocks.CSRProvider{}

		clientProvider := NewMutualTLSClientProvider(csrProvider)

		// when
		client := clientProvider.CreateClient(credentials)

		// then
		require.NotNil(t, client)
	})
}
