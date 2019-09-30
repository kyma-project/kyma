package apicontroller

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/listener"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
)

func TestApiService_List(t *testing.T) {
	namespace := "test-namespace"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"
	t.Run("Should filter by namespace", func(t *testing.T) {
		api1 := fixAPIWith("test-1", namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api2 := fixAPIWith("test-2", "different-namespace", hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api3 := fixAPIWith("test-3", namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)

		dynamicClient, err := newDynamicClient(api1, api2, api3)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, nil, nil)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha2.Api{
			api1, api3,
		}, result)
	})

	t.Run("Should filter by namespace and hostname", func(t *testing.T) {
		api1 := fixAPIWith("test-1", namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api2 := fixAPIWith("test-2", "different-namespace", hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api3 := fixAPIWith("test-3", namespace, "different-hostname", serviceName, jwksUri, issuer, servicePort, nil, nil)

		dynamicClient, err := newDynamicClient(api1, api2, api3)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, nil, &hostname)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha2.Api{
			api1,
		}, result)
	})

	t.Run("Should filter by namespace and serviceName", func(t *testing.T) {
		serviceName := "abc"

		api1 := fixAPIWith("test-1", namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api2 := fixAPIWith("test-2", "different-namespace", hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api3 := fixAPIWith("test-3", namespace, hostname, "different-service-name", jwksUri, issuer, servicePort, nil, nil)

		dynamicClient, err := newDynamicClient(api1, api2, api3)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, &serviceName, nil)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha2.Api{
			api1,
		}, result)
	})

	t.Run("Should filter by namespace serviceName and hostname", func(t *testing.T) {
		serviceName := "abc"
		hostname := "cba"

		api1 := fixAPIWith("test-1", namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api2 := fixAPIWith("test-2", "different-namespace", hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		api3 := fixAPIWith("test-3", namespace, hostname, "different-service-name", jwksUri, issuer, servicePort, nil, nil)

		dynamicClient, err := newDynamicClient(api1, api2, api3)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, &serviceName, nil)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha2.Api{
			api1,
		}, result)
	})
}

func TestApiService_Find(t *testing.T) {
	name := "test-api"
	namespace := "test-namespace"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"
	t.Run("Should find an API", func(t *testing.T) {
		api := fixAPIWith(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)

		dynamicClient, err := newDynamicClient(api)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Find(name, namespace)

		require.NoError(t, err)
		assert.Equal(t, api, result)
	})

	t.Run("Should return nil if not found", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Find(name, namespace)

		require.NoError(t, err)
		var empty *v1alpha2.Api
		assert.Equal(t, empty, result)
	})
}

func TestApiService_Create(t *testing.T) {
	name := "test-api"
	namespace := "test-namespace"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"

	newApi := fixAPIWith(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)

	t.Run("Should create an API", func(t *testing.T) {
		dynamicClient, err := newDynamicClient()
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Create(newApi)

		require.NoError(t, err)
		assert.Equal(t, newApi, result)
	})

	t.Run("Should throw an error if API already exists", func(t *testing.T) {
		existingApi := fixAPIWith(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		dynamicClient, err := newDynamicClient(existingApi)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		_, err = service.Create(newApi)

		require.Error(t, err)
	})

}

func TestApiService_Update(t *testing.T) {
	name := "test-api"
	namespace := "test-namespace"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"

	newApi := fixAPIWith(name, namespace, "new-hostname", serviceName, jwksUri, issuer, servicePort, nil, nil)

	t.Run("Should update an API", func(t *testing.T) {
		existingApi := fixAPIWith(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
		dynamicClient, err := newDynamicClient(existingApi)
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Update(newApi)

		require.NoError(t, err)
		newApi := fixAPIWith(name, namespace, "new-hostname", serviceName, jwksUri, issuer, servicePort, nil, nil)
		assert.Equal(t, newApi, result)
	})

	t.Run("Should throw an error if API doesn't exists", func(t *testing.T) {
		dynamicClient, err := newDynamicClient()
		require.NoError(t, err)
		informer := createApiFakeInformer(dynamicClient)
		client := createApiDynamicClient(dynamicClient)
		service := newApiService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		_, err = service.Update(newApi)

		require.Error(t, err)
	})

}

func TestApiService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiListener := listener.NewApi(nil, nil, nil, nil)
		service.Subscribe(apiListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiListener := listener.NewApi(nil, nil, nil, nil)
		service.Subscribe(apiListener)
		service.Subscribe(apiListener)
	})

	t.Run("Miltiple", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiListener1 := listener.NewApi(nil, nil, nil, nil)
		apiListener2 := listener.NewApi(nil, nil, nil, nil)

		service.Subscribe(apiListener1)
		service.Subscribe(apiListener2)
	})

	t.Run("Nil", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		service.Subscribe(nil)
	})
}

func TestApiService_Unubscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiListener := listener.NewApi(nil, nil, nil, nil)
		service.Subscribe(apiListener)

		service.Unsubscribe(apiListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiListener := listener.NewApi(nil, nil, nil, nil)
		service.Subscribe(apiListener)
		service.Subscribe(apiListener)

		service.Unsubscribe(apiListener)
	})

	t.Run("Miltiple", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiListener1 := listener.NewApi(nil, nil, nil, nil)
		apiListener2 := listener.NewApi(nil, nil, nil, nil)

		service.Subscribe(apiListener1)
		service.Subscribe(apiListener2)

		service.Unsubscribe(apiListener1)
	})

	t.Run("Nil", func(t *testing.T) {
		informer := fixAPIInformer(t)
		service := newApiService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		service.Unsubscribe(nil)
	})
}

func fixAPIWith(name, namespace, hostname, serviceName, jwksUri, issuer string, servicePort int, disableIstioAuthPolicyMTLS, authenticationEnabled *bool) *v1alpha2.Api {

	api := v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v1alpha2",
			Kind:       "Api",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha2.ApiSpec{
			Service: v1alpha2.Service{
				Name: serviceName,
				Port: servicePort,
			},
			Hostname: hostname,
			Authentication: []v1alpha2.AuthenticationRule{
				{
					Jwt: v1alpha2.JwtAuthentication{
						JwksUri: jwksUri,
						Issuer:  issuer,
					},
					Type: v1alpha2.AuthenticationType("JWT"),
				},
			},
			DisableIstioAuthPolicyMTLS: disableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      authenticationEnabled,
		},
	}
	return &api
}

func fixAPIInformer(t *testing.T, objects ...runtime.Object) cache.SharedIndexInformer {
	dynamicClient, err := newDynamicClient(objects...)
	require.NoError(t, err)
	return createApiFakeInformer(dynamicClient)
}

func createApiDynamicClient(dynamicClient dynamic.Interface) dynamic.NamespaceableResourceInterface {
	return dynamicClient.Resource(schema.GroupVersionResource{
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Resource: "apis",
	})
}

func createApiFakeInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	appInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamic, 10)
	return appInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Resource: "apis",
	}).Informer()
}

func newDynamicClient(objects ...runtime.Object) (*dynamicFake.FakeDynamicClient, error) {
	scheme := runtime.NewScheme()
	err := v1alpha2.AddToScheme(scheme)
	if err != nil {
		return &dynamicFake.FakeDynamicClient{}, err
	}
	result := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		converted, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		result[i] = &unstructured.Unstructured{Object: converted}
	}
	return dynamicFake.NewSimpleDynamicClient(scheme, result...), nil
}
