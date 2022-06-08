package authorization

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"

	oauthMocks "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthWithCerStrategy(t *testing.T) {

	t.Run("should add Authorization header", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}

		oauthStrategy := newOAuthWithCertStrategy(oauthClientMock, "clientId", certificate, privateKey, "www.example.com/token", nil)

		prepareCertificate, err := oauthStrategy.prepareCertificate()
		require.NoError(t, err)

		oauthClientMock.On("GetTokenMTLS", "clientId", "www.example.com/token", prepareCertificate, (*map[string][]string)(nil), (*map[string][]string)(nil)).Return("token", nil)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = oauthStrategy.AddAuthorization(request, nil)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "Bearer token", authHeader)
	})

	t.Run("should invalidate cache", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}
		oauthClientMock.On("InvalidateTokenCache", "clientId", "www.example.com/token").Return("token", nil).Once()

		authWithCertStrategy := newOAuthWithCertStrategy(oauthClientMock, "clientId", certificate, privateKey, "www.example.com/token", nil)

		// when
		authWithCertStrategy.Invalidate()

		// then
		oauthClientMock.AssertExpectations(t)
	})

	t.Run("should not add Authorization header when getting token failed", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}

		authWithCertStrategy := newOAuthWithCertStrategy(oauthClientMock, "clientId", certificate, privateKey, "www.example.com/token", nil)

		prepareCertificate, err := authWithCertStrategy.prepareCertificate()
		require.NoError(t, err)

		oauthClientMock.On("GetTokenMTLS", "clientId", "www.example.com/token", prepareCertificate, (*map[string][]string)(nil), (*map[string][]string)(nil)).Return("", apperrors.Internal("failed")).Once()

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = authWithCertStrategy.AddAuthorization(request, nil)

		// then
		require.Error(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "", authHeader)
		oauthClientMock.AssertExpectations(t)
	})

}
