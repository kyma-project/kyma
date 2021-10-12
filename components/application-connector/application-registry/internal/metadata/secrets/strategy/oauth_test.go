package strategy

import (
	"encoding/json"
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
	oauthCredentials = &model.CredentialsWithCSRF{
		Oauth: &model.Oauth{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			URL:          oauthUrl,
		},
	}
	headers = map[string][]string{
		"headerKey": {"headerValue"},
	}
	queryParameters = map[string][]string{
		"queryParameterKey": {"queryParameterValue"},
	}
	requestParameters = model.RequestParameters{
		Headers:         &headers,
		QueryParameters: &queryParameters,
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

	t.Run("should convert to credentials", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		// when
		credentials, err := oauthStrategy.ToCredentials(secretData, appCredentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, credentials.Oauth.ClientID)
		assert.Equal(t, clientSecret, credentials.Oauth.ClientSecret)
		assert.Equal(t, oauthUrl, credentials.Oauth.URL)
	})

	t.Run("should convert to credentials with additional headers and query parameters", func(t *testing.T) {
		// given
		headers, _ := json.Marshal(requestParameters.Headers)
		queryParameters, _ := json.Marshal(requestParameters.QueryParameters)
		secretData := map[string][]byte{
			OauthClientIDKey:     []byte(clientId),
			OauthClientSecretKey: []byte(clientSecret),
			HeadersKey:           headers,
			QueryParametersKey:   queryParameters,
		}

		oauthStrategy := oauth{}

		// when
		credentials, err := oauthStrategy.ToCredentials(secretData, appCredentials)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, credentials.Oauth.ClientID)
		assert.Equal(t, clientSecret, credentials.Oauth.ClientSecret)
		assert.Equal(t, oauthUrl, credentials.Oauth.URL)
		assert.Equal(t, requestParameters, *credentials.Oauth.RequestParameters)
	})

	t.Run("should convert to credentials with CSRF", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		c := &applications.Credentials{
			AuthenticationUrl: oauthUrl,
			CSRFInfo:          &applications.CSRFInfo{TokenEndpointURL: "https://test.it"},
		}
		// when
		credentials, err := oauthStrategy.ToCredentials(secretData, c)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, credentials.Oauth.ClientID)
		assert.Equal(t, clientSecret, credentials.Oauth.ClientSecret)
		assert.Equal(t, oauthUrl, credentials.Oauth.URL)
		assert.Equal(t, "https://test.it", credentials.CSRFInfo.TokenEndpointURL)
	})
}

func TestOauth_CredentialsProvided(t *testing.T) {

	testCases := []struct {
		credentials *model.CredentialsWithCSRF
		result      bool
	}{
		{
			credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					ClientID:     clientId,
					ClientSecret: clientSecret,
				},
			},
			result: true,
		},
		{
			credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					ClientID:     "",
					ClientSecret: clientSecret,
				},
			},
			result: false,
		},
		{
			credentials: &model.CredentialsWithCSRF{
				Oauth: &model.Oauth{
					ClientID:     clientId,
					ClientSecret: "",
				},
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

	t.Run("should create secret data with additional headers and query parameters", func(t *testing.T) {
		// given
		oauthCredentials = &model.CredentialsWithCSRF{
			Oauth: &model.Oauth{
				ClientID:          clientId,
				ClientSecret:      clientSecret,
				URL:               oauthUrl,
				RequestParameters: &requestParameters,
			},
		}

		oauthStrategy := oauth{}

		// when
		secretData, err := oauthStrategy.CreateSecretData(oauthCredentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, []byte(clientId), secretData[OauthClientIDKey])
		assert.Equal(t, []byte(clientSecret), secretData[OauthClientSecretKey])

		headersMarshalled, _ := json.Marshal(requestParameters.Headers)
		assert.Equal(t, headersMarshalled, secretData[HeadersKey])
		queryParamsMarshalled, _ := json.Marshal(requestParameters.QueryParameters)
		assert.Equal(t, queryParamsMarshalled, secretData[QueryParametersKey])
	})
}

func TestOauth_ToCredentialsInfo(t *testing.T) {
	t.Run("should convert to app credentials", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		// when
		appCredentials := oauthStrategy.ToCredentialsInfo(oauthCredentials, secretName)

		// then
		assert.Equal(t, applications.CredentialsOAuthType, appCredentials.Type)
		assert.Equal(t, secretName, appCredentials.SecretName)
		assert.Equal(t, oauthUrl, appCredentials.AuthenticationUrl)
	})
}

func TestOauth_ShouldUpdate(t *testing.T) {
	testCases := []struct {
		currentData SecretData
		newData     SecretData
		result      bool
	}{
		{
			currentData: SecretData{
				OauthClientIDKey:     []byte(clientId),
				OauthClientSecretKey: []byte(clientSecret),
			},
			newData: SecretData{
				OauthClientIDKey:     []byte("changed client id"),
				OauthClientSecretKey: []byte(clientSecret),
			},
			result: true,
		},
		{
			currentData: SecretData{
				OauthClientIDKey:     []byte(clientId),
				OauthClientSecretKey: []byte(clientSecret),
			},
			newData: SecretData{
				OauthClientIDKey:     []byte(username),
				OauthClientSecretKey: []byte("changed secret"),
			},
			result: true,
		},
		{
			currentData: SecretData{},
			newData: SecretData{
				OauthClientIDKey:     []byte(clientId),
				OauthClientSecretKey: []byte(clientSecret),
			},
			result: true,
		},
		{
			currentData: SecretData{
				OauthClientIDKey:     []byte(clientId),
				OauthClientSecretKey: []byte(clientSecret),
			},
			newData: SecretData{},
			result:  true,
		},
		{
			currentData: SecretData{
				OauthClientIDKey:     []byte(clientId),
				OauthClientSecretKey: []byte(clientSecret),
			},
			newData: SecretData{
				OauthClientIDKey:     []byte(clientId),
				OauthClientSecretKey: []byte(clientSecret),
			},
			result: false,
		},
	}

	t.Run("should return true when update needed", func(t *testing.T) {
		// given
		oauthStrategy := oauth{}

		for _, test := range testCases {
			// when
			result := oauthStrategy.ShouldUpdate(test.currentData, test.newData)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}
