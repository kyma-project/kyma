// Package metadata contains components for accessing Kyma storage (Applications, Minio)
package metadata

import (
	"fmt"

	alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/uuid"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	connectedApp = "connected-app"
)

// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// Create adds new ServiceDefinition.
	Create(application string, serviceDefinition *model.ServiceDefinition) (id string, err apperrors.AppError)

	// GetByID returns ServiceDefinition with provided ID.
	GetByID(application, id string) (serviceDefinition model.ServiceDefinition, err apperrors.AppError)

	// GetAll returns all ServiceDefinitions.
	GetAll(application string) (serviceDefinitions []model.ServiceDefinition, err apperrors.AppError)

	// Update updates a service definition with provided ID.
	Update(application string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError)

	// Delete deletes a ServiceDefinition.
	Delete(application, id string) apperrors.AppError

	// GetAPI gets API of a service with given ID
	GetAPI(application, serviceId string) (*model.API, apperrors.AppError)
}

type ApplicationGetter interface {
	Get(name string, options v1.GetOptions) (*alpha1.Application, error)
}

type serviceDefinitionService struct {
	uuidGenerator         uuid.Generator
	serviceAPIService     serviceapi.Service
	applicationRepository applications.ServiceRepository
	specService           specification.Service
	applicationManager    ApplicationGetter
}

// NewServiceDefinitionService creates new ServiceDefinitionService with provided dependencies.
func NewServiceDefinitionService(uuidGenerator uuid.Generator, serviceAPIService serviceapi.Service, applicationRepository applications.ServiceRepository, specService specification.Service, applicationManager ApplicationGetter) ServiceDefinitionService {
	return &serviceDefinitionService{
		uuidGenerator:         uuidGenerator,
		serviceAPIService:     serviceAPIService,
		applicationRepository: applicationRepository,
		specService:           specService,
		applicationManager:    applicationManager,
	}
}

// Create adds new ServiceDefinition. Based on ServiceDefinition a new service is added to application.
func (sds *serviceDefinitionService) Create(application string, serviceDef *model.ServiceDefinition) (string, apperrors.AppError) {
	if serviceDef.Identifier != "" {
		apperr := sds.ensureUniqueIdentifier(serviceDef.Identifier, application)
		if apperr != nil {
			return "", apperr.Append("Creating service failed")
		}
	}

	var err error
	serviceDef.ID, err = sds.uuidGenerator.NewUUID()
	if err != nil {
		return "", apperrors.Internal("Creating uuid failed, %s", err)
	}
	service := initService(serviceDef, serviceDef.Identifier, application)

	var gatewayUrl string

	appUID, apperr := sds.getApplicationUID(application)
	if apperr != nil {
		return "", apperr.Append("Getting Application UID failed")
	}

	if apiDefined(serviceDef) {
		serviceAPI, apperr := sds.serviceAPIService.New(application, appUID, serviceDef.ID, serviceDef.Api)
		if apperr != nil {
			return "", apperr.Append("Adding new API failed")
		}

		service.API = serviceAPI
		gatewayUrl = serviceAPI.GatewayURL
	}

	apperr = sds.specService.PutSpec(serviceDef, gatewayUrl)
	if apperr != nil {
		return "", apperr.Append("Determining API spec for service with ID %s failed", serviceDef.ID)
	}

	apperr = sds.applicationRepository.Create(application, *service)
	if apperr != nil {
		return "", apperr.Append("Creating service in Application failed")
	}

	return serviceDef.ID, nil
}

func (sds *serviceDefinitionService) getApplicationUID(application string) (types.UID, apperrors.AppError) {
	app, err := sds.applicationManager.Get(application, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Application %s not found", application)
			return "", apperrors.NotFound(message)
		}

		message := fmt.Sprintf("Getting Application %s failed, %s", application, err.Error())
		return "", apperrors.Internal(message)
	}

	return app.UID, nil
}

// GetByID returns ServiceDefinition with provided ID.
func (sds *serviceDefinitionService) GetByID(application, id string) (model.ServiceDefinition, apperrors.AppError) {
	service, apperr := sds.applicationRepository.Get(application, id)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Reading service with ID %s failed", id)
	}

	return sds.readService(application, service)
}

