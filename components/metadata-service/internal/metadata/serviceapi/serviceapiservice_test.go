package serviceapi

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"testing"

	k8smocks "github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts/mocks"
	asmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/accessservice/mocks"
	istiomocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/istio/mocks"
	secretsmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/secrets/mocks"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	resourceName = "resource-uuid-1"
	gatewayUrl   = "url-uuid-1"
)

func TestNewService(t *testing.T) {
	t.Run("should add all required components for API with OAuth credentials", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		remoteEnvCredentials := remoteenv.Credentials{
			Type:              remoteenv.CredentialsOAuthType,
			SecretName:        resourceName,
			AuthenticationUrl: api.Credentials.Oauth.URL,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteEnvCredentials, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", "re", "uuid-1", resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		remoteEnvServiceAPI, err := service.New("re", "uuid-1", api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, remoteEnvServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, remoteEnvServiceAPI.TargetUrl)
		assert.Equal(t, api.Credentials.Oauth.URL, remoteEnvServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "OAuth", remoteEnvServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.Credentials.SecretName)

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should add all required components for API with BasicAuth credentials", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		remoteEnvCredentials := remoteenv.Credentials{
			Type:       remoteenv.CredentialsBasicType,
			SecretName: resourceName,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteEnvCredentials, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", "re", "uuid-1", resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		remoteEnvServiceAPI, err := service.New("re", "uuid-1", api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, remoteEnvServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, remoteEnvServiceAPI.TargetUrl)
		assert.Equal(t, "Basic", remoteEnvServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.Credentials.SecretName)

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
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", "re", "uuid-1", resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, nil, istioService)

		// when
		remoteEnvServiceAPI, err := service.New("re", "uuid-1", api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, remoteEnvServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.AccessLabel)
		assert.Equal(t, api.TargetUrl, remoteEnvServiceAPI.TargetUrl)
		assert.Equal(t, "", remoteEnvServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "", remoteEnvServiceAPI.Credentials.SecretName)

		accessServiceManager.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should return error when creating access service fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, nil, nil)

		// when
		result, err := service.New("re", "uuid-1", api)

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
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteenv.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil)

		// when
		result, err := service.New("re", "uuid-1", api)

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
			Credentials: &model.Credentials{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteenv.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil)

		// when
		result, err := service.New("re", "uuid-1", api)

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
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Create", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Create",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteenv.Credentials{}, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Create", "re", "uuid-1", resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		result, err := service.New("re", "uuid-1", api)

		// then
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})
}

func TestDefaultService_Read(t *testing.T) {
	t.Run("should read API with OAuth credentials", func(t *testing.T) {
		// given
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com",
				SecretName:        "secret-name",
				Type:              remoteenv.CredentialsOAuthType,
			},
		}

		credentials := model.Credentials{
			Oauth: &model.Oauth{
				ClientID:     "clientId",
				ClientSecret: "clientSecret",
				URL:          "http://oauth.com",
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", "re", remoteEnvServiceAPi.Credentials).Return(credentials, nil)

		service := NewService(nil, nil, secretsService, nil)

		// when
		api, err := service.Read("re", remoteEnvServiceAPi)

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
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: remoteenv.Credentials{
				SecretName: "secret-name",
				Type:       remoteenv.CredentialsBasicType,
			},
		}

		credentials := model.Credentials{
			Basic: &model.Basic{
				Username: "clientUsername",
				Password: "clientPassword",
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", "re", remoteEnvServiceAPi.Credentials).Return(credentials, nil)

		service := NewService(nil, nil, secretsService, nil)

		// when
		api, err := service.Read("re", remoteEnvServiceAPi)

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
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
		}

		service := NewService(nil, nil, nil, nil)

		// when
		api, err := service.Read("re", remoteEnvServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Nil(t, api.Credentials)
		assert.Nil(t, api.Spec)
	})

	t.Run("should return error when reading OAuth secret fails", func(t *testing.T) {
		// given
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com",
				SecretName:        "secret-name",
				Type:              remoteenv.CredentialsOAuthType,
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", "re", remoteEnvServiceAPi.Credentials).
			Return(model.Credentials{}, apperrors.Internal("secret error"))

		service := NewService(nil, nil, secretsService, nil)

		// when
		api, err := service.Read("re", remoteEnvServiceAPi)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "secret error")

		secretsService.AssertExpectations(t)
	})

	t.Run("should return error when reading BasicAuth secret fails", func(t *testing.T) {
		// given
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com",
				SecretName:        "secret-name",
				Type:              remoteenv.CredentialsBasicType,
			},
		}

		secretsService := new(secretsmocks.Service)
		secretsService.On("Get", "re", remoteEnvServiceAPi.Credentials).
			Return(model.Credentials{}, apperrors.Internal("secret error"))

		service := NewService(nil, nil, secretsService, nil)

		// when
		api, err := service.Read("re", remoteEnvServiceAPi)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "secret error")

		secretsService.AssertExpectations(t)
	})
}

