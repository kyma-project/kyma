package appsecrets

import (
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

type requestParametersSecretModificationFunction func(application string, appUID types.UID, name, packageID string, newData map[string][]byte) apperrors.AppError

//go:generate mockery --name RequestParametersService
type RequestParametersService interface {
	Get(secretName string) (*model.RequestParameters, apperrors.AppError)
	Create(application string, appUID types.UID, packageID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Upsert(application string, appUID types.UID, packageID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Delete(secretName string) apperrors.AppError
}

type requestParametersService struct {
	repository   Repository
	nameResolver k8sconsts.NameResolver
}

func NewRequestParametersService(repository Repository, nameResolver k8sconsts.NameResolver) RequestParametersService {
	return &requestParametersService{
		repository:   repository,
		nameResolver: nameResolver,
	}
}

func (s *requestParametersService) Create(application string, appUID types.UID, packageID string, requestParameters *model.RequestParameters) (string, apperrors.AppError) {
	return s.modifySecret(application, appUID, packageID, requestParameters, s.createSecret)
}

func (s *requestParametersService) Get(secretName string) (*model.RequestParameters, apperrors.AppError) {
	data, err := s.repository.Get(secretName)
	if err != nil {
		return nil, err
	}

	return MapToRequestParameters(data)
}

func (s *requestParametersService) Upsert(application string, appUID types.UID, packageID string, requestParameters *model.RequestParameters) (string, apperrors.AppError) {
	return s.modifySecret(application, appUID, packageID, requestParameters, s.upsertSecret)
}

func (s *requestParametersService) Delete(secretName string) apperrors.AppError {
	return s.repository.Delete(secretName)
}

func (s *requestParametersService) modifySecret(application string, appUID types.UID, packageID string, requestParameters *model.RequestParameters, modFunction requestParametersSecretModificationFunction) (string, apperrors.AppError) {
	if requestParameters == nil || requestParameters.IsEmpty() {
		return "", nil
	}

	name := s.createSecretName(application, packageID)

	secretData, err := RequestParametersToMap(requestParameters)
	if err != nil {
		return "", err.Append("Failed to create request parameters secret data")
	}

	err = modFunction(application, appUID, name, packageID, secretData)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (s *requestParametersService) upsertSecret(application string, appUID types.UID, name, packageID string, newData map[string][]byte) apperrors.AppError {
	return s.repository.Upsert(application, appUID, name, packageID, newData)
}

func (s *requestParametersService) createSecret(application string, appUID types.UID, name, packageID string, newData map[string][]byte) apperrors.AppError {
	return s.repository.Create(application, appUID, name, packageID, newData)
}

func (s *requestParametersService) createSecretName(application, packageID string) string {
	return s.nameResolver.GetRequestParametersSecretName(application, packageID)
}
