package compassruntimeagentinit

import (
	"context"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestCertificateSecrets(t *testing.T) {
	t.Run("should return rollback function that will remove secrets", func(t *testing.T) {
		// given
		fakeKubernetesInterface := fake.NewSimpleClientset()

		// when
		configurator := NewCertificateSecretConfigurator(fakeKubernetesInterface, "test")
		rollbackFunc, err := configurator.Do("newCaSecret", "newClientSetSecret")

		// then
		require.NoError(t, err)

		// given
		caCertSecret := createSecret("newCaSecret", "test")
		clientCertSecret := createSecret("newClientSetSecret", "test")

		_, err = fakeKubernetesInterface.CoreV1().Secrets("test").Create(context.TODO(), caCertSecret, meta.CreateOptions{})
		require.NoError(t, err)

		_, err = fakeKubernetesInterface.CoreV1().Secrets("test").Create(context.TODO(), clientCertSecret, meta.CreateOptions{})
		require.NoError(t, err)

		// when
		err = rollbackFunc()
		require.NoError(t, err)

		// then
		_, err = fakeKubernetesInterface.CoreV1().Secrets("test").Get(context.TODO(), "newCaSecret", meta.GetOptions{})
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))

		_, err = fakeKubernetesInterface.CoreV1().Secrets("test").Get(context.TODO(), "newClientSetSecret", meta.GetOptions{})
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))
	})

	t.Run("should return error when rollback function failed", func(t *testing.T) {
		// given
		fakeKubernetesInterface := fake.NewSimpleClientset()

		// when
		configurator := NewCertificateSecretConfigurator(fakeKubernetesInterface, "test")
		rollbackFunc, err := configurator.Do("newCaSecret", "newClientSetSecret")

		// then
		require.NoError(t, err)

		// when
		err = rollbackFunc()
		require.Error(t, err)
	})
}

func createSecret(name, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
	}
}
