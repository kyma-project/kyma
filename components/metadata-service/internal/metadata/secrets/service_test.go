package secrets

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_CreateOauthSecret(t *testing.T) {
	t.Run("should create oauth secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Create", "re", "name", "serviceID", secretData).Return(
			nil,
		)

		// when
		err := service.CreateOauthSecret(
			"re",
			"name",
			"testID",
			"testSecret",
			"serviceID",
		)

		// then
		assert.NoError(t, err)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error on incomplete secret data", func(t *testing.T) {
		// given
		service := NewService(nil)

		// when
		err := service.CreateOauthSecret(
			"",
			"",
			"testID",
			"testSecret",
			"",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Incomplete secret data.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should return an error if secret already exists", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Create", "re", "name", "serviceID", secretData).Return(
			apperrors.AlreadyExists("Secret already exists."),
		)

		// when
		err := service.CreateOauthSecret(
			"re",
			"name",
			"testID",
			"testSecret",
			"serviceID",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret already exists.", err.Error())
		assert.Equal(t, apperrors.CodeAlreadyExists, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if creation failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Create", "re", "name", "serviceID", secretData).Return(
			apperrors.Internal("Internal error."),
		)

		// when
		err := service.CreateOauthSecret(
			"re",
			"name",
			"testID",
			"testSecret",
			"serviceID",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_GetOauthSecret(t *testing.T) {
	t.Run("should return a secret data", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Get", "re", "name").Return(
			secretData,
			nil,
		)

		// when
		clientId, clientSecret, err := service.GetOauthSecret("re", "name")

		// then
		assert.NoError(t, err)
		assert.Equal(t, "testID", clientId)
		assert.Equal(t, "testSecret", clientSecret)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if secret was not found", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Get", "re", "name").Return(
			map[string][]byte{},
			apperrors.NotFound("Secret not found."),
		)

		// when
		_, _, err := service.GetOauthSecret("re", "name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret not found.", err.Error())
		assert.Equal(t, apperrors.CodeNotFound, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if fetch failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Get", "re", "name").Return(
			map[string][]byte{},
			apperrors.Internal("Internal error."),
		)

		// when
		_, _, err := service.GetOauthSecret("re", "name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_UpdateOauthSecret(t *testing.T) {
	t.Run("should update a secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Upsert", "re", "name", "serviceID", secretData).Return(
			nil,
		)

		// when
		err := service.UpdateOauthSecret(
			"re",
			"name",
			"testID",
			"testSecret",
			"serviceID",
		)

		// then
		assert.NoError(t, err)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error on incomplete secret data", func(t *testing.T) {
		// given
		service := NewService(nil)

		// when
		err := service.UpdateOauthSecret(
			"",
			"",
			"testID",
			"testSecret",
			"",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Incomplete secret data.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should return an error if an update failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Upsert", "re", "name", "serviceID", secretData).Return(
			apperrors.Internal("Internal error."),
		)

		// when
		err := service.UpdateOauthSecret(
			"re",
			"name",
			"testID",
			"testSecret",
			"serviceID",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_CreateBasicAuthSecret(t *testing.T) {
	t.Run("should create basic auth secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeBasicAuthMap("testUsername", "testPassword")
		repositoryMock.On("Create", "re", "name", "serviceID", secretData).Return(
			nil,
		)

		// when
		err := service.CreateBasicAuthSecret("re",
			"name",
			"testUsername",
			"testPassword",
			"serviceID",
		)

		// then
		assert.NoError(t, err)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error on incomplete secret data", func(t *testing.T) {
		// given
		service := NewService(nil)

		// when
		err := service.CreateBasicAuthSecret(
			"",
			"",
			"testUsername",
			"testPassword",
			"",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Incomplete secret data.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should return an error if secret already exists", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeBasicAuthMap("testUsername", "testPassword")
		repositoryMock.On(
			"Create",
			"re", "name",
			"serviceID",
			secretData,
		).Return(apperrors.AlreadyExists("Secret already exists."))

		// when
		err := service.CreateBasicAuthSecret(
			"re",
			"name",
			"testUsername",
			"testPassword",
			"serviceID",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret already exists.", err.Error())
		assert.Equal(t, apperrors.CodeAlreadyExists, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if creation failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeBasicAuthMap("testUsername", "testPassword")
		repositoryMock.On("Create", "re", "name", "serviceID", secretData).Return(
			apperrors.Internal("Internal error."),
		)

		// when
		err := service.CreateBasicAuthSecret(
			"re",
			"name",
			"testUsername",
			"testPassword",
			"serviceID",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_GetBasicAuthSecret(t *testing.T) {
	t.Run("should return a secret data", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeBasicAuthMap("testUsername", "testPassword")
		repositoryMock.On("Get", "re", "name").Return(
			secretData,
			nil,
		)

		// when
		username, password, err := service.GetBasicAuthSecret("re", "name")

		// then
		assert.NoError(t, err)
		assert.Equal(t, "testUsername", username)
		assert.Equal(t, "testPassword", password)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if secret was not found", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Get", "re", "name").Return(
			map[string][]byte{},
			apperrors.NotFound("Secret not found."),
		)

		// when
		_, _, err := service.GetBasicAuthSecret("re", "name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret not found.", err.Error())
		assert.Equal(t, apperrors.CodeNotFound, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if fetch failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Get", "re", "name").Return(
			map[string][]byte{},
			apperrors.Internal("Internal error."),
		)

		// when
		_, _, err := service.GetBasicAuthSecret("re", "name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_UpdateBasicAuthSecret(t *testing.T) {
	t.Run("should update a secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeBasicAuthMap("testUsername", "testPassword")
		repositoryMock.On("Upsert", "re", "name", "serviceID", secretData).Return(
			nil,
		)

		// when
		err := service.UpdateBasicAuthSecret(
			"re",
			"name",
			"testUsername",
			"testPassword",
			"serviceID",
		)

		// then
		assert.NoError(t, err)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error on incomplete secret data", func(t *testing.T) {
		// given
		service := NewService(nil)

		// when
		err := service.UpdateBasicAuthSecret(
			"", "",
			"testUsername",
			"testPassword",
			"",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Incomplete secret data.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should return an error if an update failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		secretData := makeBasicAuthMap("testUsername", "testPassword")
		repositoryMock.On("Upsert", "re", "name", "serviceID", secretData).Return(
			apperrors.Internal("Internal error."),
		)

		// when
		err := service.UpdateBasicAuthSecret(
			"re",
			"name",
			"testUsername",
			"testPassword",
			"serviceID",
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_DeleteSecret(t *testing.T) {
	t.Run("should delete a secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Delete", "name").Return(
			nil,
		)
		// when
		err := service.DeleteSecret("name")

		// then
		assert.NoError(t, err)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if secret was not found", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Delete", "name").Return(
			apperrors.NotFound("Secret was not found."),
		)
		// when
		err := service.DeleteSecret("name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret was not found.", err.Error())
		assert.Equal(t, apperrors.CodeNotFound, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if deletion fails", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock)

		repositoryMock.On("Delete", "name").Return(
			apperrors.Internal("Internal error."),
		)
		// when
		err := service.DeleteSecret("name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}
