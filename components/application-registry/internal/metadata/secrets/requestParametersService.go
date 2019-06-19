package secrets

import (
	"encoding/json"
	"fmt"

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

type requestParametersSecretModificationFunction func(application, name, serviceID string, newData map[string][]byte) apperrors.AppError

type RequestParametersService interface {
	Get(secretName string) (model.RequestParameters, apperrors.AppError)
	Create(application, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
	Upsert(application, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError)
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

func (s *requestParametersService) Create(application, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError) {
	return s.modifySecret(application, serviceID, requestParameters, s.createSecret)
}

func (s *requestParametersService) Get(secretName string) (model.RequestParameters, apperrors.AppError) {
	data, err := s.repository.Get(secretName)
	if err != nil {
		return model.RequestParameters{}, err
	}

	return dataToRequestParameters(data)
}

func dataToRequestParameters(data map[string][]byte) (model.RequestParameters, apperrors.AppError) {
	headers, err := getParameterFromJsonData(data, requestParametersHeadersKey)
	if err != nil {
		return model.RequestParameters{}, nil
	}
	queryParameters, err := getParameterFromJsonData(data, requestParametersQueryParametersKey)
	if err != nil {
		return model.RequestParameters{}, nil
	}

	return model.RequestParameters{
		Headers:         &headers,
		QueryParameters: &queryParameters,
	}, nil
}

func getParameterFromJsonData(data map[string][]byte, key string) (map[string][]string, apperrors.AppError) {
	parameter := make(map[string][]string)
	if err := json.Unmarshal(data[key], &parameter); err != nil {
		return map[string][]string{}, apperrors.Internal("%s", err)
	}
	return parameter, nil
}

func (s *requestParametersService) Upsert(application, serviceID string, requestParameters *model.RequestParameters) (string, apperrors.AppError) {
	return s.modifySecret(application, serviceID, requestParameters, s.upsertSecret)
}

func (s *requestParametersService) Delete(application string, serviceId string) apperrors.AppError {
	secretName := s.createSecretName(application, serviceId)

	return s.repository.Delete(secretName)
}

func (s *requestParametersService) modifySecret(application, serviceID string, requestParameters *model.RequestParameters, modFunction requestParametersSecretModificationFunction) (string, apperrors.AppError) {
	if requestParameters == nil {
		return "", nil
	}

	name := s.createSecretName(application, serviceID)

	secretData, err := createSecretData(requestParameters)
	if err != nil {
		return "", err.Append("Failed to create request parameters secret data")
	}

	err = modFunction(application, name, serviceID, secretData)
	if err != nil {
		return "", err
	}

	return name, nil
}

func createSecretData(requestParameters *model.RequestParameters) (map[string][]byte, apperrors.AppError) {
	data := make(map[string][]byte)
	if requestParameters.Headers != nil {
		headers, _ := json.Marshal(requestParameters.Headers)
		data[requestParametersHeadersKey] = headers
	}
	if requestParameters.QueryParameters != nil {
		queryParameters, _ := json.Marshal(requestParameters.QueryParameters)
		data[requestParametersQueryParametersKey] = queryParameters
	}

	return data, nil
}

func (s *requestParametersService) upsertSecret(application, name, serviceID string, newData map[string][]byte) apperrors.AppError {
	return s.repository.Upsert(application, name, serviceID, newData)
}

func (s *requestParametersService) createSecret(application, name, serviceID string, newData map[string][]byte) apperrors.AppError {
	return s.repository.Create(application, name, serviceID, newData)
}

func (s *requestParametersService) createSecretName(application, serviceId string) string {
	name := s.nameResolver.GetResourceName(application, serviceId)

	resourceName := fmt.Sprintf(requestParamsNameFormat, name)
	if len(resourceName) > k8sResourceNameMaxLength {
		return resourceName[0 : k8sResourceNameMaxLength-1]
	}

	return resourceName
}
