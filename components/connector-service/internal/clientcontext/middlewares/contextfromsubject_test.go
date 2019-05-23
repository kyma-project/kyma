package middlewares

import (
	"fmt"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationContextFromSubjMiddleware_Middleware(t *testing.T) {
	fullSubject := "C=DE,ST=Waldorf,L=Waldorf,O=tenant,CN=test-app,OU=group"

	subjAppName := "test-app"
	subjGroup := "group"
	subjTenant := "tenant"

	testCases := []struct {
		certificateHeader string
		contextExtender   clientcontext.ContextExtender
		validationInfo    certificates.ValidationInfo
		isError           bool
	}{
		{
			certificateHeader: fullSubject,
			validationInfo:    certificates.ValidationInfo{"Organization", "OrgUnit", true},
			contextExtender:   clientcontext.ApplicationContext{Application: subjAppName, ClusterContext: clientcontext.ClusterContext{Tenant: subjTenant, Group: subjGroup}},
			isError:           false,
		},
		{
			certificateHeader: "CN=*Runtime*,C=DE,ST=Waldorf,L=Waldorf,O=tenant,OU=group",
			validationInfo:    certificates.ValidationInfo{Organization: "tenant", Unit: "group", Central: true},
			contextExtender:   clientcontext.ClusterContext{Tenant: subjTenant, Group: subjGroup},
			isError:           false,
		},
		{
			certificateHeader: "CN=test-app,C=DE,ST=Waldorf,L=Waldorf,O=tenant,OU=group",
			validationInfo:    certificates.ValidationInfo{Organization: "tenant", Unit: "group", Central: false},
			contextExtender:   clientcontext.ApplicationContext{Application: subjAppName, ClusterContext: clientcontext.ClusterContext{}},
			isError:           false,
		},
		{
			certificateHeader: "CN=test-app,C=DE,ST=Waldorf,L=Waldorf,O=,OU=",
			validationInfo:    certificates.ValidationInfo{Organization: "", Unit: "", Central: false},
			contextExtender:   clientcontext.ApplicationContext{Application: subjAppName, ClusterContext: clientcontext.ClusterContext{}},
			isError:           false,
		},
		{
			certificateHeader: "CN=*Runtime*,C=DE,ST=Waldorf,L=Waldorf,O=,OU=",
			validationInfo:    certificates.ValidationInfo{Organization: "", Unit: "", Central: true},
			contextExtender:   nil,
			isError:           true,
		},
		{
			certificateHeader: "CN=,C=DE,ST=Waldorf,L=Waldorf,O=tenant,OU=group",
			validationInfo:    certificates.ValidationInfo{Organization: "tenant", Unit: "group", Central: true},
			contextExtender:   nil,
			isError:           true,
		},
		{
			certificateHeader: "CN=,C=DE,ST=Waldorf,L=Waldorf,O=tenant,OU=group",
			validationInfo:    certificates.ValidationInfo{Organization: "tenant", Unit: "group", Central: false},
			contextExtender:   nil,
			isError:           true,
		},
	}

	t.Run("should parse context data from fullSubject", func(t *testing.T) {
		for _, test := range testCases {
			req := prepareRequestWithSubject(t, test.certificateHeader)

			ctxFromSubjMiddleware := NewContextFromSubjMiddleware(test.validationInfo)

			extender, err := ctxFromSubjMiddleware.parseContextFromSubject(req)
			if test.isError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.contextExtender, extender)
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

		middleware := NewContextFromSubjMiddleware(certificates.ValidationInfo{Organization: "", Unit: "", Central: true})

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

		middleware := NewContextFromSubjMiddleware(certificates.ValidationInfo{Organization: "Organization", Unit: "OrgUnit", Central: false})

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

		req := prepareRequestWithSubject(t, "C=DE,ST=Waldorf,L=Waldorf,O=tenant,CN=*Runtime*,OU=group")
		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(certificates.ValidationInfo{Organization: "tenant", Unit: "group", Central: true})

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

		middleware := NewContextFromSubjMiddleware(certificates.ValidationInfo{Organization: "Organization", Unit: "OrgUnit", Central: false})

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

		rr := httptest.NewRecorder()

		middleware := NewContextFromSubjMiddleware(certificates.ValidationInfo{Organization: "tenant", Unit: "group", Central: true})

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
