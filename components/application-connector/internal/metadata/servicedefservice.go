// Package metadata contains components for accessing Kyma storage (Remote Environments, Minio)
package metadata

import (
	"encoding/json"
	"github.com/go-openapi/spec"
	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/minio"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/uuid"
	"net/url"
)

const targetSwaggerVersion = "2.0"

// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// Create adds new ServiceDefinition.
	Create(remoteEnvironment string, serviceDefinition *ServiceDefinition) (id string, err apperrors.AppError)

	// GetByID returns ServiceDefinition with provided ID.
	GetByID(remoteEnvironment, id string) (serviceDefinition ServiceDefinition, err apperrors.AppError)

	// GetAll returns all ServiceDefinitions.
	GetAll(remoteEnvironment string) (serviceDefinitions []ServiceDefinition, err apperrors.AppError)

	// Update updates a service definition with provided ID.
	Update(remoteEnvironment, id string, serviceDef *ServiceDefinition) apperrors.AppError

	// Delete deletes a ServiceDefinition.
	Delete(remoteEnvironment, id string) apperrors.AppError

	// GetAPI gets API of a service with given ID
	GetAPI(remoteEnvironment, serviceId string) (*serviceapi.API, apperrors.AppError)
}

type serviceDefinitionService struct {
	uuidGenerator               uuid.Generator
	serviceAPIService           serviceapi.Service
	remoteEnvironmentRepository remoteenv.ServiceRepository
	minioService                minio.Service
}

// NewServiceDefinitionService creates new ServiceDefinitionService with provided dependencies.
func NewServiceDefinitionService(uuidGenerator uuid.Generator, serviceAPIService serviceapi.Service, remoteEnvironmentRepository remoteenv.ServiceRepository, minioService minio.Service) ServiceDefinitionService {
	return &serviceDefinitionService{
		uuidGenerator:               uuidGenerator,
		serviceAPIService:           serviceAPIService,
		remoteEnvironmentRepository: remoteEnvironmentRepository,
		minioService:                minioService,
	}
}

// Create adds new ServiceDefinition. Based on ServiceDefinition a new service is added to RemoteEnvironment.
func (sds *serviceDefinitionService) Create(remoteEnvironment string, serviceDef *ServiceDefinition) (string, apperrors.AppError) {
	id := sds.uuidGenerator.NewUUID()

	service := initService(serviceDef, id)

	if apiDefined(serviceDef) {
		serviceAPI, err := sds.serviceAPIService.New(remoteEnvironment, id, serviceDef.Api)
		if err != nil {
			return "", apperrors.Internal("failed to add new API, %s", err)
		}
		service.API = serviceAPI

		serviceDef.Api.Spec, err = modifyAPISpec(serviceDef.Api.Spec, serviceAPI.GatewayURL)
		if err != nil {
			return "", apperrors.Internal("failed to modify API spec, %s", err)
		}
	}

	err := sds.insertSpecs(id, serviceDef.Documentation, serviceDef.Api, serviceDef.Events)
	if err != nil {
		return "", apperrors.Internal("failed to insert specs, %s", err)
	}

	err = sds.remoteEnvironmentRepository.Create(remoteEnvironment, *service)
	if err != nil {
		return "", apperrors.Internal("failed to create service in remote environment, %s", err)
	}

	serviceDef.ID = id
	return id, nil
}

// GetByID returns ServiceDefinition with provided ID.
func (sds *serviceDefinitionService) GetByID(remoteEnvironment, id string) (ServiceDefinition, apperrors.AppError) {
	service, err := sds.remoteEnvironmentRepository.Get(remoteEnvironment, id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return ServiceDefinition{}, apperrors.NotFound("service with ID %s not found", id)
		}
		return ServiceDefinition{}, apperrors.Internal("failed to read service with ID %s, %s", id, err)
	}

	return sds.readService(remoteEnvironment, service)
}

// GetAll returns all ServiceDefinitions.
func (sds *serviceDefinitionService) GetAll(remoteEnvironment string) ([]ServiceDefinition, apperrors.AppError) {
	services, err := sds.remoteEnvironmentRepository.GetAll(remoteEnvironment)
	if err != nil {
		return nil, apperrors.Internal("failed to read services from remote environment, %s", err)
	}

	res := make([]ServiceDefinition, 0)
	for _, service := range services {
		res = append(res, convertServiceBaseInfo(service))
	}

	return res, nil
}

// Update updates a service with provided ID.
func (sds *serviceDefinitionService) Update(remoteEnvironment, id string, serviceDef *ServiceDefinition) apperrors.AppError {
	_, err := sds.GetByID(remoteEnvironment, id)
	if err != nil {
		if err.Code() != apperrors.CodeNotFound {
			return apperrors.NotFound("failed to get service before update, %s", err)
		}
		return apperrors.Internal("failed to read service, %s", err)
	}

	service := initService(serviceDef, id)

	if !apiDefined(serviceDef) {
		err = sds.serviceAPIService.Delete(remoteEnvironment, id)
		if err != nil {
			return apperrors.Internal("failed to delete API, %s", err)
		}
	} else {
		service.API, err = sds.serviceAPIService.Update(remoteEnvironment, id, serviceDef.Api)
		if err != nil {
			return apperrors.Internal("failed to update API, %s", err)
		}

		serviceDef.Api.Spec, err = modifyAPISpec(serviceDef.Api.Spec, service.API.GatewayURL)
		if err != nil {
			return apperrors.Internal("failed to modify API spec, %s", err)
		}
	}

	err = sds.insertSpecs(id, serviceDef.Documentation, serviceDef.Api, serviceDef.Events)
	if err != nil {
		return apperrors.Internal("failed to insert specification to Minio, %s", err)
	}

	err = sds.remoteEnvironmentRepository.Update(remoteEnvironment, *service)
	if err != nil {
		return apperrors.Internal("failed to update service in RE repository, %s")
	}

	serviceDef.ID = id
	return nil
}

