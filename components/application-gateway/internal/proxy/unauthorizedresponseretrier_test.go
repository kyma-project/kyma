package proxy

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	csrfMock "github.com/kyma-project/kyma/components/application-gateway/internal/csrf/mocks"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	authMock "github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestForbiddenResponseRetrier_CheckResponse(t *testing.T) {
	t.Run("should return response if code is different than 401 and 403", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, nil
		}
		rr := newUnauthorizedResponseRetrier("", &http.Request{}, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: 500}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	})

	t.Run("should retry if 401 occurred", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil)
		authStrategyMock.On("Invalidate").Return()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).
			Return(nil)
		csrfTokenStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock, csrfTokenStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnauthorizedResponseRetrier("id1", req, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: 401}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("should retry if 403 occurred", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")

			assert.Equal(t, "token", req.Header.Get("CSRFToken"))

			cookie, err := req.Cookie("CSRFToken")
			require.NoError(t, err)
			assert.Equal(t, "token", cookie.Value)
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil)
		authStrategyMock.On("Invalidate").Return()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).
			Run(func(args mock.Arguments) {
				req := args[0].(*http.Request)
				req.Header.Set("CSRFToken", "token")
				req.AddCookie(&http.Cookie{
					Name:  "CSRFToken",
					Value: "token",
				})
			}).
			Return(nil)
		csrfTokenStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock, csrfTokenStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnauthorizedResponseRetrier("id1", req, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusForbidden}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("should not retry if 401 occurred and flag is already set", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, nil
		}
		rr := newUnauthorizedResponseRetrier("", &http.Request{}, nil, 10, updateCacheEntryFunction)
		rr.retried = true
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})

	t.Run("should not retry if 403 occurred and flag is already set", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, nil
		}
		rr := newUnauthorizedResponseRetrier("", &http.Request{}, nil, 10, updateCacheEntryFunction)
		rr.retried = true
		response := &http.Response{StatusCode: http.StatusForbidden}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, response.StatusCode)
	})

	t.Run("should retry only once if 401 occurred", func(t *testing.T) {
		// given
		ts := NewTestServerForRetryTest(http.StatusForbidden, func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil)
		authStrategyMock.On("Invalidate").Return()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).
			Return(nil)
		csrfTokenStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock, csrfTokenStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnauthorizedResponseRetrier("id1", req, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, response.StatusCode)
	})

	t.Run("should retry only once if 403 occurred", func(t *testing.T) {
		// given
		ts := NewTestServerForRetryTest(http.StatusUnauthorized, func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil)
		authStrategyMock.On("Invalidate").Return()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).
			Return(nil)
		csrfTokenStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock, csrfTokenStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnauthorizedResponseRetrier("id1", req, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusForbidden}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})

	t.Run("should return error if failed to update entry cache", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, apperrors.Internal("failed")
		}

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnauthorizedResponseRetrier("id1", req, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})

	t.Run("should return error if failed to add authorization header", func(t *testing.T) {
		// given
		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(apperrors.Internal("failed"))
		authStrategyMock.On("Invalidate").Return()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).
			Return(nil)
		csrfTokenStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, "", authStrategyMock, csrfTokenStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnauthorizedResponseRetrier("id1", req, nil, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})
}

func newUpdateCacheEntryFunction(t *testing.T, url string, strategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy) func(id string) (*CacheEntry, apperrors.AppError) {
	return func(id string) (*CacheEntry, apperrors.AppError) {
		assert.Equal(t, "id1", id)

		proxy, err := makeProxy(url, nil, "id1", true)
		require.NoError(t, err)

		return &CacheEntry{
			Proxy:                 proxy,
			AuthorizationStrategy: &authorizationStrategyWrapper{strategy, proxy},
			CSRFTokenStrategy:     csrfTokenStrategy,
		}, nil
	}
}
