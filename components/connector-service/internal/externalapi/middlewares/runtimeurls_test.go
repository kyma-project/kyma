package middlewares

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuntimeURLs_Middleware(t *testing.T) {

	url := "https://connector-service.kyma.local"

	defaultEventsHost := "https://gateway.kyma.local.events"
	defaultMetadataHost := "https://gateway.kyma.local.metadata"

	t.Run("should set headers values when present", func(t *testing.T) {
		//given
		baseEventHost := "https://gateway.headers.events"
		baseMetadataHost := "https://gateway.headers.metadata"

		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(defaultMetadataHost, defaultEventsHost)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.APIHostsKey).(clientcontext.APIHosts)
			assert.Equal(t, baseEventHost, ctxValue.EventsHost)
			assert.Equal(t, baseMetadataHost, ctxValue.MetadataHost)
			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		request.Header.Set(BaseEventsHostHeader, baseEventHost)
		request.Header.Set(BaseMetadataHostHeader, baseMetadataHost)

		rr := httptest.NewRecorder()

		//when
		resultHandler := runtimeURLsMiddleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, request)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should use default values when headers are not present", func(t *testing.T) {
		//given
		runtimeURLsMiddleware := NewRuntimeURLsMiddleware(defaultMetadataHost, defaultEventsHost)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctxValue := ctx.Value(clientcontext.APIHostsKey).(clientcontext.APIHosts)
			assert.Equal(t, defaultEventsHost, ctxValue.EventsHost)
			assert.Equal(t, defaultMetadataHost, ctxValue.MetadataHost)
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
}
