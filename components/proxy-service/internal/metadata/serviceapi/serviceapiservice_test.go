package serviceapi

import (
	"testing"

	secretsmocks "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/secrets/mocks"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultService_Read(t *testing.T) {
	t.Run("should read API with oauth credentials", func(t *testing.T) {
		// given
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: &remoteenv.Credentials{
				Type:       "OAuth",
				SecretName: "secret-name",
				Url:        "http://oauth.com",
			},
		}

		secret := map[string][]byte{
			ClientIDKey:     []byte("clientId"),
			ClientSecretKey: []byte("clientSecret"),
		}

		secretsRepository := new(secretsmocks.Repository)
		secretsRepository.On("Get", "secret-name").Return(secret, nil)

		service := NewService(secretsRepository)

		// when
		api, err := service.Read(remoteEnvServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Equal(t, "http://oauth.com", api.Credentials.OAuth.URL)
		assert.Equal(t, "clientId", api.Credentials.OAuth.ClientID)
		assert.Equal(t, "clientSecret", api.Credentials.OAuth.ClientSecret)
		assert.Nil(t, api.Spec)

		secretsRepository.AssertExpectations(t)
	})

	t.Run("should read API without oauth credentials", func(t *testing.T) {
		// given
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
		}

		service := NewService(nil)

		// when
		api, err := service.Read(remoteEnvServiceAPi)

		// then
		require.NoError(t, err)
		assert.Equal(t, "http://target.com", api.TargetUrl)
		assert.Nil(t, api.Credentials)
		assert.Nil(t, api.Spec)
	})

	t.Run("should return error when reading secret fails", func(t *testing.T) {
		// given
		remoteEnvServiceAPi := &remoteenv.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: &remoteenv.Credentials{
				Type:       "OAuth",
				SecretName: "secret-name",
				Url:        "http://oauth.com",
			},
		}

		secretsRepository := new(secretsmocks.Repository)
		secretsRepository.On("Get", "secret-name").
			Return(nil, apperrors.Internal("secret error"))

		service := NewService(secretsRepository)

		// when
		api, err := service.Read(remoteEnvServiceAPi)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "secret error")

		secretsRepository.AssertExpectations(t)
	})
}
