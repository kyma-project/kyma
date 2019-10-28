package compass

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"testing"

	"kyma-project.io/compass-runtime-agent/internal/config"

	"github.com/stretchr/testify/require"

	"kyma-project.io/compass-runtime-agent/internal/certificates"

	"github.com/stretchr/testify/assert"
	"kyma-project.io/compass-runtime-agent/internal/graphql/mocks"

	"kyma-project.io/compass-runtime-agent/internal/graphql"
)

func newMockGQLConstructor(
	t *testing.T,
	returnedError error,
	expectedCert *tls.Certificate,
	expectedEndpoint string,
	expectedLogging bool,
	expectedInsecureConfigFetch bool) graphql.ClientConstructor {
	return func(certificate *tls.Certificate, graphqlEndpoint string, enableLogging bool, insecureConfigFetch bool) (graphql.Client, error) {
		assert.Equal(t, expectedCert, certificate)
		assert.Equal(t, expectedEndpoint, graphqlEndpoint)
		assert.Equal(t, expectedLogging, enableLogging)
		assert.Equal(t, expectedInsecureConfigFetch, insecureConfigFetch)

		return &mocks.Client{}, returnedError
	}
}

func TestClientsProvider_GetCompassConfigClient(t *testing.T) {

	runtimeConfig := config.RuntimeConfig{
		RuntimeId: "runtimeId",
		Tenant:    "tenant",
	}

	credentials := certificates.ClientCredentials{
		ClientKey:         &rsa.PrivateKey{},
		CertificateChain:  []*x509.Certificate{},
		ClientCertificate: &x509.Certificate{},
	}

	tlsCert := credentials.AsTLSCertificate()
	url := "http://api.io"
	enableLogging := true
	insecureFetch := true

	t.Run("should create new Compass Config Client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, nil, tlsCert, url, enableLogging, insecureFetch)

		provider := NewClientsProvider(constructor, false, insecureFetch, enableLogging)

		// when
		configClient, err := provider.GetDirectorClient(credentials, url, runtimeConfig)

		// then
		require.NoError(t, err)
		assert.NotNil(t, configClient)
	})

	t.Run("should return error when failed to create GraphQL client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, errors.New("error"), tlsCert, url, enableLogging, insecureFetch)

		provider := NewClientsProvider(constructor, false, insecureFetch, enableLogging)

		// when
		_, err := provider.GetDirectorClient(credentials, url, runtimeConfig)

		// then
		require.Error(t, err)
	})

}

func TestClientsProvider_GetConnectorTokenSecuredClient(t *testing.T) {

	url := "http://api.io"
	enableLogging := true
	insecureFetch := true

	t.Run("should create new Connector token-secured Client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, nil, nil, url, enableLogging, insecureFetch)

		provider := NewClientsProvider(constructor, insecureFetch, false, enableLogging)

		// when
		configClient, err := provider.GetConnectorClient(url)

		// then
		require.NoError(t, err)
		assert.NotNil(t, configClient)
	})

	t.Run("should return error when failed to create GraphQL client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, errors.New("error"), nil, url, enableLogging, insecureFetch)

		provider := NewClientsProvider(constructor, insecureFetch, false, enableLogging)

		// when
		_, err := provider.GetConnectorClient(url)

		// then
		require.Error(t, err)
	})
}

func TestClientsProvider_GetConnectorCertSecuredClient(t *testing.T) {

	credentials := certificates.ClientCredentials{
		ClientKey:         &rsa.PrivateKey{},
		CertificateChain:  []*x509.Certificate{},
		ClientCertificate: &x509.Certificate{},
	}

	tlsCert := credentials.AsTLSCertificate()
	url := "http://api.io"
	enableLogging := true
	insecureFetch := true

	t.Run("should create new Connector cert-secured Client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, nil, tlsCert, url, enableLogging, insecureFetch)

		provider := NewClientsProvider(constructor, insecureFetch, false, enableLogging)

		// when
		configClient, err := provider.GetConnectorCertSecuredClient(credentials, url)

		// then
		require.NoError(t, err)
		assert.NotNil(t, configClient)
	})

	t.Run("should return error when failed to create GraphQL client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, errors.New("error"), tlsCert, url, enableLogging, insecureFetch)

		provider := NewClientsProvider(constructor, insecureFetch, false, enableLogging)

		// when
		_, err := provider.GetConnectorCertSecuredClient(credentials, url)

		// then
		require.Error(t, err)
	})

}
