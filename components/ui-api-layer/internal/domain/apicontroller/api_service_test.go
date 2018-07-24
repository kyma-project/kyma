package apicontroller

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/informers/externalversions"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiService_List(t *testing.T) {
	t.Run("Should filter by environment", func(t *testing.T) {
		api1 := fixAPIWith("test-1", "test-1", "", "")
		api2 := fixAPIWith("test-1", "test-2", "", "")
		api3 := fixAPIWith("test-2", "test-1", "", "")

		informer := fixAPIInformer(api1, api2, api3)
		service := newApiService(informer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List("test-1", nil, nil)

		require.NoError(t, err)
		assert.Equal(t, []*v1alpha2.Api{
			api1, api3,
		}, result)
	})

	t.Run("Should filter by environment and hostname", func(t *testing.T) {
		hostname := "abc"

		api1 := fixAPIWith("test-1", "test-1", hostname, "")
		api2 := fixAPIWith("test-1", "test-2", hostname, "")
		api3 := fixAPIWith("test-2", "test-1", "cba", "")

		informer := fixAPIInformer(api1, api2, api3)
		service := newApiService(informer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List("test-1", nil, &hostname)

		require.NoError(t, err)
		assert.Equal(t, []*v1alpha2.Api{
			api1,
		}, result)
	})

	t.Run("Should filter by environment and serviceName", func(t *testing.T) {
		serviceName := "abc"

		api1 := fixAPIWith("test-2", "test-1", "", serviceName)
		api2 := fixAPIWith("test-3", "test-2", "", serviceName)
		api3 := fixAPIWith("test-4", "test-1", "", "cba")

		informer := fixAPIInformer(api1, api2, api3)
		service := newApiService(informer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List("test-1", &serviceName, nil)

		require.NoError(t, err)
		assert.Equal(t, []*v1alpha2.Api{
			api1,
		}, result)
	})

	t.Run("Should filter by environment serviceName and hostname", func(t *testing.T) {
		serviceName := "abc"
		hostname := "cba"

		api1 := fixAPIWith("test-4", "test-1", hostname, serviceName)
		api2 := fixAPIWith("test-5", "test-2", hostname, serviceName)
		api3 := fixAPIWith("test-6", "test-1", hostname, "cba")

		informer := fixAPIInformer(api1, api2, api3)
		service := newApiService(informer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List("test-1", &serviceName, nil)

		require.NoError(t, err)
		assert.Equal(t, []*v1alpha2.Api{
			api1,
		}, result)
	})
}

func fixAPIWith(name, environment, hostname, serviceName string) *v1alpha2.Api {
	return &v1alpha2.Api{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
		Spec: v1alpha2.ApiSpec{
			Hostname: hostname,
			Service: v1alpha2.Service{
				Name: serviceName,
			},
		},
	}
}

func fixAPIInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 10)

	return informerFactory.Gateway().V1alpha2().Apis().Informer()
}
