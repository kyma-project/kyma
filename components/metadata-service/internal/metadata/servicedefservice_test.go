package metadata

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"testing"

	"bytes"
	"encoding/json"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	miniomocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/minio/mocks"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	remoteenvmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv/mocks"
	serviceapimocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi/mocks"
	uuidmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	empty []byte
)

func TestServiceDefinitionService_Create(t *testing.T) {

	t.Run("should create service with API, events and documentation", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("{\"api\":\"spec\"}"),
		}
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         serviceAPI,
			Labels:      &map[string]string{"connected-app": "re"},
			Identifier:  "Some cool external identifier",
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}
		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Identifier:          "Some cool external identifier",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              true,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("New", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		serviceRepository.On("GetAll", "re").Return(nil, nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte("{\"api\":\"spec\"}"), []byte("events spec")).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should create service without API", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         nil,
			Events: &model.Events{
				Spec: []byte("test"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", mock.Anything, empty, []byte("test")).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should create service with documentation only", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:          "Some service",
			Description:   "Some cool service",
			Provider:      "Service Provider",
			Api:           nil,
			Events:        nil,
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              false,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", mock.Anything, empty, empty).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should create service without specs", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:          "Some service",
			Description:   "Some cool service",
			Provider:      "Service Provider",
			Api:           nil,
			Events:        nil,
			Documentation: nil,
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              false,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", empty, empty, empty).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should create service with modified API spec", func(t *testing.T) {
		// given
		initialAPISpec := []byte("{\"swagger\":\"2.0\"}")
		expectedAPISpec := compact([]byte(`{"schemes":["http"],"swagger":"2.0","host":"gateway-url.kyma.local","paths":null}`))

		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: initialAPISpec,
		}
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "http://gateway-url.kyma.local",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}
		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              true,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("New", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), expectedAPISpec, []byte("events spec")).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should override connected-app label", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:          "Some service",
			Description:   "Some cool service",
			Provider:      "Service Provider",
			Labels:        &map[string]string{"connected-app": "wrong-re"},
			Api:           nil,
			Events:        nil,
			Documentation: nil,
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              false,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", empty, empty, empty).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should create connected-app label if not provided", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:          "Some service",
			Description:   "Some cool service",
			Provider:      "Service Provider",
			Api:           nil,
			Events:        nil,
			Documentation: nil,
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              false,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", empty, empty, empty).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return error when adding API fails", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
		}
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         serviceAPI,
		}

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("New", "re", "uuid-1", serviceAPI).Return(nil, apperrors.Internal("some error"))

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, nil, nil)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		assert.Empty(t, serviceID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		serviceAPIService.AssertExpectations(t)
	})

	t.Run("should return error when failed to unmarshal api spec", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("invalid spec"),
		}
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("New", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		minioService := new(miniomocks.Service)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Equal(t, "", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
	})

	t.Run("should return error when adding spec to Minio fails", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         nil,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: nil,
		}

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", empty, empty, []byte("events spec")).Return(apperrors.Internal("Error"))

		service := NewServiceDefinitionService(uuidGenerator, nil, nil, minioService)

		// when
		_, err := service.Create("re", &serviceDefinition)

		// then
		require.Error(t, err)

		uuidGenerator.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return error when creating service in remote environment fails", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
		}
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         serviceAPI,
		}
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-utr",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "",
				SecretName:        "",
			},
		}
		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              false,
		}
		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("New", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(apperrors.Internal("some error"))
		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", empty, empty, empty).Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		assert.Empty(t, serviceID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return an error when identifier conflict occurs", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:          "Some service",
			Description:   "Some cool service",
			Provider:      "Service Provider",
			Identifier:    "Same identifier",
			Api:           nil,
			Events:        nil,
			Documentation: nil,
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Identifier:          "Same identifier",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              false,
		}
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("GetAll", "re").Return([]remoteenv.Service{remoteEnvService}, nil)

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.Empty(t, serviceID)
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeAlreadyExists, err.Code())

		serviceRepository.AssertExpectations(t)
	})
}