func TestDefaultService_Delete(t *testing.T) {
	t.Run("should delete an API", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Delete", resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.NoError(t, err)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should return an error if accessService deletion fails", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(apperrors.Internal("an error"))

		service := NewService(nameResolver, accessServiceManager, nil, nil)

		// when
		err := service.Delete("re", "uuid-1")

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
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(apperrors.Internal("an error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
	})

	t.Run("should return an error if istio deletion fails", func(t *testing.T) {
		// given
		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Delete", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Delete", resourceName).Return(apperrors.Internal(""))

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})
}

func TestDefaultService_Update(t *testing.T) {
	t.Run("should update an API with a new one containing an OAuth secret", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		remoteEnvCredentials := remoteenv.Credentials{
			Type:              remoteenv.CredentialsOAuthType,
			SecretName:        resourceName,
			AuthenticationUrl: api.Credentials.Oauth.URL,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Update",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteEnvCredentials, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		remoteEnvServiceAPI, err := service.Update("re", "uuid-1", api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, remoteEnvServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.AccessLabel)
		assert.Equal(t, "http://target.com", remoteEnvServiceAPI.TargetUrl)
		assert.Equal(t, "http://oauth.com", remoteEnvServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "OAuth", remoteEnvServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.Credentials.SecretName)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should update an API with a new one containing a BasicAuth secret", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		remoteEnvCredentials := remoteenv.Credentials{
			Type:       remoteenv.CredentialsBasicType,
			SecretName: resourceName,
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Update",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteEnvCredentials, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		remoteEnvServiceAPI, err := service.Update("re", "uuid-1", api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, remoteEnvServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.AccessLabel)
		assert.Equal(t, "http://target.com", remoteEnvServiceAPI.TargetUrl)
		assert.Equal(t, "Basic", remoteEnvServiceAPI.Credentials.Type)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.Credentials.SecretName)

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
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		service := NewService(
			nameResolver,
			accessServiceManager,
			secretsService,
			istioService,
		)

		// when
		remoteEnvServiceAPI, err := service.Update("re", "uuid-1", api)

		// then
		require.NoError(t, err)
		assert.Equal(t, gatewayUrl, remoteEnvServiceAPI.GatewayURL)
		assert.Equal(t, resourceName, remoteEnvServiceAPI.AccessLabel)
		assert.Equal(t, "http://target.com", remoteEnvServiceAPI.TargetUrl)
		assert.Equal(t, "", remoteEnvServiceAPI.Credentials.AuthenticationUrl)
		assert.Equal(t, "", remoteEnvServiceAPI.Credentials.SecretName)

		nameResolver.AssertExpectations(t)
		accessServiceManager.AssertExpectations(t)
		secretsService.AssertExpectations(t)
		istioService.AssertExpectations(t)
	})

	t.Run("should return error when updating access service fails", func(t *testing.T) {
		// given
		api := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).
			Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, nil, nil)

		// when
		result, err := service.Update("re", "uuid-1", api)

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
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Update",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteenv.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil)

		// when
		result, err := service.Update("re", "uuid-1", api)

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
			Credentials: &model.Credentials{
				Basic: &model.Basic{
					Username: "clientUsername",
					Password: "clientPassword",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Update",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteenv.Credentials{}, apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil)

		// when
		result, err := service.Update("re", "uuid-1", api)

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
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On("Delete", resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, nil)

		// when
		result, err := service.Update("re", "uuid-1", api)

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
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("GetResourceName", "re", "uuid-1").Return(resourceName)
		nameResolver.On("GetGatewayUrl", "re", "uuid-1").Return(gatewayUrl)

		accessServiceManager := new(asmocks.AccessServiceManager)
		accessServiceManager.On("Upsert", "re", "uuid-1", resourceName).Return(nil)

		secretsService := new(secretsmocks.Service)
		secretsService.On(
			"Update",
			"re",
			"uuid-1",
			api.Credentials,
		).Return(remoteenv.Credentials{}, nil)

		istioService := new(istiomocks.Service)
		istioService.On("Upsert", "re", "uuid-1", resourceName).Return(apperrors.Internal("some error"))

		service := NewService(nameResolver, accessServiceManager, secretsService, istioService)

		// when
		result, err := service.Update("re", "uuid-1", api)

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
}
