package apigateway

import (
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

// func TestApiRuleService_List(t *testing.T) {
// 	name1 := "test-apiRule1"
// 	namespace := "test-namespace"
// 	hostname := "test-hostname1"
// 	serviceName := "test-service-name1"
// 	servicePort1 := uint32(8080)
// 	gateway1 := "gateway1"

// 	name2 := "test-apiRule2"
// 	servicePort2 := uint32(8080)
// 	gateway2 := "gateway2"

// 	name3 := "test-apiRule3"
// 	servicePort3 := uint32(8080)
// 	gateway3 := "gateway3"

// 	defaultGeneration := int64(0)

// 	t.Run("Should filter by namespace", func(t *testing.T) {
// 		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)
// 		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2, defaultGeneration)
// 		apiRule3 := fixTestApiRule(name3, namespace, hostname, serviceName, servicePort3, gateway3, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.APIRulesQuery(context.Background(), namespace, nil, nil)

// 		require.NoError(t, err)
// 		assert.ElementsMatch(t, []*v1alpha1.APIRule{
// 			apiRule1, apiRule3,
// 		}, result)
// 	})

// 	t.Run("Should filter by namespace and hostname", func(t *testing.T) {
// 		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)
// 		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2, defaultGeneration)
// 		apiRule3 := fixTestApiRule(name3, namespace, "different-hostname", serviceName, servicePort3, gateway3, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.APIRulesQuery(context.Background(), namespace, nil, &hostname)

// 		require.NoError(t, err)
// 		assert.ElementsMatch(t, []*v1alpha1.APIRule{
// 			apiRule1,
// 		}, result)
// 	})

// 	t.Run("Should filter by namespace and serviceName", func(t *testing.T) {
// 		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)
// 		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2, defaultGeneration)
// 		apiRule3 := fixTestApiRule(name3, namespace, hostname, "different-service-name", servicePort3, gateway3, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.APIRulesQuery(context.Background(), namespace, &serviceName, nil)

// 		require.NoError(t, err)
// 		assert.ElementsMatch(t, []*v1alpha1.APIRule{
// 			apiRule1,
// 		}, result)
// 	})

// 	t.Run("Should filter by namespace serviceName and hostname", func(t *testing.T) {
// 		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)
// 		apiRule2 := fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2, defaultGeneration)
// 		apiRule3 := fixTestApiRule(name3, namespace, hostname, "different-service-name", servicePort3, gateway3, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, apiRule1, apiRule2, apiRule3)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.APIRulesQuery(context.Background(), namespace, &serviceName, &hostname)

// 		require.NoError(t, err)
// 		assert.ElementsMatch(t, []*v1alpha1.APIRule{
// 			apiRule1,
// 		}, result)
// 	})
// }

// func TestApiService_Find(t *testing.T) {
// 	name1 := "test-apiRule1"
// 	namespace := "test-namespace"
// 	hostname := "test-hostname1"
// 	serviceName := "test-service-name1"
// 	servicePort1 := uint32(8080)
// 	gateway1 := "gateway1"
// 	defaultGeneration := int64(0)

// 	t.Run("Should find an APIRule", func(t *testing.T) {
// 		apiRule1 := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, apiRule1)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.APIRuleQuery(context.Background(), apiRule1.Name, namespace)

// 		require.NoError(t, err)
// 		assert.Equal(t, apiRule1, result)
// 	})

// 	t.Run("Should return error if not found", func(t *testing.T) {
// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		_, err = service.APIRuleQuery(context.Background(), name1, namespace)

// 		require.Error(t, err)
// 	})
// }

// func TestApiService_Create(t *testing.T) {
// 	name1 := "test-apiRule1"
// 	namespace := "test-namespace"
// 	hostname := "test-hostname1"
// 	serviceName := "test-service-name1"
// 	servicePort1 := uint32(8080)
// 	gateway1 := "gateway1"
// 	defaultGeneration := int64(0)

// 	newRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)

// 	t.Run("Should create an APIRule", func(t *testing.T) {
// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
// 		require.NoError(t, err)
// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.CreateAPIRule(context.Background(), newRule.Name, newRule.Namespace, newRule.Spec)

// 		require.NoError(t, err)
// 		assert.Equal(t, newRule, result)
// 	})

// 	t.Run("Should throw an error if APIRule already exists", func(t *testing.T) {
// 		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, existingApiRule)
// 		require.NoError(t, err)
// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		_, err = service.CreateAPIRule(context.Background(), newRule.Name, newRule.Namespace, newRule.Spec)

// 		require.Error(t, err)
// 	})

// }

// func TestApiRuleService_Update(t *testing.T) {
// 	name1 := "test-apiRule1"
// 	namespace := "test-namespace"
// 	hostname := "test-hostname1"
// 	serviceName := "test-service-name1"
// 	servicePort1 := uint32(8080)
// 	gateway1 := "gateway1"
// 	defaultGeneration := int64(0)

// 	newRule := fixTestApiRule(name1, namespace, "new-hostname", serviceName, servicePort1, gateway1, defaultGeneration)

// 	t.Run("Should update an APIRule", func(t *testing.T) {
// 		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, existingApiRule)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		result, err := service.UpdateAPIRule(context.Background(), newRule.Name, newRule.Namespace, newRule.Generation, newRule.Spec)

// 		require.NoError(t, err)
// 		newRule := fixTestApiRule(name1, namespace, "new-hostname", serviceName, servicePort1, gateway1, defaultGeneration)
// 		assert.Equal(t, *newRule.Spec.Service.Host, *result.Spec.Service.Host)
// 	})

// 	t.Run("Should throw an error if APIRule doesn't exists", func(t *testing.T) {
// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		_, err = service.UpdateAPIRule(context.Background(), newRule.Name, newRule.Namespace, newRule.Generation, newRule.Spec)

// 		require.Error(t, err)
// 	})

// 	t.Run("Should throw an error on updating already modified APIRule", func(t *testing.T) {
// 		existingGeneration := int64(2)
// 		updateGeneration := int64(1)

// 		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, existingGeneration)

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, existingApiRule)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		_, err = service.UpdateAPIRule(context.Background(), newRule.Name, newRule.Namespace, updateGeneration, newRule.Spec)

