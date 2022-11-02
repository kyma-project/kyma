package init

import (
	"context"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestConfigurationSecret(t *testing.T) {
	t.Run("should create configuration secret", func(t *testing.T) {
		// given
		fakeKubernetesInterface := fake.NewSimpleClientset()
		secretConfigurator := NewConfigurationSecretConfigurator(fakeKubernetesInterface)
		connectorURL := "www.example.com"
		runtimeID := "runtimeID"
		token := "token"
		tenant := "tenant"

		config := types.CompassRuntimeAgentConfig{
			ConnectorUrl: connectorURL,
			RuntimeID:    runtimeID,
			Token:        token,
			Tenant:       tenant,
		}
		secretName := "config"

		// when
		rollbackFunc, err := secretConfigurator.Do(secretName, config)
		require.NotNil(t, rollbackFunc)
		require.NoError(t, err)

		// then
		secret, err := fakeKubernetesInterface.CoreV1().Secrets(CompassSystemNamespace).Get(context.TODO(), secretName, meta.GetOptions{})
		require.NoError(t, err)

		require.Equal(t, connectorURL, string(secret.Data[connectorURLConfigKey]))
		require.Equal(t, token, string(secret.Data[tokenConfigKey]))
		require.Equal(t, runtimeID, string(secret.Data[runtimeIdConfigKey]))
		require.Equal(t, tenant, string(secret.Data[tenantConfigKey]))

		// when
		err = rollbackFunc()
		require.NoError(t, err)

		_, err = fakeKubernetesInterface.CoreV1().Secrets(CompassSystemNamespace).Get(context.TODO(), secretName, meta.GetOptions{})
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))
	})

	t.Run("should return error when failed to create secret", func(t *testing.T) {
		// given
		fakeKubernetesInterface := fake.NewSimpleClientset()
		secretConfigurator := NewConfigurationSecretConfigurator(fakeKubernetesInterface)

		config := types.CompassRuntimeAgentConfig{}
		secretName := "config"

		// when
		secret := createSecret(secretName, CompassSystemNamespace)
		_, err := fakeKubernetesInterface.CoreV1().Secrets(CompassSystemNamespace).Create(context.Background(), secret, meta.CreateOptions{})
		require.NoError(t, err)

		rollbackFunc, err := secretConfigurator.Do(secretName, config)

		// then
		require.Nil(t, rollbackFunc)
		require.Error(t, err)
	})
}
