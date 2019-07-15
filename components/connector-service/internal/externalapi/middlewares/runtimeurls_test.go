package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi/middlewares/lookup/mocks"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeURLs_Middleware(t *testing.T) {

	url := "https://connector-service.kyma.local"

	fetchedGatewayBaseURL := "https://gateway.host"

	defaultGatewayBaseURL := "https://gateway.kyma.local"

	lookupService := &mocks.LookupService{}
	extractor := clientcontext.ExtractClientContext

	t.Run("should set fetched gateway URL value in context when lookup is enabled", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(defaultGatewayBaseURL, clientcontext.LookupEnabled, extractor, lookupService)

		appCtx := clientcontext.ClientContext{
			Group:  "testGroup",
			Tenant: "testTenant",
			ID:     "testApp",
		}

		lookupService.On("Fetch", appCtx).Return(fetchedGatewayBaseURL, nil)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.ApiURLsKey).(clientcontext.ApiURLs)
			assert.Equal(t, fetchedGatewayBaseURL, ctxValue.EventsBaseURL)
			assert.Equal(t, fetchedGatewayBaseURL, ctxValue.MetadataBaseURL)
			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		request = request.WithContext(appCtx.ExtendContext(request.Context()))

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should use default gateway value when lookup is disabled", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(defaultGatewayBaseURL, clientcontext.LookupDisabled, extractor, lookupService)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.ApiURLsKey).(clientcontext.ApiURLs)
			assert.Equal(t, defaultGatewayBaseURL, ctxValue.EventsBaseURL)
			assert.Equal(t, defaultGatewayBaseURL, ctxValue.MetadataBaseURL)
			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return code 500 when cannot read ApplicationContext", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(defaultGatewayBaseURL, clientcontext.LookupEnabled, extractor, lookupService)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return code 500 when gateway URL fetch failed", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(defaultGatewayBaseURL, clientcontext.LookupEnabled, extractor, lookupService)

		appCtx := clientcontext.ClientContext{
			Group:  "testGroup",
			Tenant: "testTenant",
			ID:     "testApp",
		}

		lookupService.On("Fetch", appCtx).Return(nil, apperrors.Internal("some error"))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

}
