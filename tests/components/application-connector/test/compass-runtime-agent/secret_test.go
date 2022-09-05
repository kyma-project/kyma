package compass_runtime_agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"testing"
)

func TestCompare(t *testing.T) {

	t.Run("should return false if can't get actual secret from cluster", func(t *testing.T) {
		//given
		coreV1 := fake.NewSimpleClientset()
		secretExpected := coreV1.CoreV1().Secrets("test")
		createFakeCredentialsSecret(t, secretExpected, "secret-expected", "test")

		//when
		secretManager := NewClient(secretExpected, "secret-expected")

		//then
		secretEqual := secretManager.Compare("secret-actual", "secret-expected", coreV1)
		assert.Equal(t, false, secretEqual)
	})

	t.Run("should return false if can't get expected secret from cluster", func(t *testing.T) {

	})

	t.Run("should return false if secrets aren't equal", func(t *testing.T) {

	})

	t.Run("should return true if secrets are equal", func(t *testing.T) {

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

	assert.NoError(t, err) //or require
}
