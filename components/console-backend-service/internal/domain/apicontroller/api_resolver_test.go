package apicontroller

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiResolver_APIsQuery(t *testing.T) {
	namespace := "test-1"

	t.Run("Should return a list of APIs for namespace", func(t *testing.T) {
		apis := []*v1alpha2.Api{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-1",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-2",
				},
			},
		}

		expected := []gqlschema.API{
			{
				Name: apis[0].Name,
			},
			{
				Name: apis[1].Name,
			},
		}

		var empty *string = nil

		service := automock.NewApiSvc()
		service.On("List", namespace, empty, empty).Return(apis, nil).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		result, err := resolver.APIsQuery(nil, namespace, nil, nil)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		var empty *string = nil

		service := automock.NewApiSvc()
		service.On("List", namespace, empty, empty).Return(nil, errors.New("test")).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		_, err = resolver.APIsQuery(nil, namespace, nil, nil)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestApiResolver_APIQuery(t *testing.T) {
	namespace := "test-1"
	name := "test-api"

	t.Run("Should return a an API in namespace", func(t *testing.T) {
		api := v1alpha2.Api{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-1",
			},
		}

		expected := gqlschema.API{
			Name: api.Name,
		}

		service := automock.NewApiSvc()
		service.On("Find", name, namespace).Return(&api, nil).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		result, err := resolver.APIQuery(nil, name, namespace)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Should return an empty object", func(t *testing.T) {
		service := automock.NewApiSvc()
		service.On("Find", name, namespace).Return(nil, nil).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		result, err := resolver.APIQuery(nil, name, namespace)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, (*gqlschema.API)(nil), result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		service := automock.NewApiSvc()
		service.On("Find", name, namespace).Return(nil, errors.New("test")).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		_, err = resolver.APIQuery(nil, name, namespace)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestApiResolver_CreateAPI(t *testing.T) {
	namespace := "test-1"
	name := "test-api"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"

	params := paramsToAPICreationInput(hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)

	t.Run("Should create an API", func(t *testing.T) {
		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
		expected := testApiToGQL(name, hostname, serviceName, jwksUri, issuer, servicePort)

		converter := automock.NewApiConv()
		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()

		service := automock.NewApiSvc()
		service.On("Create", api).Return(api, nil).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		result, err := resolver.CreateAPI(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)

		converter := automock.NewApiConv()
		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()

		service := automock.NewApiSvc()
		service.On("Create", api).Return(nil, errors.New("test")).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		_, err = resolver.CreateAPI(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

}

func TestApiResolver_UpdateAPI(t *testing.T) {
	namespace := "test-1"
	name := "test-api"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"

	params := paramsToAPICreationInput(hostname, serviceName, jwksUri, issuer, servicePort, nil, nil)

	t.Run("Should update an API", func(t *testing.T) {
		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
		expected := testApiToGQL(name, hostname, serviceName, jwksUri, issuer, servicePort)

		converter := automock.NewApiConv()
		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()

		service := automock.NewApiSvc()
		service.On("Update", api).Return(api, nil).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		result, err := resolver.UpdateAPI(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Should return an error", func(t *testing.T) {
		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)

		converter := automock.NewApiConv()
		converter.On("ToApi", name, namespace, params).Return(api, nil).Once()

		service := automock.NewApiSvc()
		service.On("Update", api).Return(nil, errors.New("test")).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		_, err = resolver.UpdateAPI(nil, name, namespace, params)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

}

func TestApiResolver_DeleteAPI(t *testing.T) {
	namespace := "test-1"
	name := "test-api"
	hostname := "test-hostname"
	serviceName := "test-service-name"
	servicePort := 8080
	jwksUri := "http://test-jwks-uri"
	issuer := "test-issuer"

	t.Run("Should update an API", func(t *testing.T) {
		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)
		expected := testApiToGQL(name, hostname, serviceName, jwksUri, issuer, servicePort)

		service := automock.NewApiSvc()
		service.On("Find", name, namespace).Return(api, nil).Once()
		service.On("Delete", name, namespace).Return(nil).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		result, err := resolver.DeleteAPI(nil, name, namespace)

		service.AssertExpectations(t)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Should return an error if api has not been found", func(t *testing.T) {
		service := automock.NewApiSvc()
		service.On("Find", name, namespace).Return(nil, errors.New("test")).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		_, err = resolver.DeleteAPI(nil, name, namespace)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Should return an error if api couldnt be removed", func(t *testing.T) {
		api := fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer, servicePort)

		service := automock.NewApiSvc()
		service.On("Find", name, namespace).Return(api, nil).Once()
		service.On("Delete", name, namespace).Return(errors.New("test")).Once()

		resolver, err := newApiResolver(service)
		require.NoError(t, err)

		_, err = resolver.DeleteAPI(nil, name, namespace)

		service.AssertExpectations(t)
		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

}

func fixTestApi(name, namespace, hostname, serviceName, jwksUri, issuer string, servicePort int) *v1alpha2.Api {
	api := v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha2",
			Kind:       "API",
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
		},
	}
	return &api
}

func testApiToGQL(name, hostname, serviceName, jwksUri, issuer string, servicePort int) gqlschema.API {
	gql := gqlschema.API{
		Name:     name,
		Hostname: hostname,
		Service: gqlschema.ApiService{
			Name: serviceName,
			Port: servicePort,
		},
		AuthenticationPolicies: []gqlschema.AuthenticationPolicy{
			{
				JwksURI: jwksUri,
				Issuer:  issuer,
				Type:    gqlschema.AuthenticationPolicyType("JWT"),
			},
		},
	}
	return gql
}

func paramsToAPICreationInput(hostname, serviceName, jwksUri, issuer string, servicePort int, disableIstioAuthPolicyMTLS, authenticationEnabled *bool) gqlschema.APIInput {
	return gqlschema.APIInput{
		Hostname:                   hostname,
		ServiceName:                serviceName,
		ServicePort:                servicePort,
		JwksURI:                    jwksUri,
		Issuer:                     issuer,
		DisableIstioAuthPolicyMTLS: disableIstioAuthPolicyMTLS,
		AuthenticationEnabled:      authenticationEnabled,
	}
}
