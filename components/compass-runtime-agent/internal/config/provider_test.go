package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	runtimeId    = "runtimeId"
	tenant       = "tenant"
	connectorURL = "https://connector.com"
	token        = "token"
)

func TestProvider(t *testing.T) {

	for _, test := range []struct {
		description          string
		configFilePath       string
		expectedConnectorURL string
		expectedToken        string
		expectedTenant       string
		expectedRuntimeId    string
	}{
		{
			description:          "get valid config",
			configFilePath:       validConfigPath,
			expectedConnectorURL: connectorURL,
			expectedToken:        token,
			expectedTenant:       tenant,
			expectedRuntimeId:    runtimeId,
		},
		{
			description:          "return error empty values when config is invalid",
			configFilePath:       invalidConfigPath,
			expectedConnectorURL: "",
			expectedToken:        "",
			expectedTenant:       "",
			expectedRuntimeId:    "",
		},
	} {
		t.Run("should "+test.description, func(t *testing.T) {
			// given
			provider := NewConfigProvider(test.configFilePath)

			// when
			connectionConfig, connErr := provider.GetConnectionConfig()
			runtimeConfig, runtimeErr := provider.GetRuntimeConfig()

			// then
			require.NoError(t, connErr)
			require.NoError(t, runtimeErr)

			assert.Equal(t, test.expectedToken, connectionConfig.Token)
			assert.Equal(t, test.expectedConnectorURL, connectionConfig.ConnectorURL)
			assert.Equal(t, test.expectedRuntimeId, runtimeConfig.RuntimeId)
			assert.Equal(t, test.expectedTenant, runtimeConfig.Tenant)

		})
	}

	t.Run("should return error when config format is invalid", func(t *testing.T) {
		// given
		provider := NewConfigProvider(invalidConfigFormatPath)

		// when
		connectionConfig, connErr := provider.GetConnectionConfig()
		runtimeConfig, runtimeErr := provider.GetRuntimeConfig()

		// then
		require.Error(t, connErr)
		require.Error(t, runtimeErr)

		assert.Empty(t, connectionConfig)
		assert.Empty(t, runtimeConfig)
	})

	t.Run("should return error when config file is empty", func(t *testing.T) {
		// given
		provider := NewConfigProvider(emptyFilePath)

		// when
		connectionConfig, connErr := provider.GetConnectionConfig()
		runtimeConfig, runtimeErr := provider.GetRuntimeConfig()

		// then
		require.Error(t, connErr)
		require.Error(t, runtimeErr)

		assert.Empty(t, connectionConfig)
		assert.Empty(t, runtimeConfig)
	})

	t.Run("should return error when file does not exist", func(t *testing.T) {
		// given
		provider := NewConfigProvider("not_existing_file")

		// when
		connectionConfig, connErr := provider.GetConnectionConfig()
		runtimeConfig, runtimeErr := provider.GetRuntimeConfig()

		// then
		require.Error(t, connErr)
		require.Error(t, runtimeErr)

		assert.Empty(t, connectionConfig)
		assert.Empty(t, runtimeConfig)
	})
}
