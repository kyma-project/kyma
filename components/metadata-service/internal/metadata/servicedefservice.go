// Package metadata contains components for accessing Kyma storage (Remote Environments, Minio)
package metadata

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/uuid"
)

const (
	connectedApp = "connected-app"
)

// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// Create adds new ServiceDefinition.
	Create(remoteEnvironment string, serviceDefinition *model.ServiceDefinition) (id string, err apperrors.AppError)

	// GetByID returns ServiceDefinition with provided ID.
	GetByID(remoteEnvironment, id string) (serviceDefinition model.ServiceDefinition, err apperrors.AppError)

	// GetAll returns all ServiceDefinitions.
	GetAll(remoteEnvironment string) (serviceDefinitions []model.ServiceDefinition, err apperrors.AppError)

	// Update updates a service definition with provided ID.
	Update(remoteEnvironment, id string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError)

	// Delete deletes a ServiceDefinition.
	Delete(remoteEnvironment, id string) apperrors.AppError

	// GetAPI gets API of a service with given ID
	GetAPI(remoteEnvironment, serviceId string) (*model.API, apperrors.AppError)
}

type serviceDefinitionService struct {
	uuidGenerator               uuid.Generator
	serviceAPIService           serviceapi.Service
	remoteEnvironmentRepository remoteenv.ServiceRepository
	specService                 specification.Service
}

// NewServiceDefinitionService creates new ServiceDefinitionService with provided dependencies.
func NewServiceDefinitionService(uuidGenerator uuid.Generator, serviceAPIService serviceapi.Service, remoteEnvironmentRepository remoteenv.ServiceRepository, specService specification.Service) ServiceDefinitionService {
	return &serviceDefinitionService{
		uuidGenerator:               uuidGenerator,
		serviceAPIService:           serviceAPIService,
		remoteEnvironmentRepository: remoteEnvironmentRepository,
		specService:                 specService,
	}
}

// Create adds new ServiceDefinition. Based on ServiceDefinition a new service is added to RemoteEnvironment.
func (sds *serviceDefinitionService) Create(remoteEnvironment string, serviceDef *model.ServiceDefinition) (string, apperrors.AppError) {
	if serviceDef.Identifier != "" {
		err := sds.ensureUniqueIdentifier(serviceDef.Identifier, remoteEnvironment)
		if err != nil {
			return "", err
		}
	}

	serviceDef.ID = sds.uuidGenerator.NewUUID()
	service := initService(serviceDef, serviceDef.Identifier, remoteEnvironment)

	var gatewayUrl string

	if apiDefined(serviceDef) {
		serviceAPI, err := sds.serviceAPIService.New(remoteEnvironment, serviceDef.ID, serviceDef.Api)
		if err != nil {
			return "", apperrors.Internal("Adding new API failed, %s", err.Error())
		}

		service.API = serviceAPI
		gatewayUrl = serviceAPI.GatewayURL
	}

	err := sds.specService.PutSpec(serviceDef, gatewayUrl)
	if err != nil {
		if err.Code() == apperrors.CodeUpstreamServerCallFailed {
			return "", apperrors.UpstreamServerCallFailed("Determining API spec for service with ID %s failed, %s", serviceDef.ID, err.Error())
		}
		return "", apperrors.Internal("Determining API spec for service with ID %s failed, %s", serviceDef.ID, err.Error())
	}

	err = sds.remoteEnvironmentRepository.Create(remoteEnvironment, *service)
	if err != nil {
		return "", apperrors.Internal("Creating service in Remote Environment failed, %s", err.Error())
	}

	return serviceDef.ID, nil
}

// GetByID returns ServiceDefinition with provided ID.
func (sds *serviceDefinitionService) GetByID(remoteEnvironment, id string) (model.ServiceDefinition, apperrors.AppError) {
	service, err := sds.remoteEnvironmentRepository.Get(remoteEnvironment, id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return model.ServiceDefinition{}, apperrors.NotFound("Service with ID %s not found", id)
		}
		return model.ServiceDefinition{}, apperrors.Internal("Reading service with ID %s failed, %s", id, err.Error())
	}

	return sds.readService(remoteEnvironment, service)
}

// GetAll returns all ServiceDefinitions.
func (sds *serviceDefinitionService) GetAll(remoteEnvironment string) ([]model.ServiceDefinition, apperrors.AppError) {
	services, err := sds.remoteEnvironmentRepository.GetAll(remoteEnvironment)
	if err != nil {
		return nil, apperrors.Internal("Reading services from Remote Environment failed, %s", err.Error())
	}

	res := make([]model.ServiceDefinition, 0)
	for _, service := range services {
		res = append(res, convertServiceBaseInfo(service))
	}

	return res, nil
}

