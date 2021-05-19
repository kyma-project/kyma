// Package metadata contains components for accessing Kyma Application
package metadata

import (
	"fmt"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=ServiceDefinitionService
// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// GetAPI gets API of a service with given ID
	GetAPIByServiceName(appName, serviceName string) (*model.API, apperrors.AppError)
	GetAPIByEntryName(appName, serviceName, entryName string) (*model.API, apperrors.AppError)
}

type serviceDefinitionService struct {
	serviceAPIService     serviceapi.Service
	applicationRepository applications.ServiceRepository
}

// NewServiceDefinitionService creates new ServiceDefinitionService with provided dependencies.
func NewServiceDefinitionService(serviceAPIService serviceapi.Service, applicationRepository applications.ServiceRepository) ServiceDefinitionService {
	return &serviceDefinitionService{
		serviceAPIService:     serviceAPIService,
		applicationRepository: applicationRepository,
	}
}

// GetAPI gets API of a service with given name
func (sds *serviceDefinitionService) GetAPIByServiceName(appName, serviceName string) (*model.API, apperrors.AppError) {
	service, err := sds.applicationRepository.GetByServiceName(appName, serviceName)

	if err != nil {
		notFoundMessage := fmt.Sprintf("service with name %s not found", serviceName)
		internalErrMessage := fmt.Sprintf("failed to get service with name '%s': %s", serviceName, err.Error())

		return nil, handleError(err, notFoundMessage, internalErrMessage)
	}

	return sds.getAPI(service)
}

func (sds *serviceDefinitionService) GetAPIByEntryName(appName, serviceName, entryName string) (*model.API, apperrors.AppError) {
	service, err := sds.applicationRepository.GetByEntryName(appName, serviceName, entryName)

	if err != nil {
		notFoundMessage := fmt.Sprintf("service with name %s and entry name %s not found", serviceName, entryName)
		internalErrMessage := fmt.Sprintf("failed to get service with name '%s' and entry name '%s': %s", serviceName, entryName, err.Error())

		return nil, handleError(err, notFoundMessage, internalErrMessage)
	}

	return sds.getAPI(service)
}

func (sds *serviceDefinitionService) getAPI(service applications.Service) (*model.API, apperrors.AppError) {

	if service.API == nil {
		return nil, apperrors.WrongInput("service '%s' has no API", service.Name)
	}

	api, err := sds.serviceAPIService.Read(service.API)
	if err != nil {
		log.Errorf("failed to read api for serviceId '%s': %s", service.Name, err.Error())
		return nil, apperrors.Internal("failed to read API for %s service, %s", service.Name, err)
	}
	return api, nil
}

func handleError(err apperrors.AppError, notFoundMessage, internalErrorMEssage string) apperrors.AppError {
	if err.Code() == apperrors.CodeNotFound {
		return apperrors.NotFound(notFoundMessage)
	}
	log.Error(internalErrorMEssage)
	return apperrors.Internal(internalErrorMEssage)
}
