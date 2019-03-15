package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	_assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	instanceName = "testExample"
)

func TestServiceService_Find(t *testing.T) {

	assert := _assert.New(t)

	t.Run("Success", func(t *testing.T) {

		service := fixService(instanceName, namespace, nil)
		serviceInformer, _ := fixServiceInformer(service)

		svc := k8s.NewServiceService(serviceInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(service, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		serviceInformer, _ := fixServiceInformer()

		svc := k8s.NewServiceService(serviceInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(instance)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {

		expectedService := fixService(instanceName, namespace, nil)
		returnedService := fixServiceWithoutTypeMeta(instanceName, namespace, nil)
		serviceInformer, _ := fixServiceInformer(returnedService)

		svc := k8s.NewServiceService(serviceInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(expectedService, instance)
	})
}

func TestServiceService_List(t *testing.T) {

	assert := _assert.New(t)

	t.Run("Success", func(t *testing.T) {

		service1 := fixService("srevice1", namespace, nil)
		service2 := fixService("service2", namespace, nil)
		service3 := fixService("service3", "differentNamespace", nil)

		serviceInformer, _ := fixServiceInformer(service1, service2, service3)
		svc := k8s.NewServiceService(serviceInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		services, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.ElementsMatch([]*v1.Service{
			service2, service1,
		}, services)
	})

	t.Run("NotFound", func(t *testing.T) {
		serviceInformer, _ := fixServiceInformer()

		svc := k8s.NewServiceService(serviceInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		var emptyArray []*v1.Service
		services, err := svc.List("notExistingNamespace", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(emptyArray, services)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {

		returnedService1 := fixServiceWithoutTypeMeta("service1", namespace, nil)
		returnedService2 := fixServiceWithoutTypeMeta("service2", namespace, nil)
		returnedService3 := fixServiceWithoutTypeMeta("service3", "differentNamespace", nil)
		expectedService1 := fixService("service1", namespace, nil)
		expectedService2 := fixService("service2", namespace, nil)

		serviceInformer, _ := fixServiceInformer(returnedService1, returnedService2, returnedService3)

		svc := k8s.NewServiceService(serviceInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, serviceInformer)

		services, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.EqualValues([]*v1.Service{
			expectedService1, expectedService2,
		}, services)
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
