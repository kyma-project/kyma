package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestServiceService_Find(t *testing.T) {
	namespace := "namespace"
	instanceName := "testExample"

	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		service := fixService(instanceName, namespace, nil)
		serviceInformer, client := fixServiceInformer(service)

		svc := k8s.NewServiceService(serviceInformer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(service, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()

		svc := k8s.NewServiceService(serviceInformer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(instance)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		expectedService := fixService(instanceName, namespace, nil)
		returnedService := fixServiceWithoutTypeMeta(instanceName, namespace, nil)
		serviceInformer, client := fixServiceInformer(returnedService)

		svc := k8s.NewServiceService(serviceInformer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(expectedService, instance)
	})
}

func TestServiceService_List(t *testing.T) {
	namespace := "namespace"

	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		service1 := fixService("service1", namespace, nil)
		service2 := fixService("service2", namespace, nil)
		service3 := fixService("service3", "differentNamespace", nil)

		serviceInformer, client := fixServiceInformer(service1, service2, service3)
		svc := k8s.NewServiceService(serviceInformer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		services, err := svc.List(namespace, nil, pager.PagingParams{})
		require.NoError(t, err)
		assert.ElementsMatch([]*v1.Service{
			service2, service1,
		}, services)
	})

	t.Run("NotFound", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()

		svc := k8s.NewServiceService(serviceInformer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		services, err := svc.List("notExistingNamespace", nil, pager.PagingParams{})
		require.NoError(t, err)
		assert.Empty(services)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		returnedService1 := fixServiceWithoutTypeMeta("service1", namespace, nil)
		returnedService2 := fixServiceWithoutTypeMeta("service2", namespace, nil)
		returnedService3 := fixServiceWithoutTypeMeta("service3", "differentNamespace", nil)
		expectedService1 := fixService("service1", namespace, nil)
		expectedService2 := fixService("service2", namespace, nil)

		serviceInformer, client := fixServiceInformer(returnedService1, returnedService2, returnedService3)

		svc := k8s.NewServiceService(serviceInformer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		services, err := svc.List(namespace, nil, pager.PagingParams{})
		require.NoError(t, err)
		assert.EqualValues([]*v1.Service{
			expectedService1, expectedService2,
		}, services)
	})
}

func TestServiceService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)
		serviceListener := listener.NewService(nil, nil, nil)
		svc.Subscribe(serviceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)
		serviceListener := listener.NewService(nil, nil, nil)

		svc.Subscribe(serviceListener)
		svc.Subscribe(serviceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)
		serviceListenerA := listener.NewService(nil, nil, nil)
		serviceListenerB := listener.NewService(nil, nil, nil)

		svc.Subscribe(serviceListenerA)
		svc.Subscribe(serviceListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)

		svc.Subscribe(nil)
	})
}

func TestServiceService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)
		serviceListener := listener.NewService(nil, nil, nil)
		svc.Subscribe(serviceListener)

		svc.Unsubscribe(serviceListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)
		serviceListener := listener.NewService(nil, nil, nil)
		svc.Subscribe(serviceListener)
		svc.Subscribe(serviceListener)

		svc.Unsubscribe(serviceListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)
		serviceListenerA := listener.NewService(nil, nil, nil)
		serviceListenerB := listener.NewService(nil, nil, nil)
		svc.Subscribe(serviceListenerA)
		svc.Subscribe(serviceListenerB)

		svc.Unsubscribe(serviceListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)

		svc.Unsubscribe(nil)
	})
}

func TestServiceService_Update(t *testing.T) {
	assert := assert.New(t)
	t.Run("Success", func(t *testing.T) {
		exampleName := "exampleService"
		exampleNamespace := "exampleNamespace"
		exampleService := fixService(exampleName, exampleNamespace, nil)
		serviceInformer, client := fixServiceInformer(exampleService)
		svc := k8s.NewServiceService(serviceInformer, client)

		update := exampleService.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		service, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(update, service)

		service, err = client.Services(exampleNamespace).Get(exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(update, service)
	})

	t.Run("NotFound", func(t *testing.T) {
		exampleName := "exampleService"
		exampleNamespace := "exampleNamespace"
		exampleService := fixService(exampleName, exampleNamespace, nil)
		serviceInformer, client := fixServiceInformer()
		svc := k8s.NewServiceService(serviceInformer, client)

		update := exampleService.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		service, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(service)

		service, err = client.Services(exampleNamespace).Get(exampleName, metav1.GetOptions{})
		require.Error(t, err)
		assert.Nil(service)
	})

	t.Run("NameMismatch", func(t *testing.T) {
		exampleName := "exampleService"
		exampleNamespace := "exampleNamespace"
		exampleService := fixService(exampleName, exampleNamespace, nil)
		serviceInformer, client := fixServiceInformer(exampleService)
		svc := k8s.NewServiceService(serviceInformer, client)

		update := exampleService.DeepCopy()
		update.Name = "NameMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		service, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(apierror.IsInvalid(err))
		assert.Nil(service)

		service, err = client.Services(exampleNamespace).Get(exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(exampleService, service)
	})
}

func TestServiceService_Delete(t *testing.T) {
	assert := assert.New(t)
	t.Run("Success", func(t *testing.T) {
		exampleName := "exampleService"
		exampleNamespace := "exampleNamespace"
		exampleService := fixService(exampleName, exampleNamespace, nil)
		serviceInformer, client := fixServiceInformer(exampleService)
		svc := k8s.NewServiceService(serviceInformer, client)

		err := svc.Delete(exampleName, exampleNamespace)

		require.NoError(t, err)
		_, err = client.Services(exampleNamespace).Get(exampleName, metav1.GetOptions{})
		assert.True(errors.IsNotFound(err))
	})
}

func fixServiceInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Services().Informer()

	return informer, client.CoreV1()
}

func fixService(name, namespace string, labels map[string]string) *v1.Service {
	service := fixServiceWithoutTypeMeta(name, namespace, labels)
	service.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
	return service
}

func fixServiceWithoutTypeMeta(name, namespace string, labels map[string]string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
