package apigateway

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/listener"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
)

func TestApiRuleService_List(t *testing.T) {

	name1 := "test-apiRule1"
	namespace := "test-namespace"
	hostname := "test-hostname1"
	serviceName := "test-service-name1"
	servicePort1 := uint32(8080)
	gateway1 := "gateway1"

	name2 := "test-apiRule2"
	servicePort2 := uint32(8080)
	gateway2 := "gateway2"

	name3 := "test-apiRule3"
	servicePort3 := uint32(8080)
	gateway3 := "gateway3"

	t.Run("Should filter by namespace", func(t *testing.T) {
		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)
		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2)
		apiRule3 := fixTestApiRule(name3, namespace, hostname, serviceName, servicePort3, gateway3)

		dynamicClient, err := newDynamicClient(apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, nil, nil)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha1.APIRule{
			apiRule1, apiRule3,
		}, result)
	})

	t.Run("Should filter by namespace and hostname", func(t *testing.T) {
		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)
		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2)
		apiRule3 := fixTestApiRule(name3, namespace, "different-hostname", serviceName, servicePort3, gateway3)

		dynamicClient, err := newDynamicClient(apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, nil, &hostname)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha1.APIRule{
			apiRule1,
		}, result)
	})

	t.Run("Should filter by namespace and serviceName", func(t *testing.T) {
		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)
		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2)
		apiRule3 := fixTestApiRule(name3, namespace, hostname, "different-service-name", servicePort3, gateway3)

		dynamicClient, err := newDynamicClient(apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, &serviceName, nil)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha1.APIRule{
			apiRule1,
		}, result)
	})

	t.Run("Should filter by namespace serviceName and hostname", func(t *testing.T) {
		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)
		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2)
		apiRule3 := fixTestApiRule(name3, namespace, hostname, "different-service-name", servicePort3, gateway3)

		dynamicClient, err := newDynamicClient(apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.List(namespace, &serviceName, nil)

		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1alpha1.APIRule{
			apiRule1,
		}, result)
	})
}

func TestApiService_Find(t *testing.T) {
	name1 := "test-apiRule1"
	namespace := "test-namespace"
	hostname := "test-hostname1"
	serviceName := "test-service-name1"
	servicePort1 := uint32(8080)
	gateway1 := "gateway1"

	t.Run("Should find an APIRule", func(t *testing.T) {
		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)

		dynamicClient, err := newDynamicClient(apiRule1)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Find(name, namespace)

		require.NoError(t, err)
		assert.Equal(t, apiRule1, result)
	})

	t.Run("Should return nil if not found", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Find(name, namespace)

		require.NoError(t, err)
		var empty *v1alpha1.APIRule
		assert.Equal(t, empty, result)
	})
}

func TestApiService_Create(t *testing.T) {
	name1 := "test-apiRule1"
	namespace := "test-namespace"
	hostname := "test-hostname1"
	serviceName := "test-service-name1"
	servicePort1 := uint32(8080)
	gateway1 := "gateway1"

	newRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)

	t.Run("Should create an APIRule", func(t *testing.T) {
		dynamicClient, err := newDynamicClient()
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Create(newRule)

		require.NoError(t, err)
		assert.Equal(t, newRule, result)
	})

	t.Run("Should throw an error if APIRule already exists", func(t *testing.T) {
		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)
		dynamicClient, err := newDynamicClient(existingApiRule)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		_, err = service.Create(newRule)

		require.Error(t, err)
	})

}

func TestApiRuleService_Update(t *testing.T) {
	name1 := "test-apiRule1"
	namespace := "test-namespace"
	hostname := "test-hostname1"
	serviceName := "test-service-name1"
	servicePort1 := uint32(8080)
	gateway1 := "gateway1"

	newRule := fixTestApiRule(name1, namespace, "new-hostname", serviceName, servicePort1, gateway1)

	t.Run("Should update an APIRule", func(t *testing.T) {
		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)
		dynamicClient, err := newDynamicClient(existingApiRule)
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := service.Update(newRule)

		require.NoError(t, err)
		newRule := fixTestApiRule(name1, namespace, "new-hostname", serviceName, servicePort1, gateway1)
		assert.Equal(t, *newRule.Spec.Service.Host, *result.Spec.Service.Host)
	})

	t.Run("Should throw an error if APIRule doesn't exists", func(t *testing.T) {
		dynamicClient, err := newDynamicClient()
		require.NoError(t, err)
		informer := createApiRuleFakeInformer(dynamicClient)
		client := createApiRuleDynamicClient(dynamicClient)
		service := newApiRuleService(informer, client)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		_, err = service.Update(newRule)

		require.Error(t, err)
	})

}

func TestApiRuleService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)
		service.Subscribe(apiRuleListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiRuleListener1 := listener.NewApiRule(nil, nil, nil, nil)
		apiRuleListener2 := listener.NewApiRule(nil, nil, nil, nil)

		service.Subscribe(apiRuleListener1)
		service.Subscribe(apiRuleListener2)
	})

	t.Run("Nil", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		service.Subscribe(nil)
	})
}

func TestApiRuleService_Unubscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)

		service.Unsubscribe(apiRuleListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)
		service.Subscribe(apiRuleListener)

		service.Unsubscribe(apiRuleListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		apiRuleListener1 := listener.NewApiRule(nil, nil, nil, nil)
		apiRuleListener2 := listener.NewApiRule(nil, nil, nil, nil)

		service.Subscribe(apiRuleListener1)
		service.Subscribe(apiRuleListener2)

		service.Unsubscribe(apiRuleListener1)
	})

	t.Run("Nil", func(t *testing.T) {
		informer := fixAPIRuleInformer(t)
		service := newApiRuleService(informer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		service.Unsubscribe(nil)
	})
}

func fixAPIRuleInformer(t *testing.T, objects ...runtime.Object) cache.SharedIndexInformer {
	dynamicClient, err := newDynamicClient(objects...)
	require.NoError(t, err)
	return createApiRuleFakeInformer(dynamicClient)
}

func createApiRuleDynamicClient(dynamicClient dynamic.Interface) dynamic.NamespaceableResourceInterface {
	return dynamicClient.Resource(schema.GroupVersionResource{
		Version:  v1alpha1.GroupVersion.Version,
		Group:    v1alpha1.GroupVersion.Group,
		Resource: "apirules",
	})
}

func createApiRuleFakeInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamic, 10)
	return informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.GroupVersion.Version,
		Group:    v1alpha1.GroupVersion.Group,
		Resource: "apirules",
	}).Informer()
}

func newDynamicClient(objects ...runtime.Object) (*dynamicFake.FakeDynamicClient, error) {
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
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
