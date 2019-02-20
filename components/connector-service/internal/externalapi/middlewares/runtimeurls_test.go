package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeURLs_Middleware(t *testing.T) {

	url := "https://connector-service.kyma.local"

	baseGatewayHost := "gateway.kyma.local"

	t.Run("should set values in context when header values present", func(t *testing.T) {
		//given
		baseHeadersEventHost := "gateway.headers.events"
		baseHeadersMetadataHost := "gateway.headers.events"

		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(baseGatewayHost, clientcontext.CtxRequired)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.APIHostsKey).(clientcontext.APIHosts)
			assert.Equal(t, baseHeadersEventHost, ctxValue.EventsHost)
			assert.Equal(t, baseHeadersMetadataHost, ctxValue.MetadataHost)
			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		request.Header.Set(BaseEventsHostHeader, baseHeadersEventHost)
		request.Header.Set(BaseMetadataHostHeader, baseHeadersMetadataHost)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should pass empty strings from headers", func(t *testing.T) {
		//given
		baseHeadersEventHost := ""
		baseHeadersMetadataHost := ""

		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(baseGatewayHost, clientcontext.CtxRequired)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.APIHostsKey).(clientcontext.APIHosts)
			assert.Equal(t, baseHeadersEventHost, ctxValue.EventsHost)
			assert.Equal(t, baseHeadersMetadataHost, ctxValue.MetadataHost)
			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		request.Header.Set(BaseEventsHostHeader, baseHeadersEventHost)
		request.Header.Set(BaseMetadataHostHeader, baseHeadersMetadataHost)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should use default values when headers are not present and not required", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(baseGatewayHost, clientcontext.CtxNotRequired)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.APIHostsKey).(clientcontext.APIHosts)
			assert.Equal(t, baseGatewayHost, ctxValue.EventsHost)
			assert.Equal(t, baseGatewayHost, ctxValue.MetadataHost)
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

	t.Run("should fail when headers are not present but required", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(baseGatewayHost, clientcontext.CtxRequired)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
