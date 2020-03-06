package secrets

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/strategy"

	"github.com/stretchr/testify/require"

	"kyma-project.io/compass-runtime-agent/internal/apperrors"

	"github.com/stretchr/testify/mock"

	strategyMocks "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/strategy/mocks"

	"github.com/stretchr/testify/assert"
	k8smocks "kyma-project.io/compass-runtime-agent/internal/k8sconsts/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/model"
)

const (
	appName    = "app"
	appUID     = types.UID("appUID")
	serviceId  = "serviceID"
	secretName = "secretName"
)

var (
	credentials = &model.CredentialsWithCSRF{
		Oauth: &model.Oauth{
			ClientID:     "clientID",
			ClientSecret: "clientSecret",
			URL:          "http://oauth.com",
		},
	}

	secretData = strategy.SecretData{
		"key":  []byte("value"),
		"key2": []byte("value2"),
	}
)

func TestService_Create(t *testing.T) {

	t.Run("should create secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, appUID, secretName, serviceId, secretData).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should do nothing if credentials are nil", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.ModificationStrategy{}
		strategyFactory := &strategyMocks.Factory{}
		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Create(appName, appUID, serviceId, nil)

		// then
		assert.NoError(t, err)
		strategyFactory.AssertNotCalled(t, "NewSecretModificationStrategy")
		nameResolver.AssertNotCalled(t, "GetResourceName")
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to initialize strategy", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.ModificationStrategy{}
		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(nil, apperrors.Internal("error"))

		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should stop execution if credentials not provided", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(false)
		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		nameResolver.AssertNotCalled(t, "GetResourceName")
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to create secret data", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(nil, apperrors.Internal("error"))

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to create secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, appUID, secretName, serviceId, secretData).Return(apperrors.Internal("error"))

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})
}

func TestService_Update(t *testing.T) {

	t.Run("should update secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)
		modStrategy.On("ShouldUpdate", strategy.SecretData{}, secretData).Return(true)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(strategy.SecretData{}, nil)
		secretsRepository.On("Upsert", appName, appUID, secretName, serviceId, secretData).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should not updated if content the same", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)
		modStrategy.On("ShouldUpdate", secretData, secretData).Return(false)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(secretData, nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should create secret if not found", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(strategy.SecretData{}, apperrors.NotFound("error"))
		secretsRepository.On("Create", appName, appUID, secretName, serviceId, secretData).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(nil, apperrors.Internal(""))

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to update secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)
		modStrategy.On("ShouldUpdate", strategy.SecretData{}, secretData).Return(true)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(strategy.SecretData{}, nil)
		secretsRepository.On("Upsert", appName, appUID, secretName, serviceId, secretData).Return(apperrors.Internal("error"))

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("should delete a secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		strategyFactory := &strategyMocks.Factory{}
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Delete", secretName).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Delete(secretName)

		// then
		require.NoError(t, err)
		assertExpectations(t, &nameResolver.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return an error failed to delete secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		strategyFactory := &strategyMocks.Factory{}
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Delete", secretName).Return(apperrors.Internal("error"))

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		err := service.Delete(secretName)

		// then
		require.Error(t, err)
		assertExpectations(t, &nameResolver.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})
}

func assertExpectations(t *testing.T, mocks ...*mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}
