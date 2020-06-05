package apigateway

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespace = "test-1"

	name         = "test-apiRule1"
	hostname1    = "test-hostname1"
	serviceName1 = "test-service-name1"
	servicePort1 = uint32(8080)
	gateway1     = "gateway1"

	hostname2    = "test-hostname2"
	serviceName2 = "test-service-name2"
	servicePort2 = uint32(8080)
	gateway2     = "gateway2"
)

func TestApiRuleResolver_APIRulesQuery(t *testing.T) {

	t.Run("Should return a list of APIRules for namespace", func(t *testing.T) {
		apiRules := []*v1alpha1.APIRule{fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1), fixTestApiRule("test-2", namespace, hostname2, serviceName2, servicePort2, gateway2)}

		expected := []gqlschema.APIRule{testApiRuleToGQL(name, hostname1, serviceName1, servicePort1, gateway1), testApiRuleToGQL("test-2", hostname2, serviceName2, servicePort2, gateway2)}

		var empty *string = nil

		service := automock.NewApiRuleSvc()
		service.On("List", namespace, empty, empty).Return(apiRules, nil).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		result, err := resolver.APIRulesQuery(nil, namespace, nil, nil)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected[0].Name, result[0].Name)
	})

	t.Run("Should return an error", func(t *testing.T) {
		var empty *string = nil

		service := automock.NewApiRuleSvc()
		service.On("List", namespace, empty, empty).Return(nil, errors.New("test")).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		_, err = resolver.APIRulesQuery(nil, namespace, nil, nil)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestApiRuleResolver_APIRuleQuery(t *testing.T) {
	t.Run("Should return a an API in namespace", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)

		expected := testApiRuleToGQL(name, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Find", name, namespace).Return(apiRule, nil).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		result, err := resolver.APIRuleQuery(nil, name, namespace)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected.Name, result.Name)
	})

	t.Run("Should return an empty object", func(t *testing.T) {
		service := automock.NewApiRuleSvc()
		service.On("Find", name, namespace).Return(nil, nil).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		result, err := resolver.APIRuleQuery(nil, name, namespace)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, (*gqlschema.APIRule)(nil), result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		service := automock.NewApiRuleSvc()
		service.On("Find", name, namespace).Return(nil, errors.New("test")).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		_, err = resolver.APIRuleQuery(nil, name, namespace)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestApiRuleResolver_CreateAPIRule(t *testing.T) {
	params := apiRuleInputParams(hostname1, serviceName1, servicePort1, gateway1)
	t.Run("Should create an APIRule", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)
		expected := testApiRuleToGQL(name, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Create", apiRule).Return(apiRule, nil).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		result, err := resolver.CreateAPIRule(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected, *result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Create", apiRule).Return(nil, errors.New("test")).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		_, err = resolver.CreateAPIRule(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestApiRuleResolver_UpdateAPIRule(t *testing.T) {
	params := apiRuleInputParams(hostname1, serviceName1, servicePort1, gateway1)
	t.Run("Should update an API", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)
		expected := testApiRuleToGQL(name, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Update", apiRule).Return(apiRule, nil).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		result, err := resolver.UpdateAPIRule(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected, *result)
	})
	t.Run("Should return an error", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Update", apiRule).Return(nil, errors.New("test")).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		_, err = resolver.UpdateAPIRule(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestApiRuleResolver_DeleteAPIRule(t *testing.T) {
	t.Run("Should delete an APIRule", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)

		expected := testApiRuleToGQL(name, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Find", name, namespace).Return(apiRule, nil).Once()
		service.On("Delete", name, namespace).Return(nil).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		result, err := resolver.DeleteAPIRule(nil, name, namespace)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected.Name, result.Name)
	})

	t.Run("Should return an error if APIRule has not been found", func(t *testing.T) {
		service := automock.NewApiRuleSvc()
		service.On("Find", name, namespace).Return(nil, errors.New("test")).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		_, err = resolver.DeleteAPIRule(nil, name, namespace)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Should return an error if APIRule couldn't be removed", func(t *testing.T) {
		apiRule := fixTestApiRule(name, namespace, hostname1, serviceName1, servicePort1, gateway1)

		service := automock.NewApiRuleSvc()
		service.On("Find", name, namespace).Return(apiRule, nil).Once()
		service.On("Delete", name, namespace).Return(errors.New("test")).Once()

		resolver, err := newApiRuleResolver(service)
		require.NoError(t, err)

		_, err = resolver.DeleteAPIRule(nil, name, namespace)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func fixTestApiRule(ruleName string, namespace string, hostName string, serviceName string, servicePort uint32, gateway string) *v1alpha1.APIRule {
	return &v1alpha1.APIRule{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v1alpha1",
			Kind:       "APIRule",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      ruleName,
			Namespace: namespace,
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

func testApiRuleToGQL(ruleName string, hostName string, serviceName string, servicePort uint32, gateway string) gqlschema.APIRule {
	return gqlschema.APIRule{
		Name: ruleName,
		Service: gqlschema.APIRuleService{
			Host: hostName,
			Port: int(servicePort),
			Name: serviceName,
		},
		Gateway: gateway,
		Rules: []gqlschema.Rule{
			{
				Path:    "*",
				Methods: []string{"GET"},
				AccessStrategies: []gqlschema.APIRuleConfig{
					{
						Name:   "allow",
						Config: gqlschema.JSON{},
					},
				},
			},
		},
	}
}

func apiRuleInputParams(hostName string, serviceName string, servicePort uint32, gateway string) gqlschema.APIRuleInput {
	return gqlschema.APIRuleInput{
		Host:        hostName,
		ServiceName: serviceName,
		ServicePort: int(servicePort),
		Gateway:     gateway,
		Rules: []gqlschema.RuleInput{
			{
				Path:    "*",
				Methods: []string{"GET"},
				AccessStrategies: []gqlschema.APIRuleConfigInput{
					{
						Name:   "allow",
						Config: gqlschema.JSON{},
					},
				},
			},
		},
	}
}
