// Package metadata contains components for accessing Kyma storage (Application, Minio)
package metadata

import (
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
)

// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// GetAPI gets API of a service with given ID
	GetAPI(appName, serviceID string) (*model.API, apperrors.AppError)
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

// GetAPI gets API of a service with given ID
func (sds *serviceDefinitionService) GetAPI(appName, serviceID string) (*model.API, apperrors.AppError) {
	service, err := sds.applicationRepository.Get(appName, serviceID)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return nil, apperrors.NotFound("service with ID %s not found", serviceID)
		}
		log.Errorf("failed to get service with id '%s': %s", serviceID, err.Error())
		return nil, apperrors.Internal("failed to read %s service, %s", serviceID, err)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("service with ID '%s' has no API", serviceID)
	}

	api, err := sds.serviceAPIService.Read(service.API)
	if err != nil {
		log.Errorf("failed to read api for serviceId '%s': %s", serviceID, err.Error())
		return nil, apperrors.Internal("failed to read API for %s service, %s", serviceID, err)
	}
	return api, nil
}
