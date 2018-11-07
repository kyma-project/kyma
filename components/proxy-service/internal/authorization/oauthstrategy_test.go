package authorization

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestAuthStrategy(t *testing.T) {

	t.Run("should add Authorization header", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("token", nil)

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = oauthStrategy.Setup(request)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "Bearer token", authHeader)
	})

	t.Run("should invalidate cache", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("InvalidateTokenCache", "clientId").Return("token", nil)

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		// when
		oauthStrategy.Reset()

		// then
		oauthClientMock.AssertExpectations(t)
	})

	t.Run("should not add Authorization header when getting token failed", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("", apperrors.Internal("failed")).Times(0)

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = oauthStrategy.Setup(request)

		// then
		require.Error(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		assert.Equal(t, "", authHeader)
		oauthClientMock.AssertExpectations(t)
	})

}