func TestServiceDefinitionService_GetAll(t *testing.T) {

	t.Run("should get all services", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("GetAll", "re").Return([]remoteenv.Service{
			{
				ID:                  "uuid-1",
				DisplayName:         "Service1",
				LongDescription:     "Service1 description",
				ProviderDisplayName: "Service1 Provider",
				Labels:              map[string]string{"connected-app": "re"},
				Tags:                nil,
				API: &remoteenv.ServiceAPI{
					TargetUrl: "http://service1.com",
					Credentials: remoteenv.Credentials{
						SecretName: "testSecret1",
					},
				},
				Events: false,
			},
			{
				ID:                  "uuid-2",
				DisplayName:         "Service2",
				LongDescription:     "Service2 description",
				Labels:              map[string]string{"connected-app": "re"},
				ProviderDisplayName: "Service2 Provider",
				Tags:                nil,
				API:                 nil,
				Events:              true,
			},
		}, nil)

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAll("re")
		require.NoError(t, err)

		// then
		assert.Len(t, result, 2)
		assert.Contains(t, result, model.ServiceDefinition{
			ID:          "uuid-1",
			Name:        "Service1",
			Labels:      &map[string]string{"connected-app": "re"},
			Description: "Service1 description",
			Provider:    "Service1 Provider",
		})
		assert.Contains(t, result, model.ServiceDefinition{
			ID:          "uuid-2",
			Name:        "Service2",
			Labels:      &map[string]string{"connected-app": "re"},
			Description: "Service2 description",
			Provider:    "Service2 Provider",
		})
	})

	t.Run("should get empty list", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("GetAll", "re").Return([]remoteenv.Service{}, nil)

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAll("re")
		require.NoError(t, err)

		// then
		assert.Len(t, result, 0)
	})
}

func TestServiceDefinitionService_GetById(t *testing.T) {

	t.Run("should get service by ID", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}

		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ProviderDisplayName: "Service Provider",
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              false,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(serviceAPI, nil)
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		minioService := new(miniomocks.Service)
		minioService.On("Get", "uuid-1").Return(empty, empty, empty, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, minioService)

		// when
		result, err := service.GetByID("re", "uuid-1")
		require.NoError(t, err)

		// then
		assert.Equal(t, "uuid-1", result.ID)
		assert.Equal(t, "Some service", result.Name)
		assert.Equal(t, "Some cool service", result.Description)
		assert.Equal(t, "Service Provider", result.Provider)
		assert.Equal(t, "http://target.com", result.Api.TargetUrl)
		assert.Equal(t, "http://oauth.com/token", result.Api.Credentials.Oauth.URL)
		assert.Equal(t, "clientId", result.Api.Credentials.Oauth.ClientID)
		assert.Equal(t, "clientSecret", result.Api.Credentials.Oauth.ClientSecret)
		assert.Nil(t, result.Events)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return error when getting service from remote environment fails", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.Internal("get error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get error")
	})

	t.Run("should return error when reading API fails", func(t *testing.T) {
		// given
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ProviderDisplayName: "Service Provider",
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              false,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(nil, apperrors.Internal("api error"))
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		minioService := new(miniomocks.Service)
		minioService.On("Get", "uuid-1").Return(empty, empty, empty, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api error")
	})

	t.Run("should return error when reading specs from Minio fails", func(t *testing.T) {
		// given
		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ProviderDisplayName: "Service Provider",
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              false,
		}

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		minioService := new(miniomocks.Service)
		minioService.On("Get", "uuid-1").Return(empty, empty, empty, apperrors.Internal("error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, minioService)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error")

		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})
}