// Delete deletes a service with given id.
func (sds *serviceDefinitionService) Delete(remoteEnvironment, id string) apperrors.AppError {
	err := sds.serviceAPIService.Delete(remoteEnvironment, id)
	if err != nil {
		return apperrors.Internal("failed to delete service, %s", err)
	}

	err = sds.remoteEnvironmentRepository.Delete(remoteEnvironment, id)
	if err != nil {
		return apperrors.Internal("failed to delete service from RE repository, %s", err)
	}

	err = sds.minioService.Remove(id)
	if err != nil {
		return apperrors.Internal("failed to delete service data from Minio, %s", err)
	}

	return nil
}

// GetAPI gets API of a service with given ID
func (sds *serviceDefinitionService) GetAPI(remoteEnvironment, serviceId string) (*serviceapi.API, apperrors.AppError) {
	service, err := sds.remoteEnvironmentRepository.Get(remoteEnvironment, serviceId)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return nil, apperrors.NotFound("service with ID %s not found", serviceId)
		}
		return nil, apperrors.Internal("failed to read %s service, %s", serviceId, err)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("service with ID '%s' has no API")
	}

	api, err := sds.serviceAPIService.Read(remoteEnvironment, service.API)
	if err != nil {
		return nil, apperrors.Internal("failed to read API for %s service, %s", serviceId, err)
	}
	return api, nil
}

func initService(serviceDef *ServiceDefinition, id string) *remoteenv.Service {
	service := remoteenv.Service{
		ID:                  id,
		DisplayName:         serviceDef.Name,
		LongDescription:     serviceDef.Description,
		ProviderDisplayName: serviceDef.Provider,
		Tags:                make([]string, 0),
	}

	service.Events = serviceDef.Events != nil

	return &service
}

func convertServiceBaseInfo(service remoteenv.Service) ServiceDefinition {
	return ServiceDefinition{
		ID:          service.ID,
		Name:        service.DisplayName,
		Description: service.LongDescription,
		Provider:    service.ProviderDisplayName,
	}
}

func (sds *serviceDefinitionService) readService(remoteEnvironment string, service remoteenv.Service) (ServiceDefinition, apperrors.AppError) {
	serviceDef := convertServiceBaseInfo(service)

	documentation, apiSpec, eventsSpec, err := sds.minioService.Get(service.ID)
	if err != nil {
		return ServiceDefinition{}, apperrors.Internal("reading specs failed, %s", err)
	}

	if service.API != nil {
		api, err := sds.serviceAPIService.Read(remoteEnvironment, service.API)
		if err != nil {
			return ServiceDefinition{}, apperrors.Internal("reading API failed, %s", err)
		}
		serviceDef.Api = api

		if apiSpec != nil {
			serviceDef.Api.Spec = apiSpec
		}
	}

	if eventsSpec != nil {
		serviceDef.Events = &Events{eventsSpec}
	}

	if documentation != nil {
		serviceDef.Documentation = documentation
	}

	return serviceDef, nil
}

func apiDefined(serviceDefinition *ServiceDefinition) bool {
	return serviceDefinition.Api != nil
}

func (sds *serviceDefinitionService) insertSpecs(id string, docs []byte, api *serviceapi.API, events *Events) apperrors.AppError {
	var documentation []byte
	var apiSpec []byte
	var eventsSpec []byte

	if docs != nil {
		documentation = docs
	}

	if api != nil {
		apiSpec = api.Spec
	}

	if events != nil {
		eventsSpec = events.Spec
	}

	return sds.minioService.Put(id, documentation, apiSpec, eventsSpec)
}

func modifyAPISpec(rawApiSpec []byte, gatewayUrl string) ([]byte, apperrors.AppError) {
	if rawApiSpec == nil {
		return rawApiSpec, nil
	}

	var apiSpec spec.Swagger
	err := json.Unmarshal(rawApiSpec, &apiSpec)
	if err != nil {
		return []byte{}, apperrors.Internal("failed to unmarshal api spec, %s", err)
	}

	if apiSpec.Swagger != targetSwaggerVersion {
		return rawApiSpec, nil
	}

	newSpec, err := updateBaseUrl(apiSpec, gatewayUrl)
	if err != nil {
		return rawApiSpec, apperrors.Internal("failed to update base url, %s", err)
	}

	modifiedSpec, err := json.Marshal(newSpec)
	if err != nil {
		return rawApiSpec, apperrors.Internal("failed to marshal updated spec, %s", err)
	}

	return modifiedSpec, nil
}

func updateBaseUrl(apiSpec spec.Swagger, gatewayUrl string) (spec.Swagger, apperrors.AppError) {
	fullUrl, err := url.Parse(gatewayUrl)
	if err != nil {
		return spec.Swagger{}, apperrors.Internal("failed to parse gateway url, %s", err)
	}

	apiSpec.Host = fullUrl.Hostname()
	apiSpec.BasePath = ""
	apiSpec.Schemes = []string{"http"}

	return apiSpec, nil
}
