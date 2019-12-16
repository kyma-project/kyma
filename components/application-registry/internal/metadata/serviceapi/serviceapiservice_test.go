package serviceapi

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	k8smocks "github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts/mocks"
	asmocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/accessservice/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	istiomocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/istio/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	secretsmocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	resourceName = "resource-uuid-1"
	gatewayUrl   = "url-uuid-1"

	appName   = "app"
	appUID    = types.UID("appUID")
	serviceId = "uuid-1"
)

func TestNewService(t *testing.T) {
	t.Run("should add all required components for API with OAuth credentials", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		applicationCredentials := applications.Credentials{
			Type:              applications.CredentialsOAuthType,
			SecretName:        resourceName,
			AuthenticationUrl: api.Credentials.Oauth.URL,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			appName,
			appUID,
			serviceId,
			api.Credentials,
		).Return(applicationCredentials, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, istioService)

		// when
		applicationServiceAPI, err := service.New(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, applicationServiceAPI.TargetUrl)
		assert.Equal(t, api.Credentials.Oauth.URL, applicationServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "OAuth", applicationServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, applicationServiceAPI.Credentials.SecretName)

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should add all required components for API with BasicAuth credentials", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		applicationCredentials := applications.Credentials{
			Type:       applications.CredentialsBasicType,
			SecretName: resourceName,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			appName,
			appUID,
			serviceId,
			api.Credentials,
		).Return(applicationCredentials, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, istioService)

		// when
		applicationServiceAPI, err := service.New(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, applicationServiceAPI.TargetUrl)
		assert.Equal(t, "Basic", applicationServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, applicationServiceAPI.Credentials.SecretName)

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should add all required components for API without credentials", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, nil, nil, istioService)

		// when
		applicationServiceAPI, err := service.New(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, applicationServiceAPI.TargetUrl)
		assert.Equal(t, "", applicationServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "", applicationServiceAPI.Credentials.SecretName)

		accessServiceManager.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should return error when creating access service fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, nil, nil, nil)

		// when
		result, err := service.New(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		accessServiceManager.AssertExpectations(t)
	})

	t.Run("should return error when creating OAuth secret fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			appName,
			appUID,
			serviceId,
			api.Credentials,
		).Return(applications.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, nil)

		// when
		result, err := service.New(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when creating BasicAuth secret fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			appName,
			appUID,
			serviceId,
			api.Credentials,
		).Return(applications.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, nil)

		// when
		result, err := service.New(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when creating istio resources fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			appName,
			appUID,
			serviceId,
			api.Credentials,
		).Return(applications.Credentials{}, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", appName, appUID, serviceId, resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, istioService)

		// when
		result, err := service.New(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should add secret name of request parameters", func(t *testing.T) {
		// given
		requestParamsSecretName := "params-app-id"

		requestParams := &model.RequestParameters{
			Headers: &map[string][]string{
				"HeaderName": {"headerVal1", "headerVal2"},
			},
			QueryParameters: &map[string][]string{
				"queryParamName": {"queryVal1", "queryVal2"},
			},
		}

		api := &model.API{
			TargetUrl:         "http://target.com",
			RequestParameters: requestParams,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)

		requestParamsService := new(secretsmocks.RequestParametersService)
		requestParamsService.On("Create", appName, appUID, serviceId, requestParams).Return(requestParamsSecretName, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		applicationServiceAPI, err := service.New(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, applicationServiceAPI.TargetUrl)
		assert.Equal(t, requestParamsSecretName, applicationServiceAPI.RequestParametersSecretName)

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})
}

