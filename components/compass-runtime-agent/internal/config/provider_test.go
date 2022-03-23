package config_test

import (
	"errors"
	"testing"

	mocks2 "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets/mocks"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	runtimeId    = "runtimeId"
	tenant       = "tenant"
	connectorURL = "https://connector.com"
	token        = "token"
	secretName   = "compass-agent-configuration"
)

var (
	secretNamespacedName = types.NamespacedName{
		Namespace: "compass-system",
		Name:      secretName,
	}
)

func TestProvider(t *testing.T) {

	configMapData := map[string][]byte{
		"CONNECTOR_URL": []byte(connectorURL),
		"TOKEN":         []byte(token),
		"TENANT":        []byte(tenant),
		"RUNTIME_ID":    []byte(runtimeId),
	}

	validConfigSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: "compass-system"},
		Data:       configMapData,
	}

	fakeClient := fake.NewSimpleClientset(validConfigSecret)
	secretsRepo := secrets.NewRepository(func(namespace string) secrets.Manager {
		return fakeClient.CoreV1().Secrets(namespace)
	})

	configProvider := config.NewConfigProvider(secretNamespacedName, secretsRepo)

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

	secretsRepo := &mocks2.Repository{}
	secretsRepo.On("Get", secretNamespacedName).Return(nil, errors.New("error"))
	configProvider := config.NewConfigProvider(secretNamespacedName, secretsRepo)

	t.Run("should return error when failed to get config map for Connection config", func(t *testing.T) {
		// when
		connectionConfig, err := configProvider.GetConnectionConfig()

		// then
		require.Error(t, err)
		assert.Empty(t, connectionConfig)
	})

	//t.Run("should return error when failed to get config map for Runtime config", func(t *testing.T) {
	//	// when
	//	runtimeConfig, err := configProvider.GetRuntimeConfig()
	//
	//	// then
	//	require.Error(t, err)
	//	assert.Empty(t, runtimeConfig)
	//})
}
