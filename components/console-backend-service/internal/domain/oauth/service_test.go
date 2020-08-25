package oauth

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/stretchr/testify/assert"

	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
)

func TestOAuthService_Query(t *testing.T) {
	const namespace = "default"

	t.Run("Should filter by namespace", func(t *testing.T) {
		client1 := createMockClient("client 1", namespace)
		client2 := createMockClient("client 2", "other-namespace")
		service := setupServiceWithClients(t, client1, client2)

		result, err := service.OAuth2ClientsQuery(context.Background(), namespace)

		require.NoError(t, err)

		assert.Equal(t, 1, len(result))
		assert.Equal(t, client1.Namespace, result[0].Namespace)
	})

	t.Run("Should find client by name", func(t *testing.T) {
		client1 := createMockClient("client 1", namespace)
		client2 := createMockClient("client 2", namespace)
		service := setupServiceWithClients(t, client1, client2)

		result, err := service.OAuth2ClientQuery(context.Background(), "client 1", namespace)

		require.NoError(t, err)

		assert.Equal(t, client1.Name, result.Name)
	})

	t.Run("Should return error if client is not found", func(t *testing.T) {
		client1 := createMockClient("client 1", namespace)
		service := setupServiceWithClients(t, client1)

		_, err := service.OAuth2ClientQuery(context.Background(), "client 2", namespace)

		require.Error(t, err)
	})
}

func setupServiceWithClients(t *testing.T, clients ...*v1alpha1.OAuth2Client) *Resolver {
	// NewFakeGenericServiceFactory requires array of runtime.Object
	objects := make([]runtime.Object, len(clients))
	for i, client := range clients {
		objects[i] = client
	}

	serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, objects...)
	require.NoError(t, err)

	service := New(serviceFactory)
	err = service.Enable()
	require.NoError(t, err)

	serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

	return service
}

func createMockClient(name, namespace string) *v1alpha1.OAuth2Client {
	return &v1alpha1.OAuth2Client{
		TypeMeta: v1.TypeMeta{
			APIVersion: "hydra.ory.sh/v1alpha1",
			Kind:       "OAuth2Client",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.OAuth2ClientSpec{
			Scope:      "read write",
			SecretName: "secret-name",
		},
	}
}

func createMockClientSpec() v1alpha1.OAuth2ClientSpec {
	return v1alpha1.OAuth2ClientSpec{
		GrantTypes:    []v1alpha1.GrantType{"client_credentials"},
		ResponseTypes: []v1alpha1.ResponseType{"id_token"},
		Scope:         "read write",
		SecretName:    "secret-name",
	}
}
