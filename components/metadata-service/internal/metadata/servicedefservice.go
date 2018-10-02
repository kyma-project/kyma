// Package metadata contains components for accessing Kyma storage (Remote Environments, Minio)
package metadata

import (
	"encoding/json"
	"github.com/go-openapi/spec"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/minio"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/uuid"
	"net/url"
)

const (
	targetSwaggerVersion = "2.0"
	connectedApp         = "connected-app"
)

// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// Create adds new ServiceDefinition.
	Create(remoteEnvironment string, serviceDefinition *ServiceDefinition) (id string, err apperrors.AppError)

	// GetByID returns ServiceDefinition with provided ID.
	GetByID(remoteEnvironment, id string) (serviceDefinition ServiceDefinition, err apperrors.AppError)

	// GetAll returns all ServiceDefinitions.
	GetAll(remoteEnvironment string) (serviceDefinitions []ServiceDefinition, err apperrors.AppError)

	// Update updates a service definition with provided ID.
	Update(remoteEnvironment, id string, serviceDef *ServiceDefinition) (ServiceDefinition, apperrors.AppError)

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
	if serviceDef.Identifier != "" {
		err := sds.ensureUniqueIdentifier(serviceDef.Identifier, remoteEnvironment)
		if err != nil {
			return "", err
		}
	}

	id := sds.uuidGenerator.NewUUID()

	service := initService(serviceDef, id, serviceDef.Identifier, remoteEnvironment)

	if apiDefined(serviceDef) {
		serviceAPI, err := sds.serviceAPIService.New(remoteEnvironment, id, serviceDef.Api)
		if err != nil {
			return "", apperrors.Internal("service definition service: adding new API failed, %s", err)
		}
		service.API = serviceAPI

		serviceDef.Api.Spec, err = modifyAPISpec(serviceDef.Api.Spec, serviceAPI.GatewayURL)
		if err != nil {
			return "", apperrors.Internal("service definition service: modifying API spec failed, %s", err)
		}
	}

	err := sds.insertSpecs(id, serviceDef.Documentation, serviceDef.Api, serviceDef.Events)
	if err != nil {
		return "", apperrors.Internal("service definition service: inserting specs failed, %s", err)
	}

	err = sds.remoteEnvironmentRepository.Create(remoteEnvironment, *service)
	if err != nil {
		return "", apperrors.Internal("service definition service: creating service in Remote Environment failed, %s", err)
	}

	serviceDef.ID = id
	return id, nil
}

// GetByID returns ServiceDefinition with provided ID.
func (sds *serviceDefinitionService) GetByID(remoteEnvironment, id string) (ServiceDefinition, apperrors.AppError) {
	service, err := sds.remoteEnvironmentRepository.Get(remoteEnvironment, id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return ServiceDefinition{}, apperrors.NotFound("service definition service: service with ID %s not found", id)
		}
		return ServiceDefinition{}, apperrors.Internal("service definition service: reading service with ID %s failed, %s", id, err)
	}

	return sds.readService(remoteEnvironment, service)
}

// GetAll returns all ServiceDefinitions.
func (sds *serviceDefinitionService) GetAll(remoteEnvironment string) ([]ServiceDefinition, apperrors.AppError) {
	services, err := sds.remoteEnvironmentRepository.GetAll(remoteEnvironment)
	if err != nil {
		return nil, apperrors.Internal("service definition service: reading all services from Remote Environment failed, %s", err)
	}

	res := make([]ServiceDefinition, 0)
	for _, service := range services {
		res = append(res, convertServiceBaseInfo(service))
	}

	return res, nil
}

// Update updates a service with provided ID.
func (sds *serviceDefinitionService) Update(remoteEnvironment, id string, serviceDef *ServiceDefinition) (ServiceDefinition, apperrors.AppError) {
	existingSvc, err := sds.GetByID(remoteEnvironment, id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return ServiceDefinition{}, apperrors.NotFound("service definition service: updating %s service failed, %s", id, err.Error())
		}
		return ServiceDefinition{}, apperrors.Internal("service definition service: updating %s service failed, %s", id, err.Error())
	}

	service := initService(serviceDef, id, existingSvc.Identifier, remoteEnvironment)

	if !apiDefined(serviceDef) {
		err = sds.serviceAPIService.Delete(remoteEnvironment, id)
		if err != nil {
			return ServiceDefinition{}, apperrors.Internal("service definition service: updating %s service failed, deleting API failed, %s", id, err.Error())
		}
	} else {
		service.API, err = sds.serviceAPIService.Update(remoteEnvironment, id, serviceDef.Api)
		if err != nil {
			return ServiceDefinition{}, apperrors.Internal("service definition service: updating %s service failed, updating API failed, %s", id, err.Error())
		}

		serviceDef.Api.Spec, err = modifyAPISpec(serviceDef.Api.Spec, service.API.GatewayURL)
		if err != nil {
			return ServiceDefinition{}, apperrors.Internal("service definition service: updating %s service failed, modifying API spec failed, %s", id, err.Error())
		}
	}

	err = sds.insertSpecs(id, serviceDef.Documentation, serviceDef.Api, serviceDef.Events)
	if err != nil {
		return ServiceDefinition{}, apperrors.Internal("service definition service: updating %s service failed, inserting specification to Minio failed, %s", id, err.Error())
	}

	err = sds.remoteEnvironmentRepository.Update(remoteEnvironment, *service)
	if err != nil {
		return ServiceDefinition{}, apperrors.Internal("service definition service: updating %s service failed, updating service in Remote Environment repository failed, %s", id, err.Error())
	}

	serviceDef.ID = id
	return convertServiceBaseInfo(*service), nil
}

