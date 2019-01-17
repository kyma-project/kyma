package authorization

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
	oauthMocks "github.com/kyma-project/kyma/components/application-proxy/internal/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/application-proxy/internal/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthStrategy(t *testing.T) {

	t.Run("should add Authorization header", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("token", nil)

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = oauthStrategy.AddAuthorization(request, proxyStub)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "Bearer token", authHeader)
	})

	t.Run("should invalidate cache", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}
		oauthClientMock.On("InvalidateTokenCache", "clientId").Return("token", nil).Once()

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		// when
		oauthStrategy.Invalidate()

		// then
		oauthClientMock.AssertExpectations(t)
	})

	t.Run("should not add Authorization header when getting token failed", func(t *testing.T) {
		// given
		oauthClientMock := &oauthMocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("", apperrors.Internal("failed")).Once()

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = oauthStrategy.AddAuthorization(request, proxyStub)

		// then
		require.Error(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "", authHeader)
		oauthClientMock.AssertExpectations(t)
	})

}
