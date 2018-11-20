package proxy

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	authMock "github.com/kyma-project/kyma/components/proxy-service/internal/authorization/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization"
)

func TestForbiddenResponseRetrier_CheckResponse(t *testing.T) {
	t.Run("should return response if code is different than 401 and 403", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, nil
		}
		rr := newUnathorizedResponseRetrier("", &http.Request{}, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: 500}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t,http.StatusInternalServerError, response.StatusCode)
	})

	t.Run("should retry if 401 occurred", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil)
		authStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnathorizedResponseRetrier("id1", req, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: 401}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t,http.StatusOK, response.StatusCode)
	})

	t.Run("should retry if 403 occurred", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil)
		authStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnathorizedResponseRetrier("id1", req, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusForbidden}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t,http.StatusOK, response.StatusCode)
	})

	t.Run("should not retry if 401 occurred and flag is already set", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, nil
		}
		rr := newUnathorizedResponseRetrier("", &http.Request{}, 10, updateCacheEntryFunction)
		rr.retried = true
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t,http.StatusUnauthorized, response.StatusCode)
	})

	t.Run("should not retry if 403 occurred and flag is already set", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, nil
		}
		rr := newUnathorizedResponseRetrier("", &http.Request{}, 10, updateCacheEntryFunction)
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
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil)
		authStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnathorizedResponseRetrier("id1", req, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t,http.StatusForbidden, response.StatusCode)
	})

	t.Run("should retry only once if 403 occurred", func(t *testing.T) {
		// given
		ts := NewTestServerForRetryTest(http.StatusUnauthorized, func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil)
		authStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, ts.URL, authStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnathorizedResponseRetrier("id1", req, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusForbidden}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
		assert.Equal(t,http.StatusUnauthorized, response.StatusCode)
	})

	t.Run("should return error if failed to update entry cache", func(t *testing.T) {
		// given
		updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
			return nil, apperrors.Internal("failed")
		}

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnathorizedResponseRetrier("id1", req, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.Error(t, err)
		assert.Equal(t,http.StatusUnauthorized, response.StatusCode)
	})

	t.Run("should return error if failed to add authorization header", func(t *testing.T) {
		// given
		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(apperrors.Internal("failed"))
		authStrategyMock.On("Invalidate").Return()

		updateCacheEntryFunction := newUpdateCacheEntryFunction(t, "", authStrategyMock)

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		rr := newUnathorizedResponseRetrier("id1", req, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: http.StatusUnauthorized}

		// when
		err = rr.RetryIfFailedToAuthorize(response)

		// then
		assert.Error(t, err)
		assert.Equal(t,http.StatusUnauthorized, response.StatusCode)
	})
}

func newUpdateCacheEntryFunction(t *testing.T, url string, strategy authorization.Strategy) func(id string) (*CacheEntry, apperrors.AppError){
	return func(id string) (*CacheEntry, apperrors.AppError) {
		assert.Equal(t, "id1", id)

		proxy, err := makeProxy(url, "id1", true)
		require.NoError(t, err)

		return &CacheEntry{
			Proxy: proxy,
			AuthorizationStrategy: strategy,
		}, nil
	}
}

