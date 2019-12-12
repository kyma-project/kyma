package secrets

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	k8sResourceNameMaxLength = 64

	requestParamsNameFormat = "params-%s"

	requestParametersHeadersKey         = "headers"
	requestParametersQueryParametersKey = "queryParameters"
)

type requestParametersSecretModificationFunction func(application string, appUID types.UID, name, serviceID string, newData map[string][]byte) apperrors.AppError

type RequestParametersService interface {
	Get(secretName string) (*model.RequestParameters, apperrors.AppError)
	Create(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Upsert(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Delete(application, serviceId string) apperrors.AppError
}

type requestParametersService struct {
	nameResolver k8sconsts.NameResolver
	repository   Repository
}

func NewRequestParametersService(repository Repository, nameResolver k8sconsts.NameResolver) RequestParametersService {
	return &requestParametersService{
		nameResolver: nameResolver,
		repository:   repository,
	}
}

func (s *requestParametersService) Create(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError) {
	return s.modifySecret(application, appUID, serviceID, requestParameters, s.createSecret)
}

func (s *requestParametersService) Get(secretName string) (*model.RequestParameters, apperrors.AppError) {
	data, err := s.repository.Get(secretName)
	if err != nil {
		return nil, err
	}

	return model.MapToRequestParameters(data)
}

func (s *requestParametersService) Upsert(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError) {
	return s.modifySecret(application, appUID, serviceID, requestParameters, s.upsertSecret)
}

func (s *requestParametersService) Delete(application string, serviceId string) apperrors.AppError {
	secretName := s.createSecretName(application, serviceId)

	return s.repository.Delete(secretName)
}

func (s *requestParametersService) modifySecret(application string, appUID types.UID, serviceID string, requestParameters *model.RequestParameters, modFunction requestParametersSecretModificationFunction) (string, apperrors.AppError) {
	if requestParameters == nil {
		return "", nil
	}

	name := s.createSecretName(application, serviceID)

	secretData, err := model.RequestParametersToMap(requestParameters)
	if err != nil {
		return "", err.Append("Failed to create request parameters secret data")
	}

	err = modFunction(application, appUID, name, serviceID, secretData)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (s *requestParametersService) upsertSecret(application string, appUID types.UID, name, serviceID string, newData map[string][]byte) apperrors.AppError {
	return s.repository.Upsert(application, appUID, name, serviceID, newData)
}

func (s *requestParametersService) createSecret(application string, appUID types.UID, name, serviceID string, newData map[string][]byte) apperrors.AppError {
	return s.repository.Create(application, appUID, name, serviceID, newData)
}

func (s *requestParametersService) createSecretName(application, serviceId string) string {
	name := s.nameResolver.GetResourceName(application, serviceId)

	resourceName := fmt.Sprintf(requestParamsNameFormat, name)
	if len(resourceName) > k8sResourceNameMaxLength {
		return resourceName[0 : k8sResourceNameMaxLength-1]
	}

	return resourceName
}
