package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationContextFromSubjMiddleware_Middleware(t *testing.T) {
	fullSubject := "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=tenant\\;group\\;test-app,OU=OrgUnit"

	subjAppName := "test-app"
	subjGroup := "group"
	subjTenant := "tenant"

	testCases := []struct {
		subject string
		app     string
		group   string
		tenant  string
	}{
		{subject: fullSubject, app: subjAppName, group: subjGroup, tenant: subjTenant},
		{subject: "CN=tenant\\;group\\;test-app,C=DE,ST=Waldorf,L=Waldorf,O=Organization,OU=OrgUnit", app: subjAppName, group: subjGroup, tenant: subjTenant},
		{subject: "C=DE,ST=Waldorf,L=Waldorf,O=Organization,OU=OrgUnit,CN=tenant\\;group\\;test-app", app: subjAppName, group: subjGroup, tenant: subjTenant},
		{subject: "not matching", app: "", group: "", tenant: ""},
		{subject: "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=test-app,OU=OrgUnit", app: subjAppName, group: "", tenant: ""},
		{subject: "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=tenant\\;group,OU=OrgUnit", app: "", group: subjGroup, tenant: subjTenant},
	}

	t.Run("should parse context data from fullSubject", func(t *testing.T) {
		for _, test := range testCases {
			req := prepareRequestWithSubject(t, test.subject)

			app, group, tenant := parseContextFromSubject(req)

			assert.Equal(t, test.app, app)
			assert.Equal(t, test.group, group)
			assert.Equal(t, test.tenant, tenant)
		}
	})

	t.Run("should create ApplicationContext", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			appCtx, ok := ctx.Value(clientcontext.ApplicationContextKey).(clientcontext.ApplicationContext)
			require.True(t, ok)

			assert.Equal(t, subjAppName, appCtx.Application)
			assert.Equal(t, subjGroup, appCtx.ClusterContext.Group)
			assert.Equal(t, subjTenant, appCtx.ClusterContext.Tenant)
			w.WriteHeader(http.StatusOK)
		})

		req := prepareRequestWithSubject(t, fullSubject)

		rr := httptest.NewRecorder()

		middleware := NewAppContextFromSubjMiddleware()

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should create ApplicationContext with empty ClusterContext if CN only with app", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			appCtx, ok := ctx.Value(clientcontext.ApplicationContextKey).(clientcontext.ApplicationContext)
			require.True(t, ok)

			assert.Equal(t, subjAppName, appCtx.Application)
			assert.Empty(t, appCtx.ClusterContext)
			w.WriteHeader(http.StatusOK)
		})

		req := prepareRequestWithSubject(t, "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=test-app,OU=OrgUnit")
		rr := httptest.NewRecorder()

		middleware := NewAppContextFromSubjMiddleware()

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should create ClusterContext when no app name in CN", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ClusterContextKey).(clientcontext.ClusterContext)
			require.True(t, ok)

			assert.Equal(t, subjGroup, clusterCtx.Group)
			assert.Equal(t, subjTenant, clusterCtx.Tenant)
			w.WriteHeader(http.StatusOK)
		})

		req := prepareRequestWithSubject(t, "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=tenant\\;group,OU=OrgUnit")
		rr := httptest.NewRecorder()

		middleware := NewAppContextFromSubjMiddleware()

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 when Common Name is empty", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := prepareRequestWithSubject(t, "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=,OU=OrgUnit")
		rr := httptest.NewRecorder()

		middleware := NewAppContextFromSubjMiddleware()

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
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

func prepareRequestWithSubject(t *testing.T, subject string) *http.Request {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.Header.Set(clientcontext.SubjectHeader, subject)

	return req
}
