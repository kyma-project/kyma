package config_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	runtimeId     = "runtimeId"
	tenant        = "tenant"
	connectorURL  = "https://connector.com"
	token         = "token"
	configMapName = "compass-agent-configuration"
)

func TestProvider(t *testing.T) {

	configMapData := map[string]string{
		"CONNECTOR_URL": connectorURL,
		"TOKEN":         token,
		"TENANT":        tenant,
		"RUNTIME_ID":    runtimeId,
	}

	validConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: configMapName},
		Data:       configMapData,
	}

	crManager := &mocks.ConfigMapManager{}
	crManager.On("Get", configMapName, metav1.GetOptions{}).Return(validConfigMap, nil)
	configProvider := config.NewConfigProvider(configMapName, crManager)

	t.Run("should get Connection config", func(t *testing.T) {
		// when
		connectionConfig, err := configProvider.GetConnectionConfig()

		// then
		require.NoError(t, err)
		assert.Equal(t, connectorURL, connectionConfig.ConnectorURL)
		assert.Equal(t, token, connectionConfig.Token)

	})

	t.Run("should get Runtime config", func(t *testing.T) {
		// when
		runtimeConfig, err := configProvider.GetRuntimeConfig()

		// then
		require.NoError(t, err)
		assert.Equal(t, runtimeId, runtimeConfig.RuntimeId)
		assert.Equal(t, tenant, runtimeConfig.Tenant)
	})

}

func TestProvider_Errors(t *testing.T) {

	crManager := &mocks.ConfigMapManager{}
	crManager.On("Get", configMapName, metav1.GetOptions{}).Return(nil, errors.New("error"))
	configProvider := config.NewConfigProvider(configMapName, crManager)

	t.Run("should return error when failed to get config map for Connection config", func(t *testing.T) {
		// when
		connectionConfig, err := configProvider.GetConnectionConfig()

		// then
		require.Error(t, err)
		assert.Empty(t, connectionConfig)
	})

	t.Run("should return error when failed to get config map for Runtime config", func(t *testing.T) {
		// when
		runtimeConfig, err := configProvider.GetRuntimeConfig()

		// then
		require.Error(t, err)
		assert.Empty(t, runtimeConfig)
	})
}