func TestDefaultService_Read(t *testing.T) {
	t.Run("should read API with OAuth credentials", func(t *testing.T) {
		// given
		applicationServiceAPi := &applications.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: applications.Credentials{
				AuthenticationUrl: "http://oauth.com",
				SecretName:        "secret-name",
				Type:              applications.CredentialsOAuthType,
			},
		}

		credentials := model.CredentialsWithCSRF{
			Oauth: &model.Oauth{
				ClientID:     "clientId",
				ClientSecret: "clientSecret",
				URL:          "http://oauth.com",
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", appName, applicationServiceAPi.Credentials).Return(credentials, nil)

		service := NewService(nil, nil, secretsService, nil, nil)

		// when
		api, err := service.Read(appName, applicationServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Equal(t, "http://oauth.com", api.Credentials.Oauth.URL)
		assert.Equal(t, "clientId", api.Credentials.Oauth.ClientID)
		assert.Equal(t, "clientSecret", api.Credentials.Oauth.ClientSecret)
		assert.Nil(t, api.Spec)

		secretsService.AssertExpectations(t)
	})

	t.Run("should read API with BasicAuth credentials", func(t *testing.T) {
		// given
		applicationServiceAPi := &applications.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: applications.Credentials{
				SecretName: "secret-name",
				Type:       applications.CredentialsBasicType,
			},
		}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "clientUsername",
				Password: "clientPassword",
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", appName, applicationServiceAPi.Credentials).Return(credentials, nil)

		service := NewService(nil, nil, secretsService, nil, nil)

		// when
		api, err := service.Read(appName, applicationServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Equal(t, "clientUsername", api.Credentials.Basic.Username)
		assert.Equal(t, "clientPassword", api.Credentials.Basic.Password)
		assert.Nil(t, api.Spec)

		secretsService.AssertExpectations(t)
	})

	t.Run("should read API without credentials", func(t *testing.T) {
		// given
		applicationServiceAPi := &applications.ServiceAPI{
			TargetUrl: "http://target.com",
		}

		service := NewService(nil, nil, nil, nil, nil)

		// when
		api, err := service.Read(appName, applicationServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Nil(t, api.Credentials)
		assert.Nil(t, api.Spec)
	})

	t.Run("should return error when reading OAuth secret fails", func(t *testing.T) {
		// given
		applicationServiceAPi := &applications.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: applications.Credentials{
				AuthenticationUrl: "http://oauth.com",
				SecretName:        "secret-name",
				Type:              applications.CredentialsOAuthType,
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", appName, applicationServiceAPi.Credentials).
			Return(model.CredentialsWithCSRF{}, apperrors.Internal("secret error"))

		service := NewService(nil, nil, secretsService, nil, nil)

		// when
		api, err := service.Read(appName, applicationServiceAPi)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "secret error")

		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when reading BasicAuth secret fails", func(t *testing.T) {
		// given
		applicationServiceAPi := &applications.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: applications.Credentials{
				AuthenticationUrl: "http://oauth.com",
				SecretName:        "secret-name",
				Type:              applications.CredentialsBasicType,
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", appName, applicationServiceAPi.Credentials).
			Return(model.CredentialsWithCSRF{}, apperrors.Internal("secret error"))

		service := NewService(nil, nil, secretsService, nil, nil)

		// when
		api, err := service.Read(appName, applicationServiceAPi)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "secret error")

		secretsService.AssertExpectations(t)
	})

	t.Run("should read API with request parameters", func(t *testing.T) {
		// given
		requestParamsSecretName := "params-app-id"

		requestParams := &model.RequestParameters{
			Headers: &map[string][]string{
				"HeaderName": {"headerVal1", "headerVal2"},
			},
			QueryParameters: &map[string][]string{
				"queryParamName": {"queryVal1", "queryVal2"},
			},
		}

		applicationServiceAPi := &applications.ServiceAPI{
			TargetUrl:                   "http://target.com",
			RequestParametersSecretName: requestParamsSecretName,
		}

		secretsService := new(secretsmocks.Service)

		requestParamsService := new(secretsmocks.RequestParametersService)
		requestParamsService.On("Get", requestParamsSecretName).Return(requestParams, nil)

		service := NewService(nil, nil, secretsService, requestParamsService, nil)

		// when
		api, err := service.Read(appName, applicationServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Equal(t, requestParams, api.RequestParameters)
		assert.Nil(t, api.Spec)

		secretsService.AssertExpectations(t)
	})

}

func TestDefaultService_Delete(t *testing.T) {
	t.Run("should delete an API", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Delete", resourceName).Return(nil)

		requestParamsService := new(secretsmocks.RequestParametersService)
		requestParamsService.On("Delete", appName, serviceId).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		err := service.Delete(appName, serviceId)

		// then
		assert.NoError(t, err)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
		requestParamsService.AssertExpectations(t)
	})

	t.Run("should return an error if accessService deletion fails", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(apperrors.Internal("an error"))

		service := NewService(nameResolver, accessServiceManager, nil, nil, nil)

		// when
		err := service.Delete(appName, serviceId)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
	})

	t.Run("should return an error if secret deletion fails", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(apperrors.Internal("an error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, nil)

		// when
		err := service.Delete(appName, serviceId)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return an error if request params secret fails", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Delete", appName, serviceId).Return(apperrors.Internal("error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, nil)

		// when
		err := service.Delete(appName, serviceId)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		requestParamsService.AssertExpectations(t)
	})

	t.Run("should return an error if istio deletion fails", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Delete", resourceName).Return(apperrors.Internal(""))

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Delete", appName, serviceId).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		err := service.Delete(appName, serviceId)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
		requestParamsService.AssertExpectations(t)
	})
}

func TestDefaultService_Update(t *testing.T) {

	t.Run("should update an API with a new one containing an OAuth secret", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		applicationCredentials := applications.Credentials{
			Type:              applications.CredentialsOAuthType,
			SecretName:        resourceName,
			AuthenticationUrl: api.Credentials.Oauth.URL,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Upsert", appName, appUID, serviceId, api.Credentials).Return(applicationCredentials, nil)

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Delete", appName, serviceId).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		applicationServiceAPI, err := service.Update(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, "http://target.com", applicationServiceAPI.TargetUrl)
		assert.Equal(t, "http://oauth.com", applicationServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "OAuth", applicationServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, applicationServiceAPI.Credentials.SecretName)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should update an API with a new one containing a BasicAuth secret", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		applicationCredentials := applications.Credentials{
			Type:       applications.CredentialsBasicType,
			SecretName: resourceName,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Upsert", appName, appUID, serviceId, api.Credentials).Return(applicationCredentials, nil)

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Delete", appName, serviceId).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		applicationServiceAPI, err := service.Update(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, "http://target.com", applicationServiceAPI.TargetUrl)
		assert.Equal(t, "Basic", applicationServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, applicationServiceAPI.Credentials.SecretName)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should update an API with a new one not containing a secret", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl:   "http://target.com",
			Credentials: nil,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Delete", appName, serviceId).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		applicationServiceAPI, err := service.Update(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, "http://target.com", applicationServiceAPI.TargetUrl)
		assert.Equal(t, "", applicationServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "", applicationServiceAPI.Credentials.SecretName)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should return error when updating access service fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).
			Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, nil, nil, nil)

		// when
		result, err := service.Update(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
	})

	t.Run("should return error when updating OAuth secret fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Upsert", appName, appUID, serviceId, api.Credentials).Return(applications.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, nil)

		// when
		result, err := service.Update(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when updating BasicAuth secret fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Upsert", appName, appUID, serviceId, api.Credentials).Return(applications.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, nil)

		// when
		result, err := service.Update(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when deleting secret fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl:   "http://target.com",
			Credentials: nil,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil, nil)

		// when
		result, err := service.Update(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when updating istio resources fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Upsert", appName, appUID, serviceId, api.Credentials).Return(applications.Credentials{}, nil)

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Delete", appName, serviceId).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", appName, appUID, serviceId, resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		result, err := service.Update(appName, appUID, serviceId, api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should update an API adding request parameters", func(t *testing.T) {
		// given
		requestParamsSecretName := "params-app-id"

		requestParams := &model.RequestParameters{
			Headers: &map[string][]string{
				"HeaderName": {"headerVal1", "headerVal2"},
			},
			QueryParameters: &map[string][]string{
				"queryParamName": {"queryVal1", "queryVal2"},
			},
		}

		api := &model.API{
			TargetUrl:         "http://target.com",
			RequestParameters: requestParams,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", appName, serviceId).Return(resourceName)
		nameResolver.On("GetGatewayUrl", appName, serviceId).Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		requestParamsService := &secretsmocks.RequestParametersService{}
		requestParamsService.On("Upsert", appName, appUID, serviceId, requestParams).Return(requestParamsSecretName, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", appName, appUID, serviceId, resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, requestParamsService, istioService)

		// when
		applicationServiceAPI, err := service.Update(appName, appUID, serviceId, api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, applicationServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, applicationServiceAPI.AccessLabel)
		assert.Equal(t, requestParamsSecretName, applicationServiceAPI.RequestParametersSecretName)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})
}
