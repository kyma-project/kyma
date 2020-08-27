package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"
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
			name: "Structured CloudEvent is missing id",
			provideMessage: func() (string, http.Header) {
				return newStructuredCloudEventPayload(ceType, ceSource, ceSpecVersion), getStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent is missing type",
			provideMessage: func() (string, http.Header) {
				return newStructuredCloudEventPayload(ceID, ceSource, ceSpecVersion), getStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent is missing source",
			provideMessage: func() (string, http.Header) {
				return newStructuredCloudEventPayload(ceID, ceType, ceSpecVersion), getStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent is missing specversion",
			provideMessage: func() (string, http.Header) {
				return newStructuredCloudEventPayload(ceID, ceType, ceSource), getStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Structured CloudEvent is valid with required attributes",
			provideMessage: func() (string, http.Header) {
				return newStructuredCloudEventPayload(ceID, ceType, ceSource, ceSpecVersion), getStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "Structured CloudEvent is valid with required attributes and more",
			provideMessage: func() (string, http.Header) {
				return newStructuredCloudEventPayload(ceID, ceType, ceSource, ceSpecVersion, ceData, ceTime, ceDataContentType), getStructuredMessageHeaders()
			},
			wantStatusCode: http.StatusNoContent,
		},
		// binary cloudevents
		{
			name: "Binary CloudEvent is missing CE-ID header",
			provideMessage: func() (string, http.Header) {
				headers := getBinaryMessageHeaders()
				headers.Del(ceIDHeader)
				return newBinaryCloudEventPayload(), headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent is missing CE-Type header",
			provideMessage: func() (string, http.Header) {
				headers := getBinaryMessageHeaders()
				headers.Del(ceTypeHeader)
				return newBinaryCloudEventPayload(), headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent is missing CE-Source header",
			provideMessage: func() (string, http.Header) {
				headers := getBinaryMessageHeaders()
				headers.Del(ceSourceHeader)
				return newBinaryCloudEventPayload(), headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent is missing CE-SpecVersion header",
			provideMessage: func() (string, http.Header) {
				headers := getBinaryMessageHeaders()
				headers.Del(ceSpecVersionHeader)
				return newBinaryCloudEventPayload(), headers
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Binary CloudEvent is valid with required headers",
			provideMessage: func() (string, http.Header) {
				return newBinaryCloudEventPayload(), getBinaryMessageHeaders()
			},
			wantStatusCode: http.StatusNoContent,
		},
	}

	logger := zap.NewExample()
	defer func() { _ = logger.Sync() }()

	emsMockServer := startEmsMockServer(t, emsTokenEndpoint, emsEventsEndpoint)
	defer emsMockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env := newEnvConfig(emsMockServer.URL, emsMockServer.URL)
	authCfg := oauth.Config(env)
	httpClient := authCfg.Client(ctx)
	connectionArgs := sender.ConnectionArgs{MaxIdleConns: 1, MaxIdleConnsPerHost: 1}
	messageSender, err := sender.NewHttpMessageSender(&connectionArgs, env.EmsCEURL, httpClient)
	if err != nil {
		t.Fatal("Unable to create message sender", zap.Error(err))
	}

	handler := NewHandler(receiver.NewHttpMessageReceiver(env.Port), messageSender, logger)
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Fatalf("Failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp, err := sendEvent(testCase.provideMessage())
			if err != nil {
				t.Errorf("Failed to send event with error: %v", err)
			}
			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("Test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}
		})
	}

}

func startEmsMockServer(t *testing.T, tokenEndpoint, eventsEndpoint string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case tokenEndpoint:
			{
				token := "access_token=some-secure-access-token&token_type=bearer&expires_in=86400"
				if _, err := w.Write([]byte(token)); err != nil {
					t.Errorf("Failed to write HTTP response")
				}
			}
		case eventsEndpoint:
			{
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			{
				t.Errorf("EMS Mock server supports the following endpoints only: [%s, %s]", tokenEndpoint, eventsEndpoint)
			}
		}
	}))
}

func waitForHandlerToStart(t *testing.T) {
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
				if resp, err := http.Get(handlerHealthEndpoint); err != nil {
					continue
				} else if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
	}
}

func sendEvent(body string, headers http.Header) (*http.Response, error) {
	bodyBytes := []byte(body)
	req, err := http.NewRequest(http.MethodPost, handlerPublishEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}

	client := http.Client{}
	defer client.CloseIdleConnections()

	return client.Do(req)
}

func newEnvConfig(emsUrl, authUrl string) gateway.EnvConfig {
	return gateway.EnvConfig{
		Port:          8080,
		EmsCEURL:      fmt.Sprintf("%s%s", emsUrl, emsEventsEndpoint),
		TokenEndpoint: fmt.Sprintf("%s%s", authUrl, emsTokenEndpoint),
	}
}

func newStructuredCloudEventPayload(attributes ...string) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	buffer.WriteString(strings.Join(attributes, ","))
	buffer.WriteString("}")
	return buffer.String()
}

func getStructuredMessageHeaders() http.Header {
	return http.Header{"Content-Type": []string{"application/cloudevents+json"}}
}

func newBinaryCloudEventPayload() string {
	return "payload"
}

func getBinaryMessageHeaders() http.Header {
	headers := make(http.Header)
	headers.Add(ceIDHeader, "testID")
	headers.Add(ceTypeHeader, "testType")
	headers.Add(ceSourceHeader, "testSource")
	headers.Add(ceSpecVersionHeader, "1.0")
	return headers
}

const (
	// ems endpoints
	emsTokenEndpoint  = "/token"
	emsEventsEndpoint = "/events"

	// handler endpoints
	handlerHealthEndpoint  = "http://localhost:8080/healthz"
	handlerPublishEndpoint = "http://localhost:8080/publish"

	// binary cloudevent headers
	ceIDHeader          = "CE-ID"
	ceTypeHeader        = "CE-Type"
	ceSourceHeader      = "CE-Source"
	ceSpecVersionHeader = "CE-SpecVersion"

	// structured cloudevent attributes
	ceID              = `"id":"testID"`
	ceType            = `"type":"testType"`
	ceSource          = `"source":"testSource"`
	ceSpecVersion     = `"specversion":"1.0"`
	ceData            = `"data":"{'key':'value'}"`
	ceTime            = `"time":"2020-01-01T17:01:00Z"`
	ceDataContentType = `"datacontenttype":"application/json"`
)
