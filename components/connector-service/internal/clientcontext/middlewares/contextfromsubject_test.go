package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"

	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationContextFromSubjMiddleware_Middleware(t *testing.T) {
	fullSubject := certificates.CertInfo{Hash: "", Subject: "C=DE,ST=Waldorf,L=Waldorf,O=tenant,CN=test-app,OU=group"}

	subjAppName := "test-app"
	subjGroup := "group"
	subjTenant := "tenant"

	hp := &mocks.HeaderParser{}

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

		req := prepareRequestWithSubject(t, fullSubject.Subject)

		hp.On("ParseCertificateHeader", *req).Return(fullSubject, nil)

		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(hp, true)

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

		certInfo := certificates.CertInfo{Hash: "", Subject: "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=test-app,OU=OrgUnit"}

		req := prepareRequestWithSubject(t, certInfo.Subject)

		hp.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(hp, false)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should create ClusterContext when CN equal *Runtime*", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clusterCtx, ok := ctx.Value(clientcontext.ClusterContextKey).(clientcontext.ClusterContext)
			require.True(t, ok)

			assert.Equal(t, subjGroup, clusterCtx.Group)
			assert.Equal(t, subjTenant, clusterCtx.Tenant)
			w.WriteHeader(http.StatusOK)
		})

		certInfo := certificates.CertInfo{Hash: "", Subject: "C=DE,ST=Waldorf,L=Waldorf,O=tenant,CN=*Runtime*,OU=group"}

		req := prepareRequestWithSubject(t, certInfo.Subject)

		hp.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(hp, true)

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

		certInfo := certificates.CertInfo{Hash: "", Subject: "C=DE,ST=Waldorf,L=Waldorf,O=Organization,CN=,OU=OrgUnit"}

		req := prepareRequestWithSubject(t, certInfo.Subject)

		hp.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(hp, false)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should return 400 when no Subject in certificate header is passed", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("GET", "/", nil)

		require.NoError(t, err)

		hp.On("ParseCertificateHeader", *req).Return(certificates.CertInfo{}, nil)

		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(hp, true)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func prepareRequestWithSubject(t *testing.T, subject string) *http.Request {
	certificateHeader := fmt.Sprintf("Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"%s\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default", subject)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	req.Header.Set(certificates.ClientCertHeader, certificateHeader)

	return req
}
