package handler

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	// mock server endpoints
	tokenEndpoint  = "/token"
	eventsEndpoint = "/events"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		provideMessage func() (string, http.Header)
		wantStatusCode int
	}{
		// structured cloudevents
		{
			name: "Structured CloudEvent without id",
			provideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutID, testingutils.GetStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent without type",
			provideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutType, testingutils.GetStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent without specversion",
			provideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutSpecVersion, testingutils.GetStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent without source",
			provideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayloadWithoutSource, testingutils.GetStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent is valid",
			provideMessage: func() (string, http.Header) {
				return testingutils.StructuredCloudEventPayload, testingutils.GetStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusNoContent,
		},
		// binary cloudevents
		{
			name: "Binary CloudEvent without CE-ID header",
			provideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeIDHeader)
				return testingutils.BinaryCloudEventPayload, headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent without CE-Type header",
			provideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeTypeHeader)
				return testingutils.BinaryCloudEventPayload, headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent without CE-SpecVersion header",
			provideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeSpecVersionHeader)
				return testingutils.BinaryCloudEventPayload, headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent without CE-Source header",
			provideMessage: func() (string, http.Header) {
				headers := testingutils.GetBinaryMessageHeaders()
				headers.Del(testingutils.CeSourceHeader)
				return testingutils.BinaryCloudEventPayload, headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent is valid with required headers",
			provideMessage: func() (string, http.Header) {
				return testingutils.BinaryCloudEventPayload, testingutils.GetBinaryMessageHeaders()
			},
			wantStatusCode: http.StatusNoContent,
		},
	}

	var (
		port            = 8888
		healthEndpoint  = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishEndpoint = fmt.Sprintf("http://localhost:%d/publish", port)
	)

	mockServer := testingutils.NewMockServer()
	mockServer.Start(t, tokenEndpoint, eventsEndpoint)
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
	cfg := testingutils.NewEnvConfig(emsCEURL, authURL, testingutils.WithPort(port))
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(emsCEURL, client)

	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("Failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t, healthEndpoint)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()
			resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
			if err != nil {
				t.Errorf("Failed to send event with error: %v", err)
			}
			_ = resp.Body.Close()
			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("Test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}
		})
	}
}

func TestHandlerTimeout(t *testing.T) {
	t.Parallel()

	var (
		port               = 9999
		requestTimeout     = time.Nanosecond  // short request timeout
		serverResponseTime = time.Millisecond // long server response time
		healthEndpoint     = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishEndpoint    = fmt.Sprintf("http://localhost:%d/publish", port)
	)

	mockServer := testingutils.NewMockServer(testingutils.WithResponseTime(serverResponseTime))
	mockServer.Start(t, tokenEndpoint, eventsEndpoint)
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
	cfg := testingutils.NewEnvConfig(emsCEURL, authURL, testingutils.WithPort(port), testingutils.WithRequestTimeout(requestTimeout))
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(emsCEURL, client)

	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("Failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t, healthEndpoint)

	body, headers := testingutils.StructuredCloudEventPayload, testingutils.GetStructuredMessageHeaders()
	resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
	if err != nil {
		t.Errorf("Failed to send event with error: %v", err)
	}
	_ = resp.Body.Close()
	if http.StatusInternalServerError != resp.StatusCode {
		t.Errorf("Test failed, want status code:%d but got:%d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func waitForHandlerToStart(t *testing.T, healthEndpoint string) {
	timeout := time.After(time.Second * 30)
	tick := time.Tick(time.Second * 1)

	for {
		select {
		case <-timeout:
			{
				t.Fatal("Failed to start handler")
			}
		case <-tick:
			{
				if resp, err := http.Get(healthEndpoint); err != nil {
					continue
				} else if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
	}
}
