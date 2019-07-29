package apiresources

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	syncmodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"k8s.io/apimachinery/pkg/types"
)

type ApiIDToSecretNameMap map[string]string

//go:generate mockery -name=ResourcesService
type ResourcesService interface {
	CreateApiResources(application syncmodel.Application, apiDefinition syncmodel.APIDefinition) apperrors.AppError
	CreateEventApiResources(application syncmodel.Application, eventApiDefinition syncmodel.EventAPIDefinition) apperrors.AppError
	CreateSecrets(application syncmodel.Application, apiDefinition syncmodel.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	UpdateApiResources(application syncmodel.Application, apiDefinition syncmodel.APIDefinition) apperrors.AppError
	UpdateEventApiResources(application syncmodel.Application, eventApiDefinition syncmodel.EventAPIDefinition) apperrors.AppError
	UpdateSecrets(application syncmodel.Application, apiDefinition syncmodel.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	DeleteApiResources(application syncmodel.Application, apiDefinition syncmodel.APIDefinition) apperrors.AppError
	DeleteEventApiResources(application syncmodel.Application, eventApiDefinition syncmodel.EventAPIDefinition) apperrors.AppError
	DeleteSecrets(application syncmodel.Application, apiDefinition syncmodel.APIDefinition) apperrors.AppError
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
