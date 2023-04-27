package resources

import (
	"context"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Test_resourceReconciler_Reconcile(t *testing.T) {
	fakeLogger := zap.NewNop().Sugar()
	t.Run("should not reconcile not owned resources", func(t *testing.T) {
		ctx := context.Background()
		client := fake.NewClientBuilder().Build()
		namespacedName := types.NamespacedName{Namespace: "", Name: DefaultingWebhookName}
		webhookConfig := WebhookConfig{
			CABundle:         []byte("certificate content"),
			ServiceName:      "test-webhook-service",
			ServiceNamespace: "test-namespace",
		}
		r := &resourceReconciler{
			webhookConfig: webhookConfig,
			secretName:    "test-secret-name",
			client:        client,
			logger:        fakeLogger,
		}
		err := createTestResources(ctx, client)
		require.NoError(t, err)

		_, err = r.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
		require.Error(t, err)
		require.True(t, k8serrors.IsNotFound(err))

		for _, res := range getTestResources() {
			r := res
			err := client.Get(ctx, types.NamespacedName{Name: res.GetName(), Namespace: res.GetNamespace()}, r)
			require.NoError(t, err)
			require.Equal(t, r.GetResourceVersion(), "1")
		}
	})

}

func createTestResources(ctx context.Context, client ctrlclient.Client) error {
	resources := getTestResources()
	for _, res := range resources {
		r := res
		err := client.Create(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func getTestResources() []ctrlclient.Object {
	return []ctrlclient.Object{
		&admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-my-mutationwebhookconfig",
			}},
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-my-validationwebhookconfig",
			}},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "not-my-secret",
				Namespace: "default",
			},
		},
	}
}
