package nats_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats/mock"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestHandlerHealth(t *testing.T) {
	testCases := []struct {
		name                    string
		givenNatsServerShutdown bool
		wantLivenessStatusCode  int
		wantReadinessStatusCode int
	}{
		{
			name:                    "NATS handler is healthy",
			givenNatsServerShutdown: false,
			wantLivenessStatusCode:  health.StatusCodeHealthy,
			wantReadinessStatusCode: health.StatusCodeHealthy,
		},
		{
			name:                    "NATS handler is unhealthy",
			givenNatsServerShutdown: true,
			wantLivenessStatusCode:  health.StatusCodeHealthy,
			wantReadinessStatusCode: health.StatusCodeNotHealthy,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handlerMock := mock.StartOrDie(context.TODO(), t)
			defer handlerMock.Stop()

			if tc.givenNatsServerShutdown {
				handlerMock.ShutdownNatsServerAndWait()
			}

			testingutils.WaitForEndpointStatusCodeOrFail(handlerMock.GetLivenessEndpoint(), tc.wantLivenessStatusCode)
			testingutils.WaitForEndpointStatusCodeOrFail(handlerMock.GetReadinessEndpoint(), tc.wantReadinessStatusCode)
		})
	}
}
