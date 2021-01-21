package tracing_test

import (
	"github.com/kyma-project/kyma/common/logger/tracing"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	t.Run("with trace and span in header, should put traceid and spanid to context", func(t *testing.T) {
		//GIVEN
		var outRequest *http.Request
		middleware := tracing.NewTracingMiddleware(func(w http.ResponseWriter, r *http.Request) {
			outRequest = r
		})
		resp := httptest.NewRecorder()

		r, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)
		r.Header[tracing.TRACE_HEADER_KEY] = []string{"mytrace"}
		r.Header[tracing.SPAN_HEADER_KEY] = []string{"myspan"}

		//WHEN
		middleware.ServeHTTP(resp, r)

		//THEN
		ctx := outRequest.Context()
		assert.Equal(t, "myspan", ctx.Value(tracing.SPAN_KEY))
		assert.Equal(t, "mytrace", ctx.Value(tracing.TRACE_KEY))
	})

	t.Run("wihtout trace and span should not change the context", func(t *testing.T) {
		//GIVEN
		var enhancedRequest *http.Request
		middleware := tracing.NewTracingMiddleware(func(w http.ResponseWriter, r *http.Request) {
			enhancedRequest = r
		})
		resp := httptest.NewRecorder()

		r, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)

		//WHEN
		middleware.ServeHTTP(resp, r)

		//THEN
		ctx := enhancedRequest.Context()
		assert.Equal(t, ctx, r.Context())
	})
}