func TestServiceDefinitionService_Update(t *testing.T) {

	t.Run("should update a service", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("{\"api\":\"spec\"}"),
		}

		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "Identifier",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",

			Credentials: remoteenv.Credentials{
				SecretName: "secret-name",
			},
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			Identifier:          "Identifier",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Update", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(serviceAPI, nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		serviceRepository.On("Update", "re", remoteEnvService).Return(nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte("{\"api\":\"spec\"}"), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return not found when update a not existing service", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("{\"api\":\"spec\"}"),
		}

		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Update", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(serviceAPI, nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.NotFound("missing"))

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte("{\"api\":\"spec\"}"), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})

	t.Run("should update a service when no API was given", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "Identifier",
			Api:         nil,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			Identifier:          "Identifier",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		serviceRepository.On("Update", "re", remoteEnvService).Return(nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte(nil), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should preserve a service identifier", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "DifferentIdentifier",
			Api:         nil,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Identifier:          "ServiceIdentifier",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		serviceRepository.On("Update", "re", remoteEnvService).Return(nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte(nil), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return an error if cache initialization failed", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("api docs"),
		}

		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "Identifier",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			Identifier:          "Identifier",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Update", "re", "uuid-1", serviceAPI).Return(&remoteenv.ServiceAPI{}, apperrors.Internal("an error"))

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte(nil), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceRepository.AssertExpectations(t)
	})

	t.Run("should return an error if API update failed", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("api docs"),
		}

		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "Identifier",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			Identifier:          "Identifier",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Update", "re", "uuid-1", serviceAPI).Return(&remoteenv.ServiceAPI{}, apperrors.Internal("an error"))

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte(nil), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
	})

	t.Run("should return an error if Minio data update failed", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("{\"api\":\"spec\"}"),
		}

		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "Identifier",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			Identifier:          "Identifier",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Update", "re", "uuid-1", serviceAPI).Return(&remoteenv.ServiceAPI{}, nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte("{\"api\":\"spec\"}"), []byte("events spec")).Return(apperrors.Internal("an error"))
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return an error if remoteenv update failed", func(t *testing.T) {
		// given
		serviceAPI := &model.API{
			TargetUrl: "http://target.com",
			Credentials: &model.Credentials{
				Oauth: &model.Oauth{
					URL:          "http://oauth.com/token",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: []byte("{\"api\":\"spec\"}"),
		}

		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
			Identifier:  "Identifier",
			Api:         serviceAPI,
			Events: &model.Events{
				Spec: []byte("events spec"),
			},
			Documentation: []byte("documentation"),
		}

		remoteEnvServiceAPI := &remoteenv.ServiceAPI{
			TargetUrl:   "http://target.com",
			AccessLabel: "access-label",
			GatewayURL:  "gateway-url",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "http://oauth.com/token",
				SecretName:        "secret-name",
			},
		}

		remoteEnvService := remoteenv.Service{
			ID:                  "uuid-1",
			Identifier:          "Identifier",
			DisplayName:         "Some service",
			LongDescription:     "Some cool service",
			ShortDescription:    "Some cool service",
			ProviderDisplayName: "Service Provider",
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 remoteEnvServiceAPI,
			Events:              true,
		}

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Update", "re", "uuid-1", serviceAPI).Return(remoteEnvServiceAPI, nil)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(serviceAPI, nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)
		serviceRepository.On("Update", "re", remoteEnvService).Return(apperrors.Internal("an error"))

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Put", "uuid-1", []byte("documentation"), []byte("{\"api\":\"spec\"}"), []byte("events spec")).Return(nil)
		minioService.On("Get", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		_, err := service.Update("re", "uuid-1", &serviceDefinition)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})
}

func TestServiceDefinitionService_Delete(t *testing.T) {

	t.Run("should delete a service", func(t *testing.T) {
		// given
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Delete", "re", "uuid-1").Return(nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		minioService := new(miniomocks.Service)
		minioService.On("Remove", "uuid-1").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, minioService)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})

	t.Run("should return an error if API deletion failed", func(t *testing.T) {
		// given
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(apperrors.Internal("an error"))

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, nil, nil)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
	})

	t.Run("should return an error if remoteenv delete failed", func(t *testing.T) {
		// given
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Delete", "re", "uuid-1").Return(apperrors.Internal("an error"))

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, nil)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
	})

	t.Run("should return an error if Minio data deletion failed", func(t *testing.T) {
		// given
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Delete", "re", "uuid-1").Return(nil)

		minioService := new(miniomocks.Service)
		minioService.On("Remove", "uuid-1").Return(apperrors.Internal("an error"))

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, minioService)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		minioService.AssertExpectations(t)
	})
}

func TestServiceDefinitionService_GetAPI(t *testing.T) {

	t.Run("should get API", func(t *testing.T) {
		// given
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{}
		remoteEnvService := remoteenv.Service{API: remoteEnvServiceAPI}
		serviceAPI := &model.API{}

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(serviceAPI, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		require.NoError(t, err)

		assert.Equal(t, serviceAPI, result)
	})

	t.Run("should return not found error if service does not exist", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.NotFound("missing"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})

	t.Run("should return internal error if service does not exist", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.Internal("some error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should return bad request if service does not have API", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, nil)

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should return internal error if reading service API fails", func(t *testing.T) {
		// given
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{}
		remoteEnvService := remoteenv.Service{API: remoteEnvServiceAPI}

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteEnvService, nil)

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Read", "re", remoteEnvServiceAPI).Return(nil, apperrors.Internal("some error"))

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")
	})
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
