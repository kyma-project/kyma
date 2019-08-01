package apiresources

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/accessservice"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"k8s.io/apimachinery/pkg/types"
)

type ApiIDToSecretNameMap map[string]string

//go:generate mockery -name=Service
type Service interface {
	CreateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError
	UpdateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError
	DeleteApiResources(applicationName string, serviceID string, secretName string) apperrors.AppError
}

type service struct {
	accessServiceManager      accessservice.AccessServiceManager
	secretsService            secrets.Service
	requestParameteresService secrets.RequestParametersService
	nameResolver              k8sconsts.NameResolver
}

// TODO: change secrets.Service interface so that it doesn't return applications.Credentials
func NewService(accessServiceManager accessservice.AccessServiceManager, secretsService secrets.Service, nameResolver k8sconsts.NameResolver) Service {
	return service{
		accessServiceManager: accessServiceManager,
		secretsService:       secretsService,
		nameResolver:         nameResolver,
	}
}

type AccessServiceManager interface {
	Create(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Upsert(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

func (s service) CreateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError {
	k8sResourceName := s.nameResolver.GetResourceName(applicationName, serviceID)
	appendedErr := s.accessServiceManager.Create(applicationName, applicationUID, serviceID, k8sResourceName)
	if credentials != nil {
		_, err := s.secretsService.Create(applicationName, applicationUID, serviceID, credentials)
		if err != nil {
			appendedErr = appendedErr.Append("", err)
		}
	}

	return appendedErr
}

func (s service) UpdateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, spec []byte) apperrors.AppError {
	k8sResourceName := s.nameResolver.GetResourceName(applicationName, serviceID)
	appendedErr := s.accessServiceManager.Upsert(applicationName, applicationUID, serviceID, k8sResourceName)
	if credentials != nil {
		_, err := s.secretsService.Upsert(applicationName, applicationUID, serviceID, credentials)
		if err != nil {
			appendedErr = appendedErr.Append("", err)
		}
	}

	return appendedErr
}

func (s service) DeleteApiResources(applicationName string, serviceID string, secretName string) apperrors.AppError {

	appendedErr := s.accessServiceManager.Delete(s.nameResolver.GetResourceName(applicationName, serviceID))

	if secretName != "" {
		err := s.secretsService.Delete(secretName)
		if err != nil {
			appendedErr = appendedErr.Append("", err)
		}
	}

	return appendedErr
}
