package strategy

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"

	"github.com/stretchr/testify/assert"
)

const (
	username   = "username"
	password   = "password"
	secretName = "secretName"
)

var (
	basicCredentials = &model.CredentialsWithCSRF{
		Basic: &model.Basic{
			Username: username, Password: password,
		},
	}
)

func TestBasicAuth_ToCredentials(t *testing.T) {

	secretData := map[string][]byte{
		BasicAuthUsernameKey: []byte(username),
		BasicAuthPasswordKey: []byte(password),
	}

	t.Run("should convert to credentials", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		// when
		credentials, err := basicAuthStrategy.ToCredentials(secretData, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, username, credentials.Basic.Username)
		assert.Equal(t, password, credentials.Basic.Password)

	})

	t.Run("should convert to credentials with CSRFInfo", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		// when
		credentials, err := basicAuthStrategy.ToCredentials(secretData, &applications.Credentials{CSRFInfo: &applications.CSRFInfo{TokenEndpointURL: "https://test.it"}})

		// then
		require.NoError(t, err)
		assert.Equal(t, username, credentials.Basic.Username)
		assert.Equal(t, password, credentials.Basic.Password)
		assert.NotNil(t, credentials.Basic)
		assert.NotNil(t, credentials.CSRFInfo)
		assert.Equal(t, "https://test.it", credentials.CSRFInfo.TokenEndpointURL)

	})
}

func TestBasicAuth_CredentialsProvided(t *testing.T) {

	testCases := []struct {
		credentials *model.CredentialsWithCSRF
		result      bool
	}{
		{
			credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: username,
					Password: password},
			},
			result: true,
		},
		{
			credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: "",
					Password: password},
			},
			result: false,
		},
		{
			credentials: &model.CredentialsWithCSRF{
				Basic: &model.Basic{
					Username: username,
					Password: ""},
			},
			result: false,
		},
		{
			credentials: nil,
			result:      false,
		},
	}

	t.Run("should check if credentials provided", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		for _, test := range testCases {
			// when
			result := basicAuthStrategy.CredentialsProvided(test.credentials)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}

func TestBasicAuth_CreateSecretData(t *testing.T) {
	t.Run("should create secret data", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		// when
		secretData, err := basicAuthStrategy.CreateSecretData(basicCredentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, []byte(username), secretData[BasicAuthUsernameKey])
		assert.Equal(t, []byte(password), secretData[BasicAuthPasswordKey])
	})
}

func TestBasicAuth_ToCredentialsInfo(t *testing.T) {
	t.Run("should convert to app credentials", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		// when
		appCredentials := basicAuthStrategy.ToCredentialsInfo(basicCredentials, secretName)

		// then
		assert.Equal(t, applications.CredentialsBasicType, appCredentials.Type)
		assert.Equal(t, secretName, appCredentials.SecretName)
		assert.Equal(t, "", appCredentials.AuthenticationUrl)
	})
}

func TestBasicAuth_ShouldUpdate(t *testing.T) {
	testCases := []struct {
		currentData SecretData
		newData     SecretData
		result      bool
	}{
		{
			currentData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte(password),
			},
			newData: SecretData{
				BasicAuthUsernameKey: []byte("changed username"),
				BasicAuthPasswordKey: []byte(password),
			},
			result: true,
		},
		{
			currentData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte(password),
			},
			newData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte("changed password"),
			},
			result: true,
		},
		{
			currentData: SecretData{},
			newData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte(password),
			},
			result: true,
		},
		{
			currentData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte(password),
			},
			newData: SecretData{},
			result:  true,
		},
		{
			currentData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte(password),
			},
			newData: SecretData{
				BasicAuthUsernameKey: []byte(username),
				BasicAuthPasswordKey: []byte(password),
			},
			result: false,
		},
	}

	t.Run("should return true when update needed", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		for _, test := range testCases {
			// when
			result := basicAuthStrategy.ShouldUpdate(test.currentData, test.newData)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}