// Delete deletes a service with given id.
func (sds *serviceDefinitionService) Delete(remoteEnvironment, id string) apperrors.AppError {
	err := sds.serviceAPIService.Delete(remoteEnvironment, id)
	if err != nil {
		return apperrors.Internal("service definition service: deleting service failed, %s", err)
	}

	err = sds.remoteEnvironmentRepository.Delete(remoteEnvironment, id)
	if err != nil {
		return apperrors.Internal("service definition service: deleting service from Remote Environment repository failed, %s", err)
	}

	err = sds.minioService.Remove(id)
	if err != nil {
		return apperrors.Internal("service definition service: deleting service data from Minio failed, %s", err)
	}

	return nil
}

// GetAPI gets API of a service with given ID
func (sds *serviceDefinitionService) GetAPI(remoteEnvironment, serviceId string) (*serviceapi.API, apperrors.AppError) {
	service, err := sds.remoteEnvironmentRepository.Get(remoteEnvironment, serviceId)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return nil, apperrors.NotFound("service definition service: service with ID %s not found", serviceId)
		}
		return nil, apperrors.Internal("service definition service: reading %s service failed, %s", serviceId, err)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("service definition service: service with ID %s has no API", service.ID)
	}

	api, err := sds.serviceAPIService.Read(remoteEnvironment, service.API)
	if err != nil {
		return nil, apperrors.Internal("service definition service: reading API for %s service failed, %s", serviceId, err)
	}
	return api, nil
}

func initService(serviceDef *ServiceDefinition, id, identifier, remoteEnvironment string) *remoteenv.Service {
	service := remoteenv.Service{
		ID:                  id,
		Identifier:          identifier,
		DisplayName:         serviceDef.Name,
		LongDescription:     serviceDef.Description,
		ProviderDisplayName: serviceDef.Provider,
		Tags:                make([]string, 0),
	}

	service.Events = serviceDef.Events != nil

	if serviceDef.ShortDescription == "" {
		service.ShortDescription = serviceDef.Description
	} else {
		service.ShortDescription = serviceDef.ShortDescription
	}

	if serviceDef.Labels != nil {
		service.Labels = overrideLabels(remoteEnvironment, *serviceDef.Labels)
	} else {
		service.Labels = map[string]string{connectedApp: remoteEnvironment}
	}

	return &service
}

func convertServiceBaseInfo(service remoteenv.Service) ServiceDefinition {
	return ServiceDefinition{
		ID:          service.ID,
		Name:        service.DisplayName,
		Identifier:  service.Identifier,
		Labels:      &service.Labels,
		Description: service.LongDescription,
		Provider:    service.ProviderDisplayName,
	}
}

func (sds *serviceDefinitionService) ensureUniqueIdentifier(identifier, remoteEnvironment string) apperrors.AppError {
	services, err := sds.GetAll(remoteEnvironment)
	if err != nil {
		return err
	}

	for _, service := range services {
		if service.Identifier == identifier {
			return apperrors.AlreadyExists("service definition service: service with Identifier %s already exists", identifier)
		}
	}

	return nil
}

func (sds *serviceDefinitionService) readService(remoteEnvironment string, service remoteenv.Service) (ServiceDefinition, apperrors.AppError) {
	serviceDef := convertServiceBaseInfo(service)

	documentation, apiSpec, eventsSpec, err := sds.minioService.Get(service.ID)
	if err != nil {
		return ServiceDefinition{}, apperrors.Internal("service definition service: reading specs failed, %s", err)
	}

	if service.API != nil {
		api, err := sds.serviceAPIService.Read(remoteEnvironment, service.API)
		if err != nil {
			return ServiceDefinition{}, apperrors.Internal("service definition service: reading API failed, %s", err)
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
		return []byte{}, apperrors.Internal("API spec: unmarshalling API spec failed, %s", err)
	}

	if apiSpec.Swagger != targetSwaggerVersion {
		return rawApiSpec, nil
	}

	newSpec, err := updateBaseUrl(apiSpec, gatewayUrl)
	if err != nil {
		return rawApiSpec, apperrors.Internal("API spec: updating base url failed, %s", err)
	}

	modifiedSpec, err := json.Marshal(newSpec)
	if err != nil {
		return rawApiSpec, apperrors.Internal("API spec: marshalling updated API spec failed, %s", err)
	}

	return modifiedSpec, nil
}

func updateBaseUrl(apiSpec spec.Swagger, gatewayUrl string) (spec.Swagger, apperrors.AppError) {
	fullUrl, err := url.Parse(gatewayUrl)
	if err != nil {
		return spec.Swagger{}, apperrors.Internal("gateway URL failed to parse gateway URL, %s", err)
	}

	apiSpec.Host = fullUrl.Hostname()
	apiSpec.BasePath = ""
	apiSpec.Schemes = []string{"http"}

	return apiSpec, nil
}

func overrideLabels(remoteEnvironment string, labels map[string]string) map[string]string {
	_, found := labels[connectedApp]
	if found {
		labels[connectedApp] = remoteEnvironment
	}

	return labels
}
