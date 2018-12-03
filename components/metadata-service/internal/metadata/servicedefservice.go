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
	Update(remoteEnvironment string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError)

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
		apperr := sds.ensureUniqueIdentifier(serviceDef.Identifier, remoteEnvironment)
		if apperr != nil {
			return "", apperr.Append("Creating service failed")
		}
	}

	serviceDef.ID = sds.uuidGenerator.NewUUID()
	service := initService(serviceDef, serviceDef.Identifier, remoteEnvironment)

	var gatewayUrl string

	if apiDefined(serviceDef) {
		serviceAPI, apperr := sds.serviceAPIService.New(remoteEnvironment, serviceDef.ID, serviceDef.Api)
		if apperr != nil {
			return "", apperr.Append("Adding new API failed")
		}

		service.API = serviceAPI
		gatewayUrl = serviceAPI.GatewayURL
	}

	apperr := sds.specService.PutSpec(serviceDef, gatewayUrl)
	if apperr != nil {
		return "", apperr.Append("Determining API spec for service with ID %s failed", serviceDef.ID)
	}

	apperr = sds.remoteEnvironmentRepository.Create(remoteEnvironment, *service)
	if apperr != nil {
		return "", apperr.Append("Creating service in Remote Environment failed")
	}

	return serviceDef.ID, nil
}

// GetByID returns ServiceDefinition with provided ID.
func (sds *serviceDefinitionService) GetByID(remoteEnvironment, id string) (model.ServiceDefinition, apperrors.AppError) {
	service, apperr := sds.remoteEnvironmentRepository.Get(remoteEnvironment, id)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Reading service with ID %s failed", id)
	}

	return sds.readService(remoteEnvironment, service)
}

// GetAll returns all ServiceDefinitions.
func (sds *serviceDefinitionService) GetAll(remoteEnvironment string) ([]model.ServiceDefinition, apperrors.AppError) {
	services, apperr := sds.remoteEnvironmentRepository.GetAll(remoteEnvironment)
	if apperr != nil {
		return nil, apperr.Append("Reading services from Remote Environment failed")
	}

	res := make([]model.ServiceDefinition, 0)
	for _, service := range services {
		res = append(res, convertServiceBaseInfo(service))
	}

	return res, nil
}

// Update updates a service with provided ID.
func (sds *serviceDefinitionService) Update(remoteEnvironment string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError) {
	existingSvc, apperr := sds.GetByID(remoteEnvironment, serviceDef.ID)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Updating %s service failed", serviceDef.ID)
	}

	service := initService(serviceDef, existingSvc.Identifier, remoteEnvironment)

	var gatewayUrl string

	if !apiDefined(serviceDef) {
		apperr = sds.serviceAPIService.Delete(remoteEnvironment, serviceDef.ID)
		if apperr != nil {
			return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, deleting API failed", serviceDef.ID)
		}
	} else {
		service.API, apperr = sds.serviceAPIService.Update(remoteEnvironment, serviceDef.ID, serviceDef.Api)
		if apperr != nil {
			return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, updating API failed", serviceDef.ID)
		}

		gatewayUrl = service.API.GatewayURL
	}

	apperr = sds.specService.PutSpec(serviceDef, gatewayUrl)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, saving specification failed", serviceDef.ID)
	}

	apperr = sds.remoteEnvironmentRepository.Update(remoteEnvironment, *service)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, updating service in Remote Environment repository failed", serviceDef.ID)
	}

	return convertServiceBaseInfo(*service), nil
}

// Delete deletes a service with given id.
func (sds *serviceDefinitionService) Delete(remoteEnvironment, id string) apperrors.AppError {
	apperr := sds.serviceAPIService.Delete(remoteEnvironment, id)
	if apperr != nil {
		return apperr.Append("Deleting service failed")
	}

	apperr = sds.remoteEnvironmentRepository.Delete(remoteEnvironment, id)
	if apperr != nil {
		return apperr.Append("Deleting service from Remote Environment repository failed")
	}

	apperr = sds.specService.RemoveSpec(id)
	if apperr != nil {
		return apperr.Append("Deleting service specification failed")
	}

	return nil
}

// GetAPI gets API of a service with given ID
func (sds *serviceDefinitionService) GetAPI(remoteEnvironment, serviceId string) (*model.API, apperrors.AppError) {
	service, apperr := sds.remoteEnvironmentRepository.Get(remoteEnvironment, serviceId)
	if apperr != nil {
		return nil, apperr.Append("Reading %s service failed", serviceId)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("Service with ID %s has no API", service.ID)
	}

	api, apperr := sds.serviceAPIService.Read(remoteEnvironment, service.API)
	if apperr != nil {
		return nil, apperr.Append("Reading API for %s service failed", serviceId)
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
	services, apperr := sds.GetAll(remoteEnvironment)
	if apperr != nil {
		return apperr.Append("Checking identifier failed")
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

	documentation, apiSpec, eventsSpec, apperr := sds.specService.GetSpec(service.ID)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Reading specs failed")
	}

	if service.API != nil {
		api, apperr := sds.serviceAPIService.Read(remoteEnvironment, service.API)
		if apperr != nil {
			return model.ServiceDefinition{}, apperr.Append("Reading API failed")
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
