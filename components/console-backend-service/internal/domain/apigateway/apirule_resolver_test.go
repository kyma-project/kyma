package apigateway

import (
	"errors"
	"testing"

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
		apiRules := []*v1alpha1.APIRule{fixTestApiRule(name, hostname1, serviceName1, servicePort1, gateway1), fixTestApiRule("test-2", hostname2, serviceName2, servicePort2, gateway2)}

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
		apiRule := fixTestApiRule(name, hostname1, serviceName1, servicePort1, gateway1)

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

//func TestApiResolver_CreateAPI(t *testing.T) {
//	namespace := "test-1"
//	name := "test-api"
//	hostname := "test-hostname"
//	serviceName := "test-service-name"
//	servicePort := 8080
//	jwksUri := "http://test-jwks-uri"
//	issuer := "test-issuer"
//
//	params := paramsToAPICreationInput(hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
//
//	t.Run("Should create an API", func(t *testing.T) {
//		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
//		expected := testApiToGQL(name, hostname, serviceName, jwksUri, issuer, servicePort)
//
//		converter := automock.NewApiConv()
//		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()
//
//		service := automock.NewApiSvc()
//		service.On("Create", api).Return(api, nil).Once()
//
//		resolver, err := newApiResolver(service)
//		require.NoError(t, err)
//
//		result, err := resolver.CreateAPI(nil, name, namespace, params)
//
//		service.AssertExpectations(t)
//		require.NoError(t, err)
//		assert.Equal(t, expected, result)
//	})
//
//	t.Run("Should return an error", func(t *testing.T) {
//		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
//
//		converter := automock.NewApiConv()
//		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()
//
//		service := automock.NewApiSvc()
//		service.On("Create", api).Return(nil, errors.New("test")).Once()
//
//		resolver, err := newApiResolver(service)
//		require.NoError(t, err)
//
//		_, err = resolver.CreateAPI(nil, name, namespace, params)
//
//		service.AssertExpectations(t)
//		require.Error(t, err)
//		assert.True(t, gqlerror.IsInternal(err))
//	})
//
//}
//
//func TestApiResolver_UpdateAPI(t *testing.T) {
//	namespace := "test-1"
//	name := "test-api"
//	hostname := "test-hostname"
//	serviceName := "test-service-name"
//	servicePort := 8080
//	jwksUri := "http://test-jwks-uri"
//	issuer := "test-issuer"
//
//	params := paramsToAPICreationInput(hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)
//
//	t.Run("Should update an API", func(t *testing.T) {
//		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
//		expected := testApiToGQL(name, hostname, serviceName, jwksUri, issuer, servicePort)
//
//		converter := automock.NewApiConv()
//		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()
//
//		service := automock.NewApiSvc()
//		service.On("Update", api).Return(api, nil).Once()
//
//		resolver, err := newApiResolver(service)
//		require.NoError(t, err)
//
//		result, err := resolver.UpdateAPI(nil, name, namespace, params)
//
//		service.AssertExpectations(t)
//		require.NoError(t, err)
//		assert.Equal(t, expected, result)
//	})
//
//	t.Run("Should return an error", func(t *testing.T) {
//		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
//
//		converter := automock.NewApiConv()
//		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()
//
//		service := automock.NewApiSvc()
//		service.On("Update", api).Return(nil, errors.New("test")).Once()
//
//		resolver, err := newApiResolver(service)
//		require.NoError(t, err)
//
//		_, err = resolver.UpdateAPI(nil, name, namespace, params)
//
//		service.AssertExpectations(t)
//		require.Error(t, err)
//		assert.True(t, gqlerror.IsInternal(err))
//	})
//
//}

func TestApiRuleResolver_DeleteAPIRule(t *testing.T) {
	t.Run("Should delete an APIRule", func(t *testing.T) {
		apiRule := fixTestApiRule(name, hostname1, serviceName1, servicePort1, gateway1)

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
		apiRule := fixTestApiRule(name, hostname1, serviceName1, servicePort1, gateway1)

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

func fixTestApiRule(ruleName string, hostName string, serviceName string, servicePort uint32, gateway string) *v1alpha1.APIRule {
	return &v1alpha1.APIRule{
		ObjectMeta: v1.ObjectMeta{
			Name: ruleName,
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
							},
						},
					},
				},
			},
		},
		Status: v1alpha1.APIRuleStatus{
			APIRuleStatus: &v1alpha1.APIRuleResourceStatus{
				Code: "OK",
			},
			VirtualServiceStatus: &v1alpha1.APIRuleResourceStatus{
				Code: "OK",
			},
			AccessRuleStatus: &v1alpha1.APIRuleResourceStatus{
				Code: "OK",
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
		Status: &gqlschema.APIRuleStatuses{
			APIRuleStatus: gqlschema.APIRuleStatus{
				Code: "OK",
			},
			VirtualServiceStatus: gqlschema.APIRuleStatus{
				Code: "OK",
			},
			AccessRuleStatus: gqlschema.APIRuleStatus{
				Code: "OK",
			},
		},
	}
}
