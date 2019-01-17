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
	basicCredentials = &model.Credentials{
		Basic: &model.Basic{Username: username, Password: password},
	}
)

func TestBasicAuth_ToCredentials(t *testing.T) {

	secretData := map[string][]byte{
		BasicAuthUsernameKey: []byte(username),
		BasicAuthPasswordKey: []byte(password),
	}

	t.Run("should convert to basicCredentials", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		// when
		credentials := basicAuthStrategy.ToCredentials(secretData, nil)

		// then
		assert.Equal(t, username, credentials.Basic.Username)
		assert.Equal(t, password, credentials.Basic.Password)

	})
}

func TestBasicAuth_CredentialsProvided(t *testing.T) {

	testCases := []struct {
		credentials *model.Credentials
		result      bool
	}{
		{
			credentials: &model.Credentials{
				Basic: &model.Basic{Username: username, Password: password},
			},
			result: true,
		},
		{
			credentials: &model.Credentials{
				Basic: &model.Basic{Username: "", Password: password},
			},
			result: false,
		},
		{
			credentials: &model.Credentials{
				Basic: &model.Basic{Username: username, Password: ""},
			},
			result: false,
		},
		{
			credentials: nil,
			result:      false,
		},
	}

	t.Run("should check if basicCredentials provided", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		for _, test := range testCases {
			// when
			result := basicAuthStrategy.CredentialsProvided(test.credentials)

			// then
			assert.Equal(t, test.result, result)
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

func TestBasicAuth_ToAppCredentials(t *testing.T) {
	t.Run("should convert to app basicCredentials", func(t *testing.T) {
		// given
		basicAuthStrategy := basicAuth{}

		// when
		appCredentials := basicAuthStrategy.ToAppCredentials(basicCredentials, secretName)

		// then
		assert.Equal(t, applications.CredentialsBasicType, appCredentials.Type)
		assert.Equal(t, secretName, appCredentials.SecretName)
		assert.Equal(t, "", appCredentials.AuthenticationUrl)
	})
}

//testCases := []struct {
//secretData     map[string][]byte
//appCredentials *applications.Credentials
//}{
//{
//secretData: map[string][]byte{
//BasicAuthUsernameKey: []byte(username),
//BasicAuthPasswordKey: []byte(password),
//},
//appCredentials:nil,
//},
//{
//secretData: map[string][]byte{},
//appCredentials:nil,
//},
//}
