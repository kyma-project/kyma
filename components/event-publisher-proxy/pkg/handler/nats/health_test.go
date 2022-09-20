package nats_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats/mock"
)

func TestReadinessCheck(t *testing.T) {
	testCases := []struct {
		name                   string
		givenHandlerEndpoint   string
		wantPanicForNilHandler bool
		wantHandlerStatusCode  int
	}{
		{
			name:                   "should panic if provided handler is nil",
			wantPanicForNilHandler: true,
		},
		{
			name:                   "should report handler status-code",
			givenHandlerEndpoint:   "/endpoint",
			wantPanicForNilHandler: false,
			wantHandlerStatusCode:  health.StatusCodeHealthy,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			// test in both default and jetstream NATS modes
			for _, serverMode := range testingutils.NATSServerModes {
				t.Run(serverMode.Name, func(t *testing.T) {
					defer func() {
						r := recover()
						if !assert.Equal(t, tc.wantPanicForNilHandler, r != nil) {
							t.Log(r)
						}
					}()

					handlerMock := mock.StartOrDie(context.TODO(), t, mock.WithJetStream(serverMode.JetStreamEnabled))
					defer handlerMock.Stop()

					var handler http.HandlerFunc
					if tc.wantPanicForNilHandler {
						handler = nats.ReadinessCheck(nil)
					} else {
						handler = nats.ReadinessCheck(handlerMock.GetHandler())
					}

					assertResponseStatusCode(t, tc.givenHandlerEndpoint, handler, tc.wantHandlerStatusCode)
				})
			}
		})
	}
}

func assertResponseStatusCode(t *testing.T, endpoint string, handler http.HandlerFunc, statusCode int) {
	writer := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, endpoint, nil)

	handler.ServeHTTP(writer, request)

	require.Equal(t, statusCode, writer.Result().StatusCode)
}
