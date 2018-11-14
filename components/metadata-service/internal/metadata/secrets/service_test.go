package secrets

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	k8smocks "github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts/mocks"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Create(t *testing.T) {
	t.Run("should create oauth secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(&repositoryMock, &nameResolverMock)

		data := makeOauthMap("clientID", "clientSecret")

		repositoryMock.On("Create", "re", "resourceName", "serviceID", data).Return(
			nil,
		)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		credentials := &model.Credentials{
			Oauth: &model.Oauth{
				ClientID:     "clientID",
				ClientSecret: "clientSecret",
				URL:          "http://oauth.com",
			},
		}

		// when
		res, err := service.Create(
			"re",
			"serviceID",
			credentials,
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, "http://oauth.com", res.AuthenticationUrl)
		assert.Equal(t, remoteenv.CredentialsOAuthType, res.Type)
		assert.Equal(t, "resourceName", res.SecretName)

		repositoryMock.AssertExpectations(t)
		nameResolverMock.AssertExpectations(t)
	})

	t.Run("should create basic auth secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(&repositoryMock, &nameResolverMock)

		data := makeBasicAuthMap("username", "password")

		repositoryMock.On("Create", "re", "resourceName", "serviceID", data).Return(
			nil,
		)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		// when
		res, err := service.Create(
			"re",
			"serviceID",
			credentials,
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, "", res.AuthenticationUrl)
		assert.Equal(t, remoteenv.CredentialsBasicType, res.Type)
		assert.Equal(t, "resourceName", res.SecretName)

		repositoryMock.AssertExpectations(t)
		nameResolverMock.AssertExpectations(t)
	})

	t.Run("should return an error on incomplete secret data", func(t *testing.T) {
		// given
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(nil, &nameResolverMock)

		nameResolverMock.On("GetResourceName", "", "").Return("")

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		// when
		_, err := service.Create(
			"",
			"",
			credentials,
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Incomplete secret data.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should return an error if secret already exists", func(t *testing.T) {
		// given
		nameResolverMock := k8smocks.NameResolver{}
		repositoryMock := mocks.Repository{}

		service := NewService(&repositoryMock, &nameResolverMock)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		secretData := makeBasicAuthMap("username", "password")
		repositoryMock.On("Create", "re", "resourceName", "serviceID", secretData).Return(
			apperrors.AlreadyExists("Secret already exists."),
		)

		// when
		_, err := service.Create(
			"re",
			"serviceID",
			credentials,
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
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(&repositoryMock, &nameResolverMock)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		secretData := makeBasicAuthMap("username", "password")
		repositoryMock.On("Create", "re", "resourceName", "serviceID", secretData).Return(
			apperrors.Internal("Internal error."),
		)

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		// when
		_, err := service.Create(
			"re",
			"serviceID",
			credentials,
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_Get(t *testing.T) {
	t.Run("should return oauth secret data", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		secretData := makeOauthMap("testID", "testSecret")
		repositoryMock.On("Get", "re", "name").Return(
			secretData,
			nil,
		)

		credentials := remoteenv.Credentials{
			Type:       remoteenv.CredentialsOAuthType,
			SecretName: "name",
		}

		// when
		res, err := service.Get("re", credentials)

		// then
		assert.NoError(t, err)
		assert.Equal(t, "testID", res.Oauth.ClientID)
		assert.Equal(t, "testSecret", res.Oauth.ClientSecret)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return basic auth secret data", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		secretData := makeBasicAuthMap("username", "password")
		repositoryMock.On("Get", "re", "name").Return(
			secretData,
			nil,
		)

		credentials := remoteenv.Credentials{
			Type:       remoteenv.CredentialsBasicType,
			SecretName: "name",
		}

		// when
		res, err := service.Get("re", credentials)

		// then
		assert.NoError(t, err)
		assert.Equal(t, "username", res.Basic.Username)
		assert.Equal(t, "password", res.Basic.Password)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if secret was not found", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		repositoryMock.On("Get", "re", "name").Return(
			map[string][]byte{},
			apperrors.NotFound("Secret not found."),
		)

		credentials := remoteenv.Credentials{
			Type:       remoteenv.CredentialsOAuthType,
			SecretName: "name",
		}

		// when
		_, err := service.Get("re", credentials)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret not found.", err.Error())
		assert.Equal(t, apperrors.CodeNotFound, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if fetch failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		repositoryMock.On("Get", "re", "name").Return(
			map[string][]byte{},
			apperrors.Internal("Internal error."),
		)

		credentials := remoteenv.Credentials{
			Type:       remoteenv.CredentialsOAuthType,
			SecretName: "name",
		}

		// when
		_, err := service.Get("re", credentials)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_Update(t *testing.T) {
	t.Run("should update oauth secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(&repositoryMock, &nameResolverMock)

		data := makeOauthMap("clientID", "clientSecret")

		repositoryMock.On("Upsert", "re", "resourceName", "serviceID", data).Return(
			nil,
		)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		credentials := &model.Credentials{
			Oauth: &model.Oauth{
				ClientID:     "clientID",
				ClientSecret: "clientSecret",
				URL:          "http://oauth.com",
			},
		}

		// when
		res, err := service.Update(
			"re",
			"serviceID",
			credentials,
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, "http://oauth.com", res.AuthenticationUrl)
		assert.Equal(t, remoteenv.CredentialsOAuthType, res.Type)

		repositoryMock.AssertExpectations(t)
		nameResolverMock.AssertExpectations(t)
	})

	t.Run("should update basic auth secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(&repositoryMock, &nameResolverMock)

		data := makeBasicAuthMap("username", "password")

		repositoryMock.On("Upsert", "re", "resourceName", "serviceID", data).Return(
			nil,
		)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		// when
		res, err := service.Update(
			"re",
			"serviceID",
			credentials,
		)

		// then
		assert.NoError(t, err)
		assert.Equal(t, remoteenv.CredentialsBasicType, res.Type)

		repositoryMock.AssertExpectations(t)
		nameResolverMock.AssertExpectations(t)
	})

	t.Run("should return an error on incomplete secret data", func(t *testing.T) {
		// given
		nameResolverMock := k8smocks.NameResolver{}
		service := NewService(nil, &nameResolverMock)

		nameResolverMock.On("GetResourceName", "", "").Return("")

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		// when
		_, err := service.Update(
			"",
			"",
			credentials,
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Incomplete secret data.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should return an error if an update failed", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		nameResolverMock := k8smocks.NameResolver{}

		service := NewService(&repositoryMock, &nameResolverMock)

		secretData := makeBasicAuthMap("username", "password")
		repositoryMock.On("Upsert", "re", "resourceName", "serviceID", secretData).Return(
			apperrors.Internal("Internal error."),
		)

		nameResolverMock.On("GetResourceName", "re", "serviceID").Return("resourceName")

		credentials := &model.Credentials{
			Basic: &model.Basic{
				Username: "username",
				Password: "password",
			},
		}

		// when
		_, err := service.Update(
			"re",
			"serviceID",
			credentials,
		)

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("should delete a secret", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		repositoryMock.On("Delete", "name").Return(
			nil,
		)
		// when
		err := service.Delete("name")

		// then
		assert.NoError(t, err)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if secret was not found", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		repositoryMock.On("Delete", "name").Return(
			apperrors.NotFound("Secret was not found."),
		)
		// when
		err := service.Delete("name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Secret was not found.", err.Error())
		assert.Equal(t, apperrors.CodeNotFound, err.Code())

		repositoryMock.AssertExpectations(t)
	})

	t.Run("should return an error if deletion fails", func(t *testing.T) {
		// given
		repositoryMock := mocks.Repository{}
		service := NewService(&repositoryMock, nil)

		repositoryMock.On("Delete", "name").Return(
			apperrors.Internal("Internal error."),
		)
		// when
		err := service.Delete("name")

		// then
		assert.Error(t, err)
		assert.Equal(t, "Internal error.", err.Error())
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		repositoryMock.AssertExpectations(t)
	})
}
