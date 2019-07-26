package apiresources

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/apiresources/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/applications"
	"k8s.io/apimachinery/pkg/types"
)

type ApiIDToSecretNameMap map[string]string

type ResourcesService interface {
	CreateApiResources(application synchronization.Application, apiDefinition synchronization.APIDefinition) apperrors.AppError
	CreateEventApiResources(application synchronization.Application, eventApiDefinition synchronization.EventAPIDefinition) apperrors.AppError
	CreateSecrets(application synchronization.Application, apiDefinition synchronization.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	UpdateApiResources(application synchronization.Application, apiDefinition synchronization.APIDefinition) apperrors.AppError
	UpdateEventApiResources(application synchronization.Application, eventApiDefinition synchronization.EventAPIDefinition) apperrors.AppError
	UpdateSecrets(application synchronization.Application, apiDefinition synchronization.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	DeleteApiResources(application synchronization.Application, apiDefinition synchronization.APIDefinition) apperrors.AppError
	DeleteEventApiResources(application synchronization.Application, eventApiDefinition synchronization.EventAPIDefinition) apperrors.AppError
	DeleteSecrets(application synchronization.Application, apiDefinition synchronization.APIDefinition) apperrors.AppError
}

type AssetStore interface {
	Create(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Update(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type AccessResources interface {
	Create(applicationName string, applicationUID types.UID, apiID, serviceName string) apperrors.AppError
	Update(applicationName string, applicationUID types.UID, apiID, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

type Secrets interface {
	Create(application string, appUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF) (applications.Credentials, apperrors.AppError)
	Upsert(application string, appUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF) (applications.Credentials, apperrors.AppError)
	Delete(name string) apperrors.AppError
}

type RequestParameters interface {
	Create(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Upsert(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Delete(application, serviceId string) apperrors.AppError
}