// GetAll returns all ServiceDefinitions.
func (sds *serviceDefinitionService) GetAll(application string) ([]model.ServiceDefinition, apperrors.AppError) {
	services, apperr := sds.applicationRepository.GetAll(application)
	if apperr != nil {
		return nil, apperr.Append("Reading services from Application failed")
	}

	res := make([]model.ServiceDefinition, 0)
	for _, service := range services {
		res = append(res, convertServiceBaseInfo(service))
	}

	return res, nil
}

// Update updates a service with provided ID.
func (sds *serviceDefinitionService) Update(application string, serviceDef *model.ServiceDefinition) (model.ServiceDefinition, apperrors.AppError) {
	existingSvc, apperr := sds.GetByID(application, serviceDef.ID)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Updating %s service failed", serviceDef.ID)
	}

	service := initService(serviceDef, existingSvc.Identifier, application)

	var gatewayUrl string

	appUID, apperr := sds.getApplicationUID(application)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Getting Application UID failed")
	}

	if !apiDefined(serviceDef) {
		apperr = sds.serviceAPIService.Delete(application, serviceDef.ID)
		if apperr != nil {
			return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, deleting API failed", serviceDef.ID)
		}
	} else {
		service.API, apperr = sds.serviceAPIService.Update(application, appUID, serviceDef.ID, serviceDef.Api)
		if apperr != nil {
			return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, updating API failed", serviceDef.ID)
		}

		gatewayUrl = service.API.GatewayURL
	}

	apperr = sds.specService.PutSpec(serviceDef, gatewayUrl)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, saving specification failed", serviceDef.ID)
	}

	apperr = sds.applicationRepository.Update(application, *service)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Updating %s service failed, updating service in Application repository failed", serviceDef.ID)
	}

	return convertServiceBaseInfo(*service), nil
}

// Delete deletes a service with given id.
func (sds *serviceDefinitionService) Delete(application, id string) apperrors.AppError {
	apperr := sds.serviceAPIService.Delete(application, id)
	if apperr != nil {
		return apperr.Append("Deleting service failed")
	}

	apperr = sds.applicationRepository.Delete(application, id)
	if apperr != nil {
		return apperr.Append("Deleting service from Application repository failed")
	}

	apperr = sds.specService.RemoveSpec(id)
	if apperr != nil {
		return apperr.Append("Deleting service specification failed")
	}

	return nil
}

// GetAPI gets API of a service with given ID
func (sds *serviceDefinitionService) GetAPI(application, serviceId string) (*model.API, apperrors.AppError) {
	service, apperr := sds.applicationRepository.Get(application, serviceId)
	if apperr != nil {
		return nil, apperr.Append("Reading %s service failed", serviceId)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("Service with ID %s has no API", service.ID)
	}

	api, apperr := sds.serviceAPIService.Read(application, service.API)
	if apperr != nil {
		return nil, apperr.Append("Reading API for %s service failed", serviceId)
	}
	return api, nil
}

func initService(serviceDef *model.ServiceDefinition, identifier, application string) *applications.Service {
	service := applications.Service{
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
		service.Labels = overrideLabels(application, *serviceDef.Labels)
	} else {
		service.Labels = map[string]string{connectedApp: application}
	}

	return &service
}

func convertServiceBaseInfo(service applications.Service) model.ServiceDefinition {
	return model.ServiceDefinition{
		ID:          service.ID,
		Name:        service.DisplayName,
		Identifier:  service.Identifier,
		Labels:      &service.Labels,
		Description: service.LongDescription,
		Provider:    service.ProviderDisplayName,
	}
}

func (sds *serviceDefinitionService) ensureUniqueIdentifier(identifier, application string) apperrors.AppError {
	services, apperr := sds.GetAll(application)
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

func (sds *serviceDefinitionService) readService(application string, service applications.Service) (model.ServiceDefinition, apperrors.AppError) {
	serviceDef := convertServiceBaseInfo(service)

	documentation, apiSpec, eventsSpec, apperr := sds.specService.GetSpec(service.ID)
	if apperr != nil {
		return model.ServiceDefinition{}, apperr.Append("Reading specs failed")
	}

	if service.API != nil {
		api, apperr := sds.serviceAPIService.Read(application, service.API)
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

func overrideLabels(application string, labels map[string]string) map[string]string {
	labels[connectedApp] = application

	return labels
}
