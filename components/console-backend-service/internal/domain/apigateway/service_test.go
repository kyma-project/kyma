package apigateway

import (
	"testing"
	"time"

	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/listener"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

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

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

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

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

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

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

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

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, apiRule1)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		result, err := service.Find(name, namespace)

		require.NoError(t, err)
		assert.Equal(t, apiRule1, result)
	})

	t.Run("Should return nil if not found", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

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
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		result, err := service.Create(newRule)

		require.NoError(t, err)
		assert.Equal(t, newRule, result)
	})

	t.Run("Should throw an error if APIRule already exists", func(t *testing.T) {
		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, existingApiRule)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

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

		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, existingApiRule)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		result, err := service.Update(newRule)

		require.NoError(t, err)
		newRule := fixTestApiRule(name1, namespace, "new-hostname", serviceName, servicePort1, gateway1)
		assert.Equal(t, *newRule.Spec.Service.Host, *result.Spec.Service.Host)
	})

	t.Run("Should throw an error if APIRule doesn't exists", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		_, err = service.Update(newRule)

		require.Error(t, err)
	})

}

func TestApiRuleService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)
		service.Subscribe(apiRuleListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		apiRuleListener1 := listener.NewApiRule(nil, nil, nil, nil)
		apiRuleListener2 := listener.NewApiRule(nil, nil, nil, nil)

		service.Subscribe(apiRuleListener1)
		service.Subscribe(apiRuleListener2)
	})

	t.Run("Nil", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		service.Subscribe(nil)
	})
}

func TestApiRuleService_Unubscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)

		service.Unsubscribe(apiRuleListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		apiRuleListener := listener.NewApiRule(nil, nil, nil, nil)
		service.Subscribe(apiRuleListener)
		service.Subscribe(apiRuleListener)

		service.Unsubscribe(apiRuleListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		apiRuleListener1 := listener.NewApiRule(nil, nil, nil, nil)
		apiRuleListener2 := listener.NewApiRule(nil, nil, nil, nil)

		service.Subscribe(apiRuleListener1)
		service.Subscribe(apiRuleListener2)

		service.Unsubscribe(apiRuleListener1)
	})

	t.Run("Nil", func(t *testing.T) {
		serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		service := NewService(serviceFactory)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

		service.Unsubscribe(nil)
	})
}