// Update updates a service with provided ID.
func (sds *serviceDefinitionService) Update(remoteEnvironment, id string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError) {
	existingSvc, err := sds.GetByID(remoteEnvironment, id)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return model.ServiceDefinition{}, apperrors.NotFound("Updating %s service failed, %s", id, err.Error())
		}
		return model.ServiceDefinition{}, apperrors.Internal("Updating %s service failed, %s", id, err.Error())
	}

	serviceDef.ID = id
	service := initService(serviceDef, existingSvc.Identifier, remoteEnvironment)

	var gatewayUrl string

	if !apiDefined(serviceDef) {
		err = sds.serviceAPIService.Delete(remoteEnvironment, id)
		if err != nil {
			return model.ServiceDefinition{}, apperrors.Internal("Updating %s service failed, deleting API failed, %s", id, err.Error())
		}
	} else {
		service.API, err = sds.serviceAPIService.Update(remoteEnvironment, id, serviceDef.Api)
		if err != nil {
			return model.ServiceDefinition{}, apperrors.Internal("Updating %s service failed, updating API failed, %s", id, err.Error())
		}

		gatewayUrl = service.API.GatewayURL
	}

	err = sds.specService.PutSpec(serviceDef, gatewayUrl)
	if err != nil {
		if err.Code() == apperrors.CodeUpstreamServerCallFailed {
			return model.ServiceDefinition{}, apperrors.UpstreamServerCallFailed("Updating %s service failed, saving specification failed, %s", id, err.Error())
		}
		return model.ServiceDefinition{}, apperrors.Internal("Updating %s service failed, saving specification failed, %s", id, err.Error())
	}

	err = sds.remoteEnvironmentRepository.Update(remoteEnvironment, *service)
	if err != nil {
		return model.ServiceDefinition{}, apperrors.Internal("Updating %s service failed, updating service in Remote Environment repository failed, %s", id, err.Error())
	}

	return convertServiceBaseInfo(*service), nil
}

// Delete deletes a service with given id.
func (sds *serviceDefinitionService) Delete(remoteEnvironment, id string) apperrors.AppError {
	err := sds.serviceAPIService.Delete(remoteEnvironment, id)
	if err != nil {
		return apperrors.Internal("Deleting service failed, %s", err.Error())
	}

	err = sds.remoteEnvironmentRepository.Delete(remoteEnvironment, id)
	if err != nil {
		return apperrors.Internal("Deleting service from Remote Environment repository failed, %s", err.Error())
	}

	err = sds.specService.RemoveSpec(id)
	if err != nil {
		return apperrors.Internal("Deleting service specification failed, %s", err.Error())
	}

	return nil
}

// GetAPI gets API of a service with given ID
func (sds *serviceDefinitionService) GetAPI(remoteEnvironment, serviceId string) (*model.API, apperrors.AppError) {
	service, err := sds.remoteEnvironmentRepository.Get(remoteEnvironment, serviceId)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return nil, apperrors.NotFound("Service with ID %s not found", serviceId)
		}
		return nil, apperrors.Internal("Reading %s service failed, %s", serviceId, err.Error())
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("Service with ID %s has no API", service.ID)
	}

	api, err := sds.serviceAPIService.Read(remoteEnvironment, service.API)
	if err != nil {
		return nil, apperrors.Internal("Reading API for %s service failed, %s", serviceId, err.Error())
	}
	return api, nil
}

func initService(serviceDef *model.ServiceDefinition, identifier, remoteEnvironment string) *remoteenv.Service {
	service := remoteenv.Service{
		ID:                  serviceDef.ID,
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

func convertServiceBaseInfo(service remoteenv.Service) model.ServiceDefinition {
	return model.ServiceDefinition{
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
			return apperrors.AlreadyExists("Service with Identifier %s already exists", identifier)
		}
	}

	return nil
}

func (sds *serviceDefinitionService) readService(remoteEnvironment string, service remoteenv.Service) (model.ServiceDefinition, apperrors.AppError) {
	serviceDef := convertServiceBaseInfo(service)

	documentation, apiSpec, eventsSpec, err := sds.specService.GetSpec(service.ID)
	if err != nil {
		return model.ServiceDefinition{}, apperrors.Internal("Reading specs failed, %s", err.Error())
	}

	if service.API != nil {
		api, err := sds.serviceAPIService.Read(remoteEnvironment, service.API)
		if err != nil {
			return model.ServiceDefinition{}, apperrors.Internal("Reading API failed, %s", err.Error())
		}
		serviceDef.Api = api

		if apiSpec != nil {
			serviceDef.Api.Spec = apiSpec
		}
	}

	if eventsSpec != nil {
		serviceDef.Events = &model.Events{Spec: eventsSpec}
	}

	if documentation != nil {
		serviceDef.Documentation = documentation
	}

	return serviceDef, nil
}

func apiDefined(serviceDefinition *model.ServiceDefinition) bool {
	return serviceDefinition.Api != nil
}

func overrideLabels(remoteEnvironment string, labels map[string]string) map[string]string {
	_, found := labels[connectedApp]
	if found {
		labels[connectedApp] = remoteEnvironment
	}

	return labels
}
