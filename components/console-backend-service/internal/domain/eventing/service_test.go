package eventing

import (
	"context"
	"testing"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
)

func TestTriggersService_List(t *testing.T) {
	t.Run("Should filter by namespace", func(t *testing.T) {
		defaultSubscriber := duckv1.Destination{
			Ref: nil,
			URI: nil,
		}
		const namespace = "default"

		trigger1 := createMockTrigger("trigger 1", namespace, defaultSubscriber)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", defaultSubscriber)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.TriggersQuery(context.Background(), namespace, nil)

		require.NoError(t, err)

		assert.Equal(t, 1, len(result))
		assert.Equal(t, trigger1.Namespace, result[0].Namespace)
	})

	t.Run("Should filter by subscriber ref", func(t *testing.T) {
		subscriber1 := duckv1.Destination{
			Ref: &duckv1.KReference{
				Kind:       "er",
				Namespace:  "s",
				Name:       "space",
				APIVersion: "ing",
			},
			URI: nil,
		}
		subscriber2 := duckv1.Destination{
			Ref: &duckv1.KReference{
				Kind:       "other",
				Namespace:  "another",
				Name:       "space",
				APIVersion: "ing",
			},
			URI: nil,
		}

		trigger1 := createMockTrigger("trigger 1", "default", subscriber1)
		trigger2 := createMockTrigger("trigger 2", "default", subscriber2)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.TriggersQuery(context.Background(), "default", &subscriber1)

		require.NoError(t, err)

		assert.Equal(t, 1, len(result))
		assert.Equal(t, trigger1.Spec.Subscriber.Ref, result[0].Spec.Subscriber.Ref)
	})

	t.Run("Should filter by subscriber uri", func(t *testing.T) {
		subscriber1 := duckv1.Destination{
			Ref: nil,
			URI: &apis.URL{
				Scheme: "wss",
				Host:   "test",
			},
		}
		subscriber2 := duckv1.Destination{
			Ref: nil,
			URI: &apis.URL{
				Scheme: "wss",
				Host:   "other",
			},
		}

		trigger1 := createMockTrigger("trigger 1", "default", subscriber1)
		trigger2 := createMockTrigger("trigger 2", "default", subscriber2)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.TriggersQuery(context.Background(), "default", &subscriber1)

		require.NoError(t, err)

		assert.Equal(t, 1, len(result))
		assert.Equal(t, trigger1.Spec.Subscriber.Ref, result[0].Spec.Subscriber.Ref)
	})
}

func TestTriggersService_Create(t *testing.T) {
	const namespace = "default"
	t.Run("Should create", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		var name = "trigger"
		trigger, err := service.CreateTrigger(context.Background(), namespace, gqlschema.TriggerCreateInput{
			Name:             &name,
			Broker:           "default",
			FilterAttributes: nil,
			Subscriber: &duckv1.Destination{
				Ref: nil,
				URI: nil,
			},
		}, nil)

		require.NoError(t, err)
		assert.Equal(t, trigger.Name, name)
		assert.Equal(t, trigger.Namespace, namespace)
	})

	t.Run("Should throw if trigger already exists", func(t *testing.T) {
		existingTrigger := createMockTrigger("trigger 1", namespace, duckv1.Destination{
			Ref: nil,
			URI: nil,
		})
		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, existingTrigger)
		require.NoError(t, err)
		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		var name = "trigger 1"
		_, err = service.CreateTrigger(context.Background(), namespace, gqlschema.TriggerCreateInput{
			Name:             &name,
			Broker:           "default",
			FilterAttributes: nil,
			Subscriber: &duckv1.Destination{
				Ref: nil,
				URI: nil,
			},
		}, nil)

		require.Error(t, err)
	})

	t.Run("Should create many", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		input1 := createMockTriggerInput("trigger 1")
		input2 := createMockTriggerInput("trigger 2")
		inputs := []*gqlschema.TriggerCreateInput{input1, input2}

		triggers, err := service.CreateTriggers(context.Background(), "default", inputs, nil)

		require.NoError(t, err)
		require.Equal(t, 2, len(triggers))
	})
}

func TestTriggersService_Delete(t *testing.T) {
	defaultSubscriber := duckv1.Destination{
		Ref: nil,
		URI: nil,
	}
	const namespace = "default"

	t.Run("Should delete", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, defaultSubscriber)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", defaultSubscriber)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		deletedTrigger, err := service.DeleteTrigger(context.Background(), namespace, trigger1.Name)

		require.NoError(t, err)
		assert.Equal(t, trigger1.Name, deletedTrigger.Name)
	})

	t.Run("Should throw if trigger does not exist", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, defaultSubscriber)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", defaultSubscriber)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		_, err = service.DeleteTrigger(context.Background(), namespace, trigger2.Name)

		require.Error(t, err)
	})

	t.Run("Should delete many", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, defaultSubscriber)
		trigger2 := createMockTrigger("trigger 2", namespace, defaultSubscriber)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		names := []string{trigger1.Name, trigger2.Name}
		deletedTriggers, err := service.DeleteTriggers(context.Background(), namespace, names)

		require.NoError(t, err)
		assert.Equal(t, 2, len(deletedTriggers))
	})

	t.Run("Should throw if some triggers are not found", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, defaultSubscriber)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", defaultSubscriber)

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		names := []string{trigger1.Name, trigger2.Name}
		_, err = service.DeleteTriggers(context.Background(), namespace, names)

		require.Error(t, err)
	})
}

func createMockTrigger(name, namespace string, subscriber duckv1.Destination) *v1alpha1.Trigger {
	return &v1alpha1.Trigger{
		TypeMeta: v1.TypeMeta{
			APIVersion: "eventing.knative.dev/v1alpha1",
			Kind:       "Trigger",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.TriggerSpec{
			Broker:     "default",
			Filter:     nil,
			Subscriber: subscriber,
		},
		Status: v1alpha1.TriggerStatus{},
	}
}

func createMockTriggerInput(name string) *gqlschema.TriggerCreateInput {
	return &gqlschema.TriggerCreateInput{
		Name:             &name,
		Broker:           "default",
		FilterAttributes: nil,
		Subscriber: &duckv1.Destination{
			Ref: nil,
			URI: nil,
		},
	}
}
