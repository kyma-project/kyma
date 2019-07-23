package apiresources

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/apiresources/assetstore/docstopic"
	"k8s.io/apimachinery/pkg/types"
)

type ApiIDToSecretNameMap map[string]string

type ResourcesService interface {
	CreateApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
	CreateEventApiResources(application compass.Application, eventApiDefinition graphql.EventAPIDefinition) apperrors.AppError
	CreateSecrets(application compass.Application, apiDefinition graphql.APIDefinition) (ApiIDToSecretNameMap, apperrors.AppError)

	UpdateApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
	UpdateEventApiResources(application compass.Application, eventApiDefinition graphql.EventAPIDefinition) apperrors.AppError
	UpdateSecrets(application compass.Application, apiDefinition graphql.APIDefinition) (ApiIDToSecretNameMap, apperrors.AppError)

	DeleteApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
	DeleteEventApiResources(application compass.Application, eventApiDefinition graphql.EventAPIDefinition) apperrors.AppError
	DeleteSecrets(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
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
