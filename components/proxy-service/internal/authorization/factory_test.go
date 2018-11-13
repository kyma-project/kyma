package authorization

import (
	oauthMocks "github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	metadatamodel "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestStrategyFactory(t *testing.T) {
	t.Run("should create no auth strategy", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}

		factory := authorizationStrategyFactory{oauthClient: oauthClientMock}

		// when
		strategy := factory.Create(nil)

		// then
		require.NotNil(t, strategy)

		// given
		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = strategy.AddAuthorizationHeader(request)

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "", authHeader)

		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorizationHeader(requestWithExternalToken)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})

	t.Run("should create basic auth strategy", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}

		factory := authorizationStrategyFactory{oauthClient: oauthClientMock}
		credentials := &metadatamodel.Credentials{
			BasicAuth: &metadatamodel.BasicAuth{
				Username: "username",
				Password: "password",
			},
		}

		// when
		strategy := factory.Create(credentials)

		// then
		require.NotNil(t, strategy)

		// given
		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = strategy.AddAuthorizationHeader(request)

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Contains(t, authHeader, "Basic ")

		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorizationHeader(requestWithExternalToken)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})

	t.Run("should create oauth strategy", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("token", nil)

		factory := authorizationStrategyFactory{oauthClient: oauthClientMock}
		credentials := &metadatamodel.Credentials{
			OAuth: &metadatamodel.OAuth{
				ClientID:     "clientId",
				ClientSecret: "clientSecret",
				URL:          "www.example.com/token",
			},
		}

		// when
		strategy := factory.Create(credentials)

		// then
		require.NotNil(t, strategy)

		// given
		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = strategy.AddAuthorizationHeader(request)

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, authHeader, "Bearer token")

		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorizationHeader(requestWithExternalToken)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})
}
