package applications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

func TestCompare(t *testing.T) {

	t.Run("should return true if secrets are equal", func(t *testing.T) {
		//given
		coreV1 := fake.NewSimpleClientset()
		secretComparator, err := NewSecretComparator(coreV1, "test", "kyma-integration")
		require.NoError(t, err)
		createFakeCredentialsSecret(t, coreV1.CoreV1().Secrets("test"), "expected", "test")
		createFakeCredentialsSecret(t, coreV1.CoreV1().Secrets("kyma-integration"), "actual", "kyma-integration")

		//when
		err = secretComparator.Compare(t, "expected", "actual")

		// then
		require.NoError(t, err)
	})

	t.Run("should return error if failed to read actual secret", func(t *testing.T) {
		//given
		coreV1 := fake.NewSimpleClientset()
		secretComparator, err := NewSecretComparator(coreV1, "test", "kyma-integration")
		require.NoError(t, err)
		createFakeCredentialsSecret(t, coreV1.CoreV1().Secrets("test"), "expected", "test")

		//when
		err = secretComparator.Compare(t, "actual", "expected")

		// then
		require.Error(t, err)
	})

	t.Run("should return error if failed to read expected secret", func(t *testing.T) {
		//given
		coreV1 := fake.NewSimpleClientset()
		secretComparator, err := NewSecretComparator(coreV1, "test", "kyma-integration")
		require.NoError(t, err)
		createFakeCredentialsSecret(t, coreV1.CoreV1().Secrets("kyma-integration"), "actual", "kyma-integration")

		//when
		err = secretComparator.Compare(t, "actual", "expected")

		// then
		require.Error(t, err)
	})

	t.Run("should return error if expected secret name is empty", func(t *testing.T) {
		//given
		secretComparator, err := NewSecretComparator(nil, "test", "kyma-integration")
		require.NoError(t, err)

		//when
		err = secretComparator.Compare(t, "actual", "")

		// then
		require.Error(t, err)
	})

	t.Run("should return error if actual secret name is empty", func(t *testing.T) {
		//given
		secretComparator, err := NewSecretComparator(nil, "test", "kyma-integration")
		require.NoError(t, err)

		//when
		err = secretComparator.Compare(t, "", "expected")

		// then
		require.Error(t, err)
	})

	t.Run("should return no error if actual and expected secret name is empty", func(t *testing.T) {
		//given
		secretComparator, err := NewSecretComparator(nil, "test", "kyma-integration")
		require.NoError(t, err)

		//when
		err = secretComparator.Compare(t, "", "")

		// then
		require.NoError(t, err)
	})
}

func createFakeCredentialsSecret(t *testing.T, secrets core.SecretInterface, secretName, namespace string) {

	secret := &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			"key1": []byte("val1"),
			"key2": []byte("val2"),
			"key3": []byte("val3"),
		},
	}

	_, err := secrets.Create(context.Background(), secret, meta.CreateOptions{})

	require.NoError(t, err)
}
