package eventing

import (
	"context"
	"fmt"
	"testing"

	"knative.dev/pkg/apis"

	"github.com/stretchr/testify/assert"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"

	"github.com/stretchr/testify/require"
)

func TestTriggersService_List(t *testing.T) {
	t.Run("Should filter by namespace and name", func(t *testing.T) {
		const namespace = "default"
		const serviceName = "serviceName"

		trigger1 := createMockTrigger("trigger 1", namespace, serviceName)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", "other-service")

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, trigger1, trigger2)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.TriggersQuery(context.Background(), namespace, serviceName)

		require.NoError(t, err)

		assert.Equal(t, 1, len(result))
		assert.Equal(t, trigger1.Namespace, result[0].Namespace)
	})

	t.Run("Should filter triggers with either Ref and URI", func(t *testing.T) {
		const namespace = "default"
		const serviceName = "serviceName"

		refTrigger := createMockTriggerWithDestination("refTrigger", namespace, duckv1.Destination{
			Ref: &duckv1.KReference{
				Kind:       "Service",
				Namespace:  namespace,
				Name:       serviceName,
				APIVersion: "",
			},
			URI: nil,
		})

		urlString := fmt.Sprintf("https://%s.%s.svc.cluster.local:80", serviceName, namespace)
		url, err := apis.ParseURL(urlString)

		require.NoError(t, err)
		uriTrigger := createMockTriggerWithDestination("uriTrigger", namespace, duckv1.Destination{
			Ref: nil,
			URI: url,
		})

		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, refTrigger, uriTrigger)
		require.NoError(t, err)

		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		result, err := service.TriggersQuery(context.Background(), namespace, serviceName)

		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestTriggersService_Create(t *testing.T) {
	const namespace = "default"
	const serviceName = "service-name"

	t.Run("Should create", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		var name = "trigger"
		trigger, err := service.CreateTrigger(context.Background(), namespace, createMockTriggerInput(name, namespace, serviceName), nil)

		require.NoError(t, err)
		assert.Equal(t, trigger.Name, name)
		assert.Equal(t, trigger.Namespace, namespace)
	})

	t.Run("Should return error if trigger already exists", func(t *testing.T) {
		const name = "trigger 1"
		existingTrigger := createMockTrigger(name, namespace, serviceName)
		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, existingTrigger)
		require.NoError(t, err)
		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		_, err = service.CreateTrigger(context.Background(), namespace, createMockTriggerInput(name, namespace, serviceName), nil)

		require.Error(t, err)
	})

	t.Run("Should create many", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := New(serviceFactory)
		err = service.Enable()
		require.NoError(t, err)

		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

		input1 := createMockTriggerInput("trigger 1", namespace, serviceName)
		input2 := createMockTriggerInput("trigger 2", namespace, serviceName)
		inputs := []*gqlschema.TriggerCreateInput{&input1, &input2}

		triggers, err := service.CreateTriggers(context.Background(), "default", inputs, nil)

		require.NoError(t, err)
		require.Equal(t, 2, len(triggers))
	})
}

func TestTriggersService_Delete(t *testing.T) {
	const serviceName = "service-name"
	const namespace = "default"

	t.Run("Should delete", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, serviceName)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", serviceName)

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

	t.Run("Should return error if trigger does not exist", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, serviceName)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", serviceName)

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
		trigger1 := createMockTrigger("trigger 1", namespace, serviceName)
		trigger2 := createMockTrigger("trigger 2", namespace, serviceName)

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

	t.Run("Should return error if some triggers are not found", func(t *testing.T) {
		trigger1 := createMockTrigger("trigger 1", namespace, serviceName)
		trigger2 := createMockTrigger("trigger 2", "other-namespace", serviceName)

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

func createMockTrigger(name, namespace, serviceName string) *v1alpha1.Trigger {
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
			Broker: "default",
			Filter: nil,
			Subscriber: duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:       "Service",
					Namespace:  namespace,
					Name:       serviceName,
					APIVersion: "",
				},
				URI: nil,
			},
		},
		Status: v1alpha1.TriggerStatus{},
	}
}

func createMockTriggerWithDestination(name, namespace string, subscriber duckv1.Destination) *v1alpha1.Trigger {
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

func createMockTriggerInput(name, namespace, serviceName string) gqlschema.TriggerCreateInput {
	return gqlschema.TriggerCreateInput{
		Name:             &name,
		Broker:           "default",
		FilterAttributes: nil,
		Subscriber: &gqlschema.SubscriberInput{
			Ref: &duckv1.KReference{
				Kind:       "Service",
				Namespace:  namespace,
				Name:       serviceName,
				APIVersion: "",
			},
			Port: nil,
			Path: nil,
		},
	}
}
