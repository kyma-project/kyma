package secrets

import (
	"fmt"
	"testing"

	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/model"

	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/strategy"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	k8smocks "kyma-project.io/compass-runtime-agent/internal/k8sconsts/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/mocks"
)

var (
	requestParameters = &model.RequestParameters{
		Headers: &map[string][]string{
			"TestHeader": []string{
				"header value",
			},
		},
		QueryParameters: &map[string][]string{
			"testQueryParam": []string{
				"query parameter value",
			},
		},
	}

	requestParamsSecretData = strategy.SecretData{
		"headers":         []byte(`{"TestHeader":["header value"]}`),
		"queryParameters": []byte(`{"testQueryParam":["query parameter value"]}`),
	}

	baseResourceName            = fmt.Sprintf("%s-%s", appName, serviceId)
	requestParametersSecretName = fmt.Sprintf("params-%s-%s", appName, serviceId)
)

func TestRequestParametersService_Create(t *testing.T) {

	t.Run("should create the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, appUID, requestParametersSecretName, serviceId, requestParamsSecretData).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Create(appName, appUID, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})

	t.Run("should return empty app requestParameters if requestParameters are nil", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		secretsRepository := &mocks.Repository{}

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Create(appName, appUID, serviceId, nil)

		// then
		assert.NoError(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to create the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, appUID, requestParametersSecretName, serviceId, requestParamsSecretData).Return(apperrors.Internal("error"))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Create(appName, appUID, serviceId, requestParameters)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})
}

func TestRequestParametersService_Get(t *testing.T) {

	t.Run("should return request parameters", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", requestParametersSecretName).Return(requestParamsSecretData, nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Get(requestParametersSecretName)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParameters.QueryParameters, createdRequestParameters.QueryParameters)
		assert.Equal(t, requestParameters.Headers, createdRequestParameters.Headers)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to get the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", requestParametersSecretName).Return(nil, apperrors.Internal(""))
		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Get(requestParametersSecretName)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})
}

func TestRequestParametersService_Upsert(t *testing.T) {

	t.Run("should upsert the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Upsert", appName, appUID, requestParametersSecretName, serviceId, requestParamsSecretData).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Upsert(appName, appUID, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})

	t.Run("should create the secret if not found", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Upsert", appName, appUID, requestParametersSecretName, serviceId, requestParamsSecretData).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Upsert(appName, appUID, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to update secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Upsert", appName, appUID, requestParametersSecretName, serviceId, requestParamsSecretData).Return(apperrors.Internal("error"))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Upsert(appName, appUID, serviceId, requestParameters)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})
}

func TestRequestParametersService_Delete(t *testing.T) {

	t.Run("should delete a secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Delete", requestParametersSecretName).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		err := service.Delete(appName, serviceId)

		// then
		require.NoError(t, err)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})

	t.Run("should return an error failed to delete secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(baseResourceName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Delete", requestParametersSecretName).Return(apperrors.Internal("error"))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		err := service.Delete(appName, serviceId)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &secretsRepository.Mock)
	})
}
