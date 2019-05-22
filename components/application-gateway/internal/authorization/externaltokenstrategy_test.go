package authorization

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization/oauth/mocks"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalAuthStrategy(t *testing.T) {

	t.Run("should use external token", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy, nil, nil)

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

	t.Run("should inject custom headers", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		headers := map[string][]string{
			"X-Custom1": []string{"custom-value"},
		}

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy, &headers, nil)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		request.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = externalTokenStrategy.AddAuthorization(request, nil)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		externalTokenHeader := request.Header.Get(httpconsts.HeaderAccessToken)
		customHeader := request.Header.Get("X-Custom1")

		assert.Equal(t, "Bearer external", authHeader)
		assert.Equal(t, "", externalTokenHeader)
		assert.Equal(t, "custom-value", customHeader)
		oauthClientMock.AssertNotCalled(t, "GetToken")
	})

	t.Run("should inject custom query parameters", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		queryParams := map[string][]string{
			"param1": []string{"custom-value"},
		}

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy, nil, &queryParams)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		request.Header.Set(httpconsts.HeaderAccessToken, "Bearer external")

		// when
		err = externalTokenStrategy.AddAuthorization(request, nil)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
		externalTokenHeader := request.Header.Get(httpconsts.HeaderAccessToken)
		customQueryParam := request.URL.Query().Get("param1")

		assert.Equal(t, "Bearer external", authHeader)
		assert.Equal(t, "", externalTokenHeader)
		assert.Equal(t, "custom-value", customQueryParam)
		oauthClientMock.AssertNotCalled(t, "GetToken")
	})

	t.Run("should use provided strategy when external token header is missing", func(t *testing.T) {
		// given
		oauthClientMock := &mocks.Client{}
		oauthClientMock.On("GetToken", "clientId", "clientSecret", "www.example.com/token").Return("token", nil).Once()

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy, nil, nil)

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

		oauthStrategy := newOAuthStrategy(oauthClientMock, "clientId", "clientSecret", "www.example.com/token")

		externalTokenStrategy := newExternalTokenStrategy(oauthStrategy, nil, nil)

		// when
		externalTokenStrategy.Invalidate()

		// then
		oauthClientMock.AssertExpectations(t)
	})
}
