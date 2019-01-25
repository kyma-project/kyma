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
)

func TestApplicationContextMiddleware_Middleware(t *testing.T) {

	emptyClusterContextMiddleware := &clusterContextMiddleware{}

	t.Run("should set context based on header", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ApplicationContextKey).(clientcontext.ApplicationContext)
			require.True(t, ok)

			assert.Equal(t, testApplication, clusterCtx.Application)
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.ApplicationHeader, testApplication)
		req.Header.Set(clientcontext.TenantHeader, testTenant)
		req.Header.Set(clientcontext.GroupHeader, testGroup)

		rr := httptest.NewRecorder()

		middleware := NewApplicationContextMiddleware(emptyClusterContextMiddleware)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 if no application header provided", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.TenantHeader, testTenant)
		req.Header.Set(clientcontext.GroupHeader, testGroup)

		rr := httptest.NewRecorder()

		middleware := NewApplicationContextMiddleware(emptyClusterContextMiddleware)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should set context based on header and cluster middleware", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ApplicationContextKey).(clientcontext.ApplicationContext)
			require.True(t, ok)

			assert.Equal(t, testApplication, clusterCtx.Application)
			w.WriteHeader(http.StatusOK)
		})

		emptyClusterContextMiddleware := &clusterContextMiddleware{
			defaultTenant: defaultTenant,
			defaultGroup:  defaultGroup,
		}

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.ApplicationHeader, testApplication)

		rr := httptest.NewRecorder()

		middleware := NewApplicationContextMiddleware(emptyClusterContextMiddleware)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
