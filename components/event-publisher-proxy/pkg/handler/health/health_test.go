package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecker(t *testing.T) {
	testCases := []struct {
		name                        string
		useCustomLivenessCheck      bool             // use the givenCustomLivenessCheck instead of the default one
		useCustomReadinessCheck     bool             // use the givenCustomReadinessCheck instead of the default one
		givenCustomLivenessCheck    http.HandlerFunc // custom liveness check (can be nil)
		givenCustomReadinessCheck   http.HandlerFunc // custom readiness check (can be nil)
		givenNextHandler            http.Handler     // next handler
		givenNextHandlerEndpoint    string           // next handler endpoint
		wantPanicForNilHealthChecks bool             // panic if provided health checks is nil
		wantLivenessStatusCode      int              // expected liveness status code
		wantReadinessStatusCode     int              // expected readiness status code
		wantNextHandlerStatusCode   int              // expected next handler status code
	}{
		{
			name:                    "should report default health checks status-codes",
			useCustomLivenessCheck:  false,
			useCustomReadinessCheck: false,
			wantLivenessStatusCode:  StatusCodeHealthy,
			wantReadinessStatusCode: StatusCodeHealthy,
		},
		{
			name:                      "should report default health checks and next handler status-codes",
			useCustomLivenessCheck:    false,
			useCustomReadinessCheck:   false,
			givenNextHandler:          handlerWithStatusCode(http.StatusNoContent),
			givenNextHandlerEndpoint:  "/endpoint",
			wantLivenessStatusCode:    StatusCodeHealthy,
			wantReadinessStatusCode:   StatusCodeHealthy,
			wantNextHandlerStatusCode: http.StatusNoContent,
		},
		{
			name:                        "should panic if provided liveness check is nil",
			useCustomLivenessCheck:      true,
			useCustomReadinessCheck:     false,
			givenCustomLivenessCheck:    nil,
			wantPanicForNilHealthChecks: true,
		},
		{
			name:                        "should panic if provided readiness check is nil",
			useCustomLivenessCheck:      false,
			useCustomReadinessCheck:     true,
			givenCustomReadinessCheck:   nil,
			wantPanicForNilHealthChecks: true,
		},
		{
			name:                        "should report custom health checks and next handler status-codes",
			useCustomLivenessCheck:      true,
			useCustomReadinessCheck:     true,
			givenCustomLivenessCheck:    handlerFuncWithStatusCode(http.StatusOK),
			givenCustomReadinessCheck:   handlerFuncWithStatusCode(http.StatusAccepted),
			givenNextHandler:            handlerWithStatusCode(http.StatusNoContent),
			givenNextHandlerEndpoint:    "/endpoint",
			wantPanicForNilHealthChecks: false,
			wantLivenessStatusCode:      http.StatusOK,
			wantReadinessStatusCode:     http.StatusAccepted,
			wantNextHandlerStatusCode:   http.StatusNoContent,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()

				if !assert.Equal(t, tc.wantPanicForNilHealthChecks, r != nil) {
					t.Log(r)
				}
			}()

			var opts []CheckerOpt
			if tc.useCustomLivenessCheck {
				opts = append(opts, WithLivenessCheck(tc.givenCustomLivenessCheck))
			}
			if tc.useCustomReadinessCheck {
				opts = append(opts, WithReadinessCheck(tc.givenCustomReadinessCheck))
			}
			checker := NewChecker(opts...)

			if tc.useCustomLivenessCheck {
				assertResponseStatusCode(t, LivenessURI, checker, tc.givenCustomLivenessCheck, tc.wantLivenessStatusCode)
			} else {
				assertResponseStatusCode(t, LivenessURI, checker, http.HandlerFunc(DefaultCheck), StatusCodeHealthy)
			}

			if tc.useCustomReadinessCheck {
				assertResponseStatusCode(t, ReadinessURI, checker, tc.givenCustomReadinessCheck, tc.wantReadinessStatusCode)
			} else {
				assertResponseStatusCode(t, ReadinessURI, checker, http.HandlerFunc(DefaultCheck), StatusCodeHealthy)
			}

			if tc.givenNextHandler != nil {
				assertResponseStatusCode(t, tc.givenNextHandlerEndpoint, checker, tc.givenNextHandler, tc.wantNextHandlerStatusCode)
			}
		})
	}
}

func assertResponseStatusCode(t *testing.T, endpoint string, checker *Checker, handler http.Handler, statusCode int) {
	writer := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, endpoint, nil)

	checker.Check(handler).ServeHTTP(writer, request)

	require.Equal(t, statusCode, writer.Result().StatusCode)
}

func handlerFuncWithStatusCode(statusCode int) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(statusCode)
	}
}

func handlerWithStatusCode(statusCode int) http.Handler {
	return handlerFuncWithStatusCode(statusCode)
}
