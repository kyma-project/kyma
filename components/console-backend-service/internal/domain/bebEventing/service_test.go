package bebEventing

import (
	"context"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicFake "k8s.io/client-go/dynamic/fake"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBebEventingService_Query(t *testing.T) {
	t.Run("Should filter by namespace and owner name", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")
		subscription2 := fixTestSubscription("test-sub-2", "default", "owner-2")

		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1, subscription2)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.EventSubscriptionsQuery(context.Background(), "owner-2", "default")

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha1.Subscription{
			subscription2,
		}, result)
	})

	t.Run("Should return empty array if not found", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")

		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.EventSubscriptionsQuery(context.Background(), "owner-1", "custom")

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestBebEventingService_Create(t *testing.T) {
	t.Run("Should create", func(t *testing.T) {
		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		params := fixTestSpecInput("owner-name")

		result, err := service.CreateEventSubscription(context.Background(), "default", "test-sub-1", params)

		require.NoError(t, err)
		require.Equal(t, "owner-name", result.ObjectMeta.OwnerReferences[0].Name)
		require.Equal(t, "default", result.ObjectMeta.Namespace)
		require.Equal(t, "test-sub-1", result.ObjectMeta.Name)
		require.Equal(t, "sap.kyma.custom.app.test-passed.v1", result.Spec.Filter.Filters[0].EventType.Value)
		require.Equal(t, "test-secret-data", result.Spec.Filter.Filters[0].EventSource.Value)
	})

	t.Run("Should return error for duplicates", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")

		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		params := fixTestSpecInput("owner-name")

		_, err = service.CreateEventSubscription(context.Background(), "default", "test-sub-1", params)

		require.Error(t, err)
	})
}

func TestBebEventingService_Update(t *testing.T) {
	t.Run("Should update", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")
		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		updatedFilters := []*gqlschema.FiltersInput{
			{
				ApplicationName: "app-2",
				Version:         "v2",
				EventName:       "test",
			},
		}
		params := fixTestSpecInput("owner-1")
		params.Filters = updatedFilters

		result, err := service.UpdateEventSubscription(context.Background(), "default", "test-sub-1", params)

		require.NoError(t, err)

		filter := result.Spec.Filter.Filters[0]
		assert.Equal(t, "sap.kyma.custom.app-2.test.v2", filter.EventType.Value)
	})

	t.Run("Should return error if not found", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")

		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		params := fixTestSpecInput("owner-name")

		_, err = service.UpdateEventSubscription(context.Background(), "default", "test-sub-2", params)

		require.Error(t, err)
	})
}

func TestBebEventingService_Delete(t *testing.T) {
	t.Run("Should delete", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")
		subscription2 := fixTestSubscription("test-sub-2", "default", "owner-2")

		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1, subscription2)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.DeleteEventSubscription(context.Background(), "default", "test-sub-2")

		require.NoError(t, err)
		assert.Equal(t, subscription2, result)
	})

	t.Run("Should return error if not found", func(t *testing.T) {
		subscription1 := fixTestSubscription("test-sub-1", "default", "owner-1")

		serviceFactory, kubeClient, err := NewFakeGenericServiceFactory(v1alpha1.AddToScheme, subscription1)
		require.NoError(t, err)

		service := New(serviceFactory, kubeClient)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.DeleteEventSubscription(context.Background(), "custom", "test-sub-1")

		require.Error(t, err)
		assert.Empty(t, result)
	})
}

func fixTestSpecInput(ownerName string) gqlschema.EventSubscriptionSpecInput {
	params := gqlschema.EventSubscriptionSpecInput{
		Filters: []*gqlschema.FiltersInput{
			{
				ApplicationName: "app",
				Version:         "v1",
				EventName:       "test-passed",
			},
		},
		OwnerRef: &v1.OwnerReference{
			APIVersion: "serverless.kyma-project.io/v1alpha",
			Kind:       "Function",
			Name:       ownerName,
		},
	}
	return params
}

func fixTestSubscription(name, namespace, ownerName string) *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: "serverless.kyma-project.io/v1alpha",
					Kind:       "Function",
					Name:       ownerName,
				},
			},
		},
	}
}

// we can't use the fakeResource version, as it accepts only runtime.Object array - and Subscription isn't a runtime.Object
// moreover we need to mock k8s client with beb secret
func NewFakeGenericServiceFactory(addToScheme func(*runtime.Scheme) error, objects ...*v1alpha1.Subscription) (*resource.GenericServiceFactory, kubernetes.Interface, error) {
	secret := &coreV1.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "kyma-installer",
			Labels: map[string]string{
				"component": "eventing",
			},
		},
		Data: map[string][]byte{
			"authentication.bebNamespace": []byte("test-secret-data"),
		},
	}

	scheme := runtime.NewScheme()
	err := addToScheme(scheme)
	if err != nil {
		return nil, nil, err
	}

	result := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		result[i], err = resource.ToUnstructured(obj)
		if err != nil {
			return nil, nil, err
		}
	}

	client := dynamicFake.NewSimpleDynamicClient(scheme, result...)
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(client, time.Second)
	kubeClient := fake.NewSimpleClientset(secret)
	return resource.NewGenericServiceFactory(client, informerFactory), kubeClient, nil
}
