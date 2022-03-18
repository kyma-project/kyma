package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testSecretName    = "test-secret"
	testNamespaceName = "test-namespace"
	testServiceName   = "test-service"
)

func Test_serviceAltNames(t *testing.T) {
	type args struct {
		serviceName string
		namespace   string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "service AltNames are generated correctly",
			args: args{serviceName: "test-service", namespace: "test-namespace"},
			// not using consts here to make it as readable as possible.
			want: []string{
				"test-service.test-namespace.svc",
				"test-service",
				"test-service.test-namespace",
				"test-service.test-namespace.svc.cluster.local",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serviceAltNames(tt.args.serviceName, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("invalid serviec altNames: serviceAltNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureWebhookSecret(t *testing.T) {

	ctx := context.Background()

	t.Run("can ensure the secret if it doesn't exist", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()

		err := EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		secret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, secret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, secret.Name, testSecretName)
		require.Equal(t, secret.Namespace, testNamespaceName)
		require.Contains(t, secret.Data, KeyFile)
		require.Contains(t, secret.Data, CertFile)
	})

	t.Run("can ensure the secret is updated if it exists", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespaceName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
		}
		err := client.Create(ctx, secret)
		require.NoError(t, err)

		err = EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		updatedSecret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, updatedSecret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, updatedSecret.Name, testSecretName)
		require.Equal(t, updatedSecret.Namespace, testNamespaceName)
		require.Contains(t, updatedSecret.Data, KeyFile)
		require.Contains(t, updatedSecret.Data, CertFile)
		require.Contains(t, updatedSecret.Labels, "dont-remove-me")
	})
}
