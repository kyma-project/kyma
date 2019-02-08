package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTenant = "tenant"
	defaultGroup  = "group"
	testTenant    = "test-tenant"
	testGroup     = "test-group"
)

func TestClusterContextMiddleware_Middleware(t *testing.T) {

	t.Run("should use default values", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ClusterContextKey).(clientcontext.ClusterContext)
			require.True(t, ok)

			assert.Equal(t, defaultTenant, clusterCtx.Tenant)
			assert.Equal(t, defaultGroup, clusterCtx.Group)
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewClusterContextMiddleware(defaultTenant, defaultGroup)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should use header values if no defaults", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ClusterContextKey).(clientcontext.ClusterContext)
			require.True(t, ok)

			assert.Equal(t, testTenant, clusterCtx.Tenant)
			assert.Equal(t, testGroup, clusterCtx.Group)
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.TenantHeader, testTenant)
		req.Header.Set(clientcontext.GroupHeader, testGroup)

		rr := httptest.NewRecorder()

		middleware := NewClusterContextMiddleware("", "")

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 if no default values and no header provided", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewClusterContextMiddleware("", "")

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
