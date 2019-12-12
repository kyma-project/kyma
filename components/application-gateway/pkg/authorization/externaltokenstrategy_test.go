package authorization

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalAuthStrategy(t *testing.T) {

	t.Run("should use external token", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token", nil)

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		request.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = externalTokenStrategy.AddAuthorization(request, nil)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		externalTokenHeader := request.Header.Get(httpconsts.HeaderAccessToken)

		assert.Equal(t, "Bearer external", authHeader)
		assert.Equal(t, "", externalTokenHeader)
		oauthClientMock.AssertNotCalled(t, "GetToken")
	})

	t.Run("should use provided strategy when external token header is missing", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token", (*map[string][]string)(nil), (*map[string][]string)(nil)).Return("token", nil).Once()

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token", nil)

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = externalTokenStrategy.AddAuthorization(request, nil)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)

		assert.Equal(t, "Bearer token", authHeader)
		oauthClientMock.AssertExpectations(t)
	})

	t.Run("should call Invalidate method on the provided strategy", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("InvalidateTokenCache", "clientId").Return("token", nil).Once()

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token", nil)

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy)

		// when
		externalTokenStrategy.Invalidate()

		// then
		oauthClientMock.AssertExpectations(t)
	})
}
