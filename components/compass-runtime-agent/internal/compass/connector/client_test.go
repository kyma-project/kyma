package connector

import (
	"testing"

	gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	gql "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kyma-project.io/compass-runtime-agent/internal/graphql"
)

const (
	TokenHeader = "Connector-Token"

	expectedConfigurationQuery = `query {
	result: configuration() {
		token { token }
		certificateSigningRequestInfo { subject keyAlgorithm }
		managementPlaneInfo { 
			directorURL
			certificateSecuredConnectorURL
		}
	}
}`

	expectedSigningQuery = `mutation {
	result: signCertificateSigningRequest(csr: "encodedCSR") {
		certificateChain
		caCertificate
		clientCertificate
	}
}`

	encodedCSR = "encodedCSR"

	token = "token"
)

var (
	connectorTokenHeaders map[string]string = map[string]string{TokenHeader: token}
)

func TestCertificateSecuredClient_Configuration(t *testing.T) {

	expectedRequest := gql.NewRequest(expectedConfigurationQuery)
	expectedRequest.Header.Set(TokenHeader, token)

	t.Run("should fetch configuration", func(t *testing.T) {
		// given
		expectedResponse := gqlschema.Configuration{
			Token: &gqlschema.Token{Token: "new-token"},
			CertificateSigningRequestInfo: &gqlschema.CertificateSigningRequestInfo{
				Subject:      "CN=app",
				KeyAlgorithm: "rsa2048",
			},
			ManagementPlaneInfo: &gqlschema.ManagementPlaneInfo{},
		}

		gqlClient := graphql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			config, ok := r.(*ConfigurationResponse)
			require.True(t, ok)
			assert.Empty(t, config)
			config.Result = expectedResponse
		}, expectedRequest)

		certSecuredClient := NewConnectorClient(gqlClient)

		// when
		configResponse, err := certSecuredClient.Configuration(connectorTokenHeaders)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, configResponse)
	})

	t.Run("should return error when failed to fetch config", func(t *testing.T) {
		// given
		gqlClient := graphql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			config, ok := r.(*ConfigurationResponse)
			require.True(t, ok)
			assert.Empty(t, config)
		}, expectedRequest)

		certSecuredClient := NewConnectorClient(gqlClient)

		// when
		configResponse, err := certSecuredClient.Configuration(connectorTokenHeaders)

		// then
		require.Error(t, err)
		assert.Empty(t, configResponse)
	})
}

func TestCertificateSecuredClient_SignCSR(t *testing.T) {

	expectedRequest := gql.NewRequest(expectedSigningQuery)
	expectedRequest.Header.Set(TokenHeader, token)

	t.Run("should sign csr", func(t *testing.T) {
		// given
		expectedResponse := gqlschema.CertificationResult{
			ClientCertificate: "clientCert",
			CertificateChain:  "certChain",
			CaCertificate:     "caCert",
		}

		gqlClient := graphql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			config, ok := r.(*CertificationResponse)
			require.True(t, ok)
			assert.Empty(t, config)
			config.Result = expectedResponse
		}, expectedRequest)

		certSecuredClient := NewConnectorClient(gqlClient)

		// when
		configResponse, err := certSecuredClient.SignCSR(encodedCSR, connectorTokenHeaders)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, configResponse)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		gqlClient := graphql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			config, ok := r.(*CertificationResponse)
			require.True(t, ok)
			assert.Empty(t, config)
		}, expectedRequest)

		certSecuredClient := NewConnectorClient(gqlClient)

		// when
		configResponse, err := certSecuredClient.SignCSR(encodedCSR, connectorTokenHeaders)

		// then
		require.Error(t, err)
		assert.Empty(t, configResponse)
	})
}
