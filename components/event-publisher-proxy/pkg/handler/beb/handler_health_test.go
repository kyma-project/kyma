package beb_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/beb/mock"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestHandlerHealth(t *testing.T) {
	var (
		requestSize        = bigRequestSize
		eventsEndpoint     = defaultEventsEndpoint
		requestTimeout     = time.Second
		serverResponseTime = time.Nanosecond
	)
	testCases := []struct {
		name                    string
		wantLivenessStatusCode  int
		wantReadinessStatusCode int
	}{
		{
			name:                    "beb handler is healthy",
			wantLivenessStatusCode:  health.StatusCodeHealthy,
			wantReadinessStatusCode: health.StatusCodeHealthy,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, testingutils.MessagingEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime)
			defer handlerMock.Close()

			testingutils.WaitForEndpointStatusCodeOrFail(handlerMock.GetLivenessEndpoint(), tc.wantLivenessStatusCode)
			testingutils.WaitForEndpointStatusCodeOrFail(handlerMock.GetReadinessEndpoint(), tc.wantReadinessStatusCode)
		})
	}
}
