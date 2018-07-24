package k8s_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func TestSecretResolver(t *testing.T) {
	// GIVEN
	fakeClientSet := fake.NewSimpleClientset(
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-secret",
				Namespace: "production",
			},
		})
	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
	// WHEN
	actualSecret, err := resolver.SecretQuery(context.Background(), "my-secret", "production")
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "my-secret", actualSecret.Name)
	assert.Equal(t, "production", actualSecret.Environment)
}

func TestSecretResolverOnNotFound(t *testing.T) {
	// GIVEN
	fakeClientSet := fake.NewSimpleClientset()
	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
	// WHEN
	secret, err := resolver.SecretQuery(context.Background(), "my-secret", "production")
	// THEN
	assert.NoError(t, err)
	assert.Nil(t, secret)
}

func TestSecretResolverOnError(t *testing.T) {
	// GIVEN
	fakeClientSet := fake.NewSimpleClientset()
	fakeClientSet.PrependReactor("get", "secrets", failingReactor)
	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
	// WHEN
	_, err := resolver.SecretQuery(context.Background(), "my-secret", "production")
	// THEN
	assert.EqualError(t, err, "cannot get Secret")
}

func failingReactor(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("custom error")
}
