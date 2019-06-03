package secrets

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"
	"testing"

	"github.com/stretchr/testify/require"

	k8smocks "github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	requestParametersSecretName = "requestParametersSecretName"
)

var (
	requestParameters = &model.RequestParameters{
		Headers: &map[string][]string{
			"test header key": []string{
				"test header value 1",
			},
		},
		QueryParameters: &map[string][]string{
			"test query parameter key": []string{
				"test query parameter value 1",
			},
		},
	}
)

func TestRequestParametersService_Create(t *testing.T) {

	t.Run("should create the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, requestParametersSecretName, serviceId, secretData).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Create(appName, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should return empty app requestParameters if requestParameters are nil", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		secretsRepository := &mocks.Repository{}

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Create(appName, serviceId, nil)

		// then
		assert.NoError(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should return error when failed to create the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, requestParametersSecretName, serviceId, secretData).Return(apperrors.Internal("error"))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Create(appName, serviceId, requestParameters)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})
}

func TestRequestParametersService_Get(t *testing.T) {

	t.Run("should return request parameters", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(secretData, nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Get(appName, requestParametersSecretName)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParameters.QueryParameters, createdRequestParameters.QueryParameters)
		assert.Equal(t, requestParameters.Headers, createdRequestParameters.Headers)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should return error when failed to get the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(nil, apperrors.Internal(""))
		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Get(appName, requestParametersSecretName)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})
}

func TestRequestParametersService_Upsert(t *testing.T) {

	t.Run("should upsert the secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(strategy.SecretData{}, nil)
		secretsRepository.On("Upsert", appName, requestParametersSecretName, serviceId, secretData).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Upsert(appName, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should not update if content is the same", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(secretData, nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Upsert(appName, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should create the secret if not found", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(strategy.SecretData{}, apperrors.NotFound("error"))
		secretsRepository.On("Create", appName, requestParametersSecretName, serviceId, secretData).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdSecret, err := service.Upsert(appName, serviceId, requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersSecretName, createdSecret)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(nil, apperrors.Internal(""))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Upsert(appName, serviceId, requestParameters)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should return error when failed to update secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(requestParametersSecretName)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", appName, requestParametersSecretName).Return(strategy.SecretData{}, nil)
		secretsRepository.On("Upsert", appName, requestParametersSecretName, serviceId, secretData).Return(apperrors.Internal("error"))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		createdRequestParameters, err := service.Upsert(appName, serviceId, requestParameters)

		// then
		require.Error(t, err)
		assert.Empty(t, createdRequestParameters)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})
}

func TestRequestParametersService_Delete(t *testing.T) {

	t.Run("should delete a secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Delete", requestParametersSecretName).Return(nil)

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		err := service.Delete(requestParametersSecretName)

		// then
		require.NoError(t, err)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})

	t.Run("should return an error failed to delete secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Delete", requestParametersSecretName).Return(apperrors.Internal("error"))

		service := NewRequestParametersService(secretsRepository, nameResolver)

		// when
		err := service.Delete(requestParametersSecretName)

		// then
		require.Error(t, err)
		assertExpectations(t, nameResolver.Mock, secretsRepository.Mock)
	})
}
