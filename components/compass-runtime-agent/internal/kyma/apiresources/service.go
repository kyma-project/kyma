package apiresources

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/accessservice"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"k8s.io/apimachinery/pkg/types"
)

type ApiIDToSecretNameMap map[string]string

//go:generate mockery -name=Service
type Service interface {
	CreateApiResources(applicationName string, applicationUID types.UID, serviceID, serviceName string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError
	UpdateApiResources(applicationName string, applicationUID types.UID, serviceID, serviceName string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError
	DeleteApiResources(application v1alpha1.Application, apiDefinition v1alpha1.Service) apperrors.AppError
}

//go:generate mockery -name=AssetStore
type AssetStore interface {
	Create(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Update(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Delete(id string) apperrors.AppError
}

//go:generate mockery -name=AccessResources
type AccessResources interface {
	Create(applicationName string, applicationUID types.UID, apiID, serviceName string) apperrors.AppError
	Update(applicationName string, applicationUID types.UID, apiID, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

//go:generate mockery -name=Secrets
type Secrets interface {
	Create(application string, appUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF) (applications.Credentials, apperrors.AppError)
	Upsert(application string, appUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF) (applications.Credentials, apperrors.AppError)
	Delete(name string) apperrors.AppError
}

//go:generate mockery -name=RequestParameters
type RequestParameters interface {
	Create(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Upsert(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Delete(application, serviceId string) apperrors.AppError
}

type service struct {
	accessServiceManager      accessservice.AccessServiceManager
	secretsService            secrets.Service
	requestParameteresService secrets.RequestParametersService
}

// TODO: change secrets.Service interface so that it doesn't return applications.Credentials
func NewService(accessServiceManager accessservice.AccessServiceManager, secretsService secrets.Service) Service {
	return service{
		accessServiceManager: accessServiceManager,
		secretsService:       secretsService,
	}
}

type AccessServiceManager interface {
	Create(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Upsert(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

func (s service) CreateApiResources(applicationName string, applicationUID types.UID, serviceID, serviceName string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError {
	appendedErr := s.accessServiceManager.Create(applicationName, applicationUID, serviceID, serviceName)
	if credentials != nil {
		_, err := s.secretsService.Create(applicationName, applicationUID, serviceID, credentials)
		if err != nil {
			appendedErr = appendError(appendedErr, err)
		}
	}

	return appendedErr
}

func (s service) UpdateApiResources(applicationName string, applicationUID types.UID, serviceID, serviceName string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError {

	appendedErr := s.accessServiceManager.Upsert(applicationName, applicationUID, serviceID, serviceName)
	if credentials != nil {
		_, err := s.secretsService.Upsert(applicationName, applicationUID, serviceID, credentials)
		if err != nil {
			appendedErr = appendError(appendedErr, err)
		}
	}

	return appendedErr
}

func (s service) DeleteApiResources(application v1alpha1.Application, apiDefinition v1alpha1.Service) apperrors.AppError {
	return nil
}

func appendError(wrapped apperrors.AppError, new apperrors.AppError) apperrors.AppError {
	if wrapped == nil {
		return new
	}

	return wrapped.Append("", new)
}
