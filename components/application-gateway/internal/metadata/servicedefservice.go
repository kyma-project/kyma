// Package metadata contains components for accessing Kyma storage (Application, Minio)
package metadata

import (
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
)

// ServiceDefinitionService is a service that manages ServiceDefinition objects.
type ServiceDefinitionService interface {
	// GetAPI gets API of a service with given ID
	GetAPI(serviceId string) (*model.API, apperrors.AppError)
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
func (sds *serviceDefinitionService) GetAPI(serviceId string) (*model.API, apperrors.AppError) {
	service, err := sds.applicationRepository.Get(serviceId)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return nil, apperrors.NotFound("service with ID %s not found", serviceId)
		}
		log.Errorf("failed to get service with id '%s': %s", serviceId, err.Error())
		return nil, apperrors.Internal("failed to read %s service, %s", serviceId, err)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("service with ID '%s' has no API", serviceId)
	}

	api, err := sds.serviceAPIService.Read(service.API)
	if err != nil {
		log.Errorf("failed to read api for serviceId '%s': %s", serviceId, err.Error())
		return nil, apperrors.Internal("failed to read API for %s service, %s", serviceId, err)
	}
	return api, nil
}
