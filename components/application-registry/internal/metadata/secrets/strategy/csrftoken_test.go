package strategy

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"

	"github.com/stretchr/testify/assert"
)

const (
	authEndpoint = "authEndpoint"
)

var (
	csrfTokenCredentails = &model.Credentials{
		CSRFToken: &model.CSRFToken{authEndpoint},
	}
)

func TestCSRFToken_ToCredentials(t *testing.T) {

	secretData := map[string][]byte{
		SCRFTokenAuthEndpoint: []byte(authEndpoint),
	}

	t.Run("should convert to credentials", func(t *testing.T) {
		// given
		csrfTokenStrategy := csrfToken{}

		// when
		credentials := csrfTokenStrategy.ToCredentials(secretData, nil)

		// then
		assert.Equal(t, authEndpoint, credentials.CSRFToken.AuthEndpoint)

	})
}

func TestCSRFToken_CredentialsProvided(t *testing.T) {

	testCases := []struct {
		credentials *model.Credentials
		result      bool
	}{
		{
			credentials: &model.Credentials{
				CSRFToken: &model.CSRFToken{authEndpoint},
			},
			result: true,
		},
		{
			credentials: &model.Credentials{
				CSRFToken: &model.CSRFToken{""},
			},
			result: false,
		},
	}

	t.Run("should check if credentials provided", func(t *testing.T) {
		// given
		csrfTokenStrategy := csrfToken{}

		for _, test := range testCases {
			// when
			result := csrfTokenStrategy.CredentialsProvided(test.credentials)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}

func TestCSRFToken_CreateSecretData(t *testing.T) {
	t.Run("should create secret data", func(t *testing.T) {
		// given
		csrfTokenStrategy := csrfToken{}

		// when
		secretData, err := csrfTokenStrategy.CreateSecretData(csrfTokenCredentails)

		//then
		require.NoError(t, err)
		assert.Equal(t, []byte(authEndpoint), secretData[SCRFTokenAuthEndpoint])
	})
}

func TestCSRFToken_ToCredentialsInfo(t *testing.T) {
	t.Run("should convert to app credentials", func(t *testing.T) {
		// given
		csrfTokenStrategy := csrfToken{}

		// when
		appCredentials := csrfTokenStrategy.ToCredentialsInfo(csrfTokenCredentails, secretName)

		// then
		assert.Equal(t, applications.CredentialsCSRFTokenType, appCredentials.Type)
		assert.Equal(t, secretName, appCredentials.SecretName)
		assert.Equal(t, "", appCredentials.AuthenticationUrl)
	})
}

func TestCSRFToken_ShouldUpdate(t *testing.T) {
	testCases := []struct {
		currentData SecretData
		newData     SecretData
		result      bool
	}{
		{
			currentData: SecretData{
				SCRFTokenAuthEndpoint: []byte(authEndpoint),
			},
			newData: SecretData{
				SCRFTokenAuthEndpoint: []byte("changed endpoint"),
			},
			result: true,
		},
		{
			currentData: SecretData{},
			newData: SecretData{
				SCRFTokenAuthEndpoint: []byte(authEndpoint),
			},
			result: true,
		},
		{
			currentData: SecretData{
				SCRFTokenAuthEndpoint: []byte(authEndpoint),
			},
			newData: SecretData{},
			result:  true,
		},
		{
			currentData: SecretData{
				SCRFTokenAuthEndpoint: []byte(authEndpoint),
			},
			newData: SecretData{
				SCRFTokenAuthEndpoint: []byte(authEndpoint),
			},
			result: false,
		},
	}

	t.Run("should return true when update needed", func(t *testing.T) {
		// given
		csrfTokenStrategy := csrfToken{}

		for _, test := range testCases {
			// when
			result := csrfTokenStrategy.ShouldUpdate(test.currentData, test.newData)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}
