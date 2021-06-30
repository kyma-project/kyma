package authorization

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"

	oauthMocks "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		err = strategy.AddAuthorization(request, nil)

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "", authHeader)

		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorization(requestWithExternalToken, nil)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})

	t.Run("should create basic auth strategy", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}

		factory := authorizationStrategyFactory{oauthClient: oauthClientMock}
		credentials := &Credentials{
			BasicAuth: &BasicAuth{
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
		err = strategy.AddAuthorization(request, nil)

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Contains(t, authHeader, "Basic ")

		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorization(requestWithExternalToken, nil)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})

	t.Run("should create oauth strategy", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token", (*map[string][]string)(nil), (*map[string][]string)(nil)).Return("token", nil)

		factory := authorizationStrategyFactory{oauthClient: oauthClientMock}
		credentials := &Credentials{
			OAuth: &OAuth{
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
		err = strategy.AddAuthorization(request, nil)

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, authHeader, "Bearer token")

		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorization(requestWithExternalToken, nil)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})

	t.Run("should create certificate gen strategy", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}

		factory := authorizationStrategyFactory{oauthClient: oauthClientMock}
		credentials := &Credentials{
			CertificateGen: &CertificateGen{
				Certificate: certificate,
				PrivateKey:  privateKey,
			},
		}
		clientCert := clientcert.NewClientCertificate(nil)

		// when
		strategy := factory.Create(credentials)

		// then
		require.NotNil(t, strategy)

		// given
		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = strategy.AddAuthorization(request, func(cert *tls.Certificate) {
			clientCert.SetCertificate(cert)
		})

		// then
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		expectedCert, err := tls.X509KeyPair(credentials.CertificateGen.Certificate, credentials.CertificateGen.PrivateKey)
		assert.Nil(t, err)
		assert.Equal(t, expectedCert, *clientCert.GetCertificate())
		// given
		requestWithExternalToken, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		requestWithExternalToken.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = strategy.AddAuthorization(requestWithExternalToken, nil)

		// then
		authHeader = requestWithExternalToken.Header.Get(httpconsts.HeaderAuthorization)
		assert.Nil(t, err)
		assert.Equal(t, "Bearer external", authHeader)
	})
}
