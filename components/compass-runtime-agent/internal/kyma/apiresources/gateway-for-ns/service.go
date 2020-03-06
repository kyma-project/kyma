package gateway_for_ns

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

type Service interface {
	CreateAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError
	UpsertAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError
	DeleteAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError
	DeleteResourcesOfNonExistentAPI(existentRuntimeApplication v1alpha1.Application, directorApplication model.Application, name string) apperrors.AppError
}