// 		require.Error(t, err)
// 	})

// 	t.Run("Should throw an error on updating readonly APIRule", func(t *testing.T) {
// 		existingApiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1, defaultGeneration)
// 		existingApiRule.OwnerReferences = []v1.OwnerReference{{
// 			Kind: "Subscription",
// 			Name: "sup",
// 		}}

// 		serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, existingApiRule)
// 		require.NoError(t, err)

// 		service := New(serviceFactory)
// 		err = service.Enable()
// 		require.NoError(t, err)

// 		serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

// 		_, err = service.UpdateAPIRule(context.Background(), newRule.Name, newRule.Namespace, defaultGeneration, newRule.Spec)

// 		require.Error(t, err)
// 	})
// }

// Apaprently watch does not work with fake client

//func TestApiRuleService_Subscribe(t *testing.T) {
//	name := "test-apiRule1"
//	namespace := "test-namespace"
//	servicePort1 := uint32(8080)
//
//	newRule := fixTestApiRule(name, namespace, "new-hostname", "service", servicePort1, "gateway", defaultGeneration)
//
//	serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme)
//	require.NoError(t, err)
//
//	service := New(serviceFactory)
//	err = service.Enable()
//	require.NoError(t, err)
//	serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))
//
//	ctx, cancel := context.WithCancel(context.Background())
//	channel, err := service.APIRuleEventSubscription(ctx, namespace, nil)
//	require.NoError(t, err)
//	created, err := service.CreateAPIRule(context.Background(), newRule.Name, newRule.Namespace, newRule.Spec)
//	require.NoError(t, err)
//
//	var event *gqlschema.APIRuleEvent
//	timeout := time.After(1 * time.Second)
//	select {
//	case event = <-channel:
//		break
//	case <-timeout:
//		break
//	}
//	require.NotNil(t, event)
//
//	assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, event.Type)
//	assert.Equal(t, created, event.APIRule)
//
//	cancel()
//	_, opened := <-channel
//	assert.False(t, opened)
//}

func fixTestApiRule(ruleName string, namespace string, hostName string, serviceName string, servicePort uint32, gateway string, generation int64) *v1alpha1.APIRule {
	return &v1alpha1.APIRule{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v1alpha1",
			Kind:       "APIRule",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:       ruleName,
			Namespace:  namespace,
			Generation: generation,
		},
		Spec: v1alpha1.APIRuleSpec{
			Service: &v1alpha1.Service{
				Host: &hostName,
				Name: &serviceName,
				Port: &servicePort,
			},
			Gateway: &gateway,
			Rules: []v1alpha1.Rule{
				{
					Path:    "*",
					Methods: []string{"GET"},
					AccessStrategies: []*rulev1alpha1.Authenticator{
						{
							Handler: &rulev1alpha1.Handler{
								Name: "allow",
								Config: &runtime.RawExtension{
									Raw: []byte("{}"),
								},
							},
						},
					},
				},
			},
		},
	}
}
