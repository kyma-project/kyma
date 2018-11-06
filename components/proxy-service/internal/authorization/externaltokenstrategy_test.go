package authorization

import (
	"github.com/stretchr/testify/require"
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/mocks"
)

func TestExternalAuthStrategy(t *testing.T) {

	t.Run("should use external token", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token")

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		basicAuthStrategy := newExternalTokenStrategy(oauthStrategy)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		request.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = basicAuthStrategy.Setup(request)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		externalTokenHeader := request.Header.Get(httpconsts.HeaderAccessToken)

		assert.Equal(t, "Bearer external", authHeader)
		assert.Equal(t, "", externalTokenHeader)
		oauthClientMock.AssertNotCalled(t, "GetToken")
	})

	t.Run("should use provider strategy in external token header is missing", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("token", nil)

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		basicAuthStrategy := newExternalTokenStrategy(oauthStrategy)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = basicAuthStrategy.Setup(request)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)

		assert.Equal(t, "Bearer token", authHeader)
		oauthClientMock.AssertExpectations(t)
	})


	t.Run("should call Reset method on the provided strategy", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("InvalidateTokenCache", "clientId").Return("token", nil)

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		basicAuthStrategy := newExternalTokenStrategy(oauthStrategy)

		// when
		basicAuthStrategy.Reset()

		// then
		oauthClientMock.AssertExpectations(t)
	})
}
