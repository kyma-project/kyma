package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testApplication = "test-app"
	testTenant      = "test-tenant"
	testGroup       = "test-group"
)

func TestApplicationContextMiddleware_Middleware(t *testing.T) {

	t.Run("should set context based on header", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			applicationCtx, ok := ctx.Value(clientcontext.ClientContextKey).(clientcontext.ClientContext)
			require.True(t, ok)

			assert.Equal(t, testApplication, applicationCtx.ID)
			assert.Equal(t, testTenant, applicationCtx.Tenant)
			assert.Equal(t, testGroup, applicationCtx.Group)

			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.ApplicationHeader, testApplication)
		req.Header.Set(clientcontext.TenantHeader, testTenant)
		req.Header.Set(clientcontext.GroupHeader, testGroup)

		rr := httptest.NewRecorder()

		clusterContextStrategy := clientcontext.NewClusterContextStrategy(true)

		middleware := NewContextMiddleware(clusterContextStrategy, clientcontext.ApplicationHeader)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should use empty cluster context if disabled strategy", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			applicationCtx, ok := ctx.Value(clientcontext.ClientContextKey).(clientcontext.ClientContext)
			require.True(t, ok)

			assert.Equal(t, testApplication, applicationCtx.ID)
			assert.Empty(t, applicationCtx.Tenant)
			assert.Empty(t, applicationCtx.Group)

			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.ApplicationHeader, testApplication)
		req.Header.Set(clientcontext.TenantHeader, testTenant)
		req.Header.Set(clientcontext.GroupHeader, testGroup)

		rr := httptest.NewRecorder()

		clusterContextStrategy := clientcontext.NewClusterContextStrategy(false)

		middleware := NewContextMiddleware(clusterContextStrategy, clientcontext.ApplicationHeader)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 if no application header provided and ctx is enabled", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.TenantHeader, testTenant)
		req.Header.Set(clientcontext.GroupHeader, testGroup)

		rr := httptest.NewRecorder()

		clusterContextStrategy := clientcontext.NewClusterContextStrategy(true)

		middleware := NewContextMiddleware(clusterContextStrategy, clientcontext.ApplicationHeader)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should return 400 if no application header provided and ctx is disabled", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		clusterContextStrategy := clientcontext.NewClusterContextStrategy(false)

		middleware := NewContextMiddleware(clusterContextStrategy, clientcontext.RuntimeIDHeader)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

}
