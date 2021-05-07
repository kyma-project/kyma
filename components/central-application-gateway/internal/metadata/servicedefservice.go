// Package metadata contains components for accessing Kyma storage (Application, Minio)
package metadata

import (
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
	GetAPI(appName, serviceName, apiName string) (*model.API, apperrors.AppError)
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
func (sds *serviceDefinitionService) GetAPI(appName, serviceName, apiName string) (*model.API, apperrors.AppError) {
	service, err := sds.applicationRepository.Get(appName, serviceName, apiName)

	// will not happen err is always nil
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return nil, apperrors.NotFound("service with name %s not found", serviceName)
		}
		log.Errorf("failed to get service with name '%s': %s", serviceName, err.Error())
		return nil, apperrors.Internal("failed to read %s service, %s", serviceName, err)
	}

	if service.API == nil {
		return nil, apperrors.WrongInput("service with name '%s' and api '%s' has no API", serviceName, apiName)
	}

	api, err := sds.serviceAPIService.Read(service.API)
	if err != nil {
		log.Errorf("failed to read api for serviceId '%s': %s", serviceName, err.Error())
		return nil, apperrors.Internal("failed to read API for %s service, %s", serviceName, err)
	}
	return api, nil
}
