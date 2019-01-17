package strategy

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"

	"github.com/stretchr/testify/assert"
)

const (
	clientId     = "clientId"
	clientSecret = "clientSecret"
	oauthUrl     = "oauthUrl"
)

var (
	oauthCredentials = &model.Credentials{
		Oauth: &model.Oauth{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			URL:          oauthUrl,
		},
	}
)

func TestOauth_ToCredentials(t *testing.T) {

	secretData := map[string][]byte{
		OauthClientIDKey:     []byte(clientId),
		OauthClientSecretKey: []byte(clientSecret),
	}

	appCredentials := &applications.Credentials{
		AuthenticationUrl: oauthUrl,
	}

	t.Run("should convert to basicCredentials", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		// when
		credentials := oauthStrategy.ToCredentials(secretData, appCredentials)

		// then
		assert.Equal(t, clientId, credentials.Oauth.ClientID)
		assert.Equal(t, clientSecret, credentials.Oauth.ClientSecret)
		assert.Equal(t, oauthUrl, credentials.Oauth.URL)
	})
}

func TestOauth_CredentialsProvided(t *testing.T) {

	testCases := []struct {
		credentials *model.Credentials
		result      bool
	}{
		{
			credentials: &model.Credentials{
				Oauth: &model.Oauth{ClientID: clientId, ClientSecret: clientSecret},
			},
			result: true,
		},
		{
			credentials: &model.Credentials{
				Oauth: &model.Oauth{ClientID: "", ClientSecret: clientSecret},
			},
			result: false,
		},
		{
			credentials: &model.Credentials{
				Oauth: &model.Oauth{ClientID: clientId, ClientSecret: ""},
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
		oauthStrategy := oauth{}

		for _, test := range testCases {
			// when
			result := oauthStrategy.CredentialsProvided(test.credentials)

			// then
			assert.Equal(t, test.result, result)
			assert.Equal(t, test.result, result)
		}
	})
}

func TestOauth_CreateSecretData(t *testing.T) {
	t.Run("should create secret data", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		// when
		secretData, err := oauthStrategy.CreateSecretData(oauthCredentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, []byte(clientId), secretData[OauthClientIDKey])
		assert.Equal(t, []byte(clientSecret), secretData[OauthClientSecretKey])
	})
}

func TestOauth_ToAppCredentials(t *testing.T) {
	t.Run("should convert to app basicCredentials", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		// when
		appCredentials := oauthStrategy.ToAppCredentials(oauthCredentials, secretName)

		// then
		assert.Equal(t, applications.CredentialsOAuthType, appCredentials.Type)
		assert.Equal(t, secretName, appCredentials.SecretName)
		assert.Equal(t, oauthUrl, appCredentials.AuthenticationUrl)
	})
}
