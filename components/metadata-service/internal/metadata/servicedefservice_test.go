package metadata

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"testing"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	remoteenvmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv/mocks"
	serviceapimocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi/mocks"
	specmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/mocks"
	uuidmocks "github.com/kyma-project/kyma/components/metadata-service/internal/metadata/uuid/mocks"
	"github.com/stretchr/testify/assert"
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "gateway-url").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
	})

	t.Run("should create service without API", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			Name:        "Some service",
			Description: "Some cool service",
			Provider:    "Service Provider",
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
			Labels:              map[string]string{"connected-app": "re"},
			Tags:                make([]string, 0),
			API:                 nil,
			Events:              true,
		}

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Create", "re", remoteEnvService).Return(nil)
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, nil, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.NoError(t, err)
		assert.Equal(t, "uuid-1", serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		require.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
		assert.Empty(t, serviceID)

		serviceAPIService.AssertExpectations(t)
	})

	t.Run("should return error when saving spec fails", func(t *testing.T) {
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(apperrors.Internal("error"))

		service := NewServiceDefinitionService(uuidGenerator, nil, nil, specService)

		// when
		_, err := service.Create("re", &serviceDefinition)

		// then
		require.Error(t, err)

		uuidGenerator.AssertExpectations(t)
		specService.AssertExpectations(t)
	})

	t.Run("should return internal error when creating service in remote environment fails", func(t *testing.T) {
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
			GatewayURL:  "gateway-url",
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
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "gateway-url").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
		assert.Empty(t, serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
	})

	t.Run("should return not found error when creating service in remote environment that not exists", func(t *testing.T) {
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
			GatewayURL:  "gateway-url",
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
		serviceRepository.On("Create", "re", remoteEnvService).Return(apperrors.NotFound("some error"))
		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "gateway-url").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, specService)

		// when
		serviceID, err := service.Create("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
		assert.Equal(t, err.Code(), apperrors.CodeNotFound)
		assert.Empty(t, serviceID)

		uuidGenerator.AssertExpectations(t)
		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeAlreadyExists, err.Code())
		assert.Empty(t, serviceID)

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

		// then
		require.NoError(t, err)
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

		// then
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("should return not found error if cannot find Remote Environment", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("GetAll", "re").Return(nil, apperrors.NotFound("Remote Environment re not found"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		_, err := service.GetAll("re")

		//then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
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
		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(empty, empty, empty, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		result, err := service.GetByID("re", "uuid-1")

		// then
		require.NoError(t, err)
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
		specService.AssertExpectations(t)
	})

	t.Run("should return internal error when getting service from remote environment fails", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.Internal("get error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get error")
		assert.Equal(t, err.Code(), apperrors.CodeInternal)
	})

	t.Run("should return not found error when getting service from remote environment that not exists", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.NotFound("get error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get error")
		assert.Equal(t, err.Code(), apperrors.CodeNotFound)
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
		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(empty, empty, empty, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "api error")
	})

	t.Run("should return error when reading specs fails", func(t *testing.T) {
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
		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(empty, empty, empty, apperrors.Internal("error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, specService)

		// when
		_, err := service.GetByID("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error")

		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "gateway-url").Return(nil)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
			ID:            "uuid-1",
			Name:          "Some service",
			Description:   "Some cool service",
			Provider:      "Service Provider",
			Api:           serviceAPI,
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

		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})

	t.Run("should update a service when no API was given", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
	})

	t.Run("should preserve a service identifier", func(t *testing.T) {
		// given
		serviceDefinition := model.ServiceDefinition{
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("PutSpec", &serviceDefinition, "").Return(nil)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.Error(t, err)
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
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
	})

	t.Run("should return an error if spec update failed", func(t *testing.T) {
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
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)
		specService.On("PutSpec", &serviceDefinition, "").Return(apperrors.Internal("Error"))

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		specService.AssertExpectations(t)
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
			ID:          "uuid-1",
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

		specService := new(specmocks.Service)
		specService.On("GetSpec", "uuid-1").Return(nil, nil, nil, nil)
		specService.On("PutSpec", &serviceDefinition, "gateway-url").Return(nil)

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		_, err := service.Update("re", &serviceDefinition)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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

		specService := new(specmocks.Service)
		specService.On("RemoveSpec", "uuid-1").Return(nil)

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, specService)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		require.NoError(t, err)

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
	})

	t.Run("should return an error when trying to delete service, but RE is not found", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Delete", "re", "uuid-1").Return(apperrors.NotFound("A not found error"))

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		uuidGenerator := new(uuidmocks.Generator)
		uuidGenerator.On("NewUUID").Return("uuid-1")

		service := NewServiceDefinitionService(uuidGenerator, serviceAPIService, serviceRepository, nil)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.NotEmpty(t, err.Error())
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

	t.Run("should return an error if spec deletion failed", func(t *testing.T) {
		// given
		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Delete", "re", "uuid-1").Return(nil)

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Delete", "re", "uuid-1").Return(nil)

		specService := new(specmocks.Service)
		specService.On("RemoveSpec", "uuid-1").Return(apperrors.Internal("an error"))

		service := NewServiceDefinitionService(nil, serviceAPIService, serviceRepository, specService)

		// when
		err := service.Delete("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		serviceAPIService.AssertExpectations(t)
		serviceRepository.AssertExpectations(t)
		specService.AssertExpectations(t)
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
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.Nil(t, result)
	})

	t.Run("should return internal error if service does not exist", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, apperrors.Internal("some error"))

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")
		assert.Nil(t, result)
	})

	t.Run("should return bad request if service does not have API", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "re", "uuid-1").Return(remoteenv.Service{}, nil)

		service := NewServiceDefinitionService(nil, nil, serviceRepository, nil)

		// when
		result, err := service.GetAPI("re", "uuid-1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Nil(t, result)
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
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")
		assert.Nil(t, result)
	})
}
