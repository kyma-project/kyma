package connector

import (
	"testing"

	gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql/mocks"
	gql "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

	setExpectedConfigFunc := func(config gqlschema.Configuration) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			response, ok := args[1].(*ConfigurationResponse)
			require.True(t, ok)
			assert.Empty(t, response)
			response.Result = config
		}
	}

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

		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ConfigurationResponse{}).
			Return(nil).
			Run(setExpectedConfigFunc(expectedResponse)).
			Once()

		certSecuredClient := NewConnectorClient(client)

		// when
		configResponse, err := certSecuredClient.Configuration(connectorTokenHeaders)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, configResponse)
	})

	t.Run("should return error when failed to fetch config", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &ConfigurationResponse{}).
			Return(errors.New("error"))

		certSecuredClient := NewConnectorClient(client)

		// when
		configResponse, err := certSecuredClient.Configuration(connectorTokenHeaders)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to get configuration")
		assert.Empty(t, configResponse)
	})
}

func TestCertificateSecuredClient_SignCSR(t *testing.T) {
	expectedRequest := gql.NewRequest(expectedSigningQuery)
	expectedRequest.Header.Set(TokenHeader, token)

	setExpectedCertFunc := func(cert gqlschema.CertificationResult) func(args mock.Arguments) {
		return func(args mock.Arguments) {
			response, ok := args[1].(*CertificationResponse)
			require.True(t, ok)
			assert.Empty(t, response)
			response.Result = cert
		}
	}

	t.Run("should sign csr", func(t *testing.T) {
		// given
		expectedResponse := gqlschema.CertificationResult{
			ClientCertificate: "clientCert",
			CertificateChain:  "certChain",
			CaCertificate:     "caCert",
		}

		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &CertificationResponse{}).
			Return(nil).
			Run(setExpectedCertFunc(expectedResponse)).
			Once()

		certSecuredClient := NewConnectorClient(client)

		// when
		configResponse, err := certSecuredClient.SignCSR(encodedCSR, connectorTokenHeaders)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, configResponse)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.
			On("Do", expectedRequest, &CertificationResponse{}).
			Return(errors.New("error"))

		certSecuredClient := NewConnectorClient(client)

		// when
		configResponse, err := certSecuredClient.SignCSR(encodedCSR, connectorTokenHeaders)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to generate certificate")
		assert.Empty(t, configResponse)
	})
}
