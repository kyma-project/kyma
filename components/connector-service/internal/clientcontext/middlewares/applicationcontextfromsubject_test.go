package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	subject = "/C=DE/ST=Waldorf/L=Waldorf/O=Organization/OU=OrgUnit/CN=test-app"
)

func TestApplicationContextFromSubjMiddleware_Middleware(t *testing.T) {
	expectedCName := "test-app"

	t.Run("should extract Common Name", func(t *testing.T) {
		//given
		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.SubjectHeader, subject)

		//when
		cname := extractApplicationFromSubject(req)

		//then
		assert.Equal(t, expectedCName, cname)
	})

	t.Run("should set context based on header", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ApplicationContextKey).(clientcontext.ApplicationContext)
			require.True(t, ok)

			assert.Equal(t, expectedCName, clusterCtx.Application)
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)
		req.Header.Set(clientcontext.SubjectHeader, subject)

		rr := httptest.NewRecorder()

		middleware := NewAppContextFromSubjMiddleware()

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 when no Subject header is passed", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		middleware := NewAppContextFromSubjMiddleware()

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
