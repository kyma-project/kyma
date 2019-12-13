package secrets

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"

	"github.com/stretchr/testify/mock"

	strategyMocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy/mocks"

	k8smocks "github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/mocks"
	"github.com/stretchr/testify/assert"
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

	appCredentials = applications.Credentials{
		Type:       applications.CredentialsBasicType,
		SecretName: secretName,
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
		modStrategy.On("ToCredentialsInfo", credentials, secretName).Return(appCredentials)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Create", appName, appUID, secretName, serviceId, secretData).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, appCredentials.Type, createdCredentials.Type)
		assert.Equal(t, appCredentials.SecretName, createdCredentials.SecretName)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return empty app credentials if credentials are nil", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.ModificationStrategy{}
		strategyFactory := &strategyMocks.Factory{}
		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Create(appName, appUID, serviceId, nil)

		// then
		assert.NoError(t, err)
		assert.Empty(t, createdCredentials)
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
		createdCredentials, err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return empty app credentials if credentials not provided", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(false)
		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assert.Empty(t, createdCredentials)
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
		createdCredentials, err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
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
		createdCredentials, err := service.Create(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})
}

func TestService_Get(t *testing.T) {

	t.Run("should return credentials", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.AccessStrategy{}
		modStrategy.On("ToCredentials", secretData, &appCredentials).Return(*credentials, nil)
		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretAccessStrategy", &appCredentials).Return(modStrategy, nil)
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(secretData, nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Get(appName, appCredentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, credentials.Oauth.ClientID, createdCredentials.Oauth.ClientID)
		assert.Equal(t, credentials.Oauth.ClientSecret, createdCredentials.Oauth.ClientSecret)
		assert.Equal(t, credentials.Oauth.URL, createdCredentials.Oauth.URL)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to initialize strategy", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.AccessStrategy{}
		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretAccessStrategy", &appCredentials).Return(nil, apperrors.Internal(""))
		secretsRepository := &mocks.Repository{}

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Get(appName, appCredentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		modStrategy := &strategyMocks.AccessStrategy{}
		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretAccessStrategy", &appCredentials).Return(modStrategy, nil)
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(nil, apperrors.Internal(""))
		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Get(appName, appCredentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
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
		modStrategy.On("ToCredentialsInfo", credentials, secretName).Return(appCredentials)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(strategy.SecretData{}, nil)
		secretsRepository.On("Upsert", appName, appUID, secretName, serviceId, secretData).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, appCredentials.Type, createdCredentials.Type)
		assert.Equal(t, appCredentials.SecretName, createdCredentials.SecretName)
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
		modStrategy.On("ToCredentialsInfo", credentials, secretName).Return(appCredentials)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(secretData, nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, appCredentials.Type, createdCredentials.Type)
		assert.Equal(t, appCredentials.SecretName, createdCredentials.SecretName)
		assertExpectations(t, &nameResolver.Mock, &modStrategy.Mock, &strategyFactory.Mock, &secretsRepository.Mock)
	})

	t.Run("should create secret if not found", func(t *testing.T) {
		// given
		nameResolver := &k8smocks.NameResolver{}
		nameResolver.On("GetResourceName", appName, serviceId).Return(secretName)

		modStrategy := &strategyMocks.ModificationStrategy{}
		modStrategy.On("CredentialsProvided", credentials).Return(true)
		modStrategy.On("CreateSecretData", credentials).Return(secretData, nil)
		modStrategy.On("ToCredentialsInfo", credentials, secretName).Return(appCredentials)

		strategyFactory := &strategyMocks.Factory{}
		strategyFactory.On("NewSecretModificationStrategy", credentials).Return(modStrategy, nil)

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", secretName).Return(strategy.SecretData{}, apperrors.NotFound("error"))
		secretsRepository.On("Create", appName, appUID, secretName, serviceId, secretData).Return(nil)

		service := NewService(secretsRepository, nameResolver, strategyFactory)

		// when
		createdCredentials, err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, appCredentials.Type, createdCredentials.Type)
		assert.Equal(t, appCredentials.SecretName, createdCredentials.SecretName)
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
		createdCredentials, err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
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
		createdCredentials, err := service.Upsert(appName, appUID, serviceId, credentials)

		// then
		require.Error(t, err)
		assert.Empty(t, createdCredentials)
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
