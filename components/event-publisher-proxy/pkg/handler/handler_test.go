package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/sirupsen/logrus"

	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

const (
	// mock server endpoints
	tokenEndpoint         = "/token"
	eventsEndpoint        = "/events"
	eventsHTTP400Endpoint = "/events400"
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

	port, err := generatePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}

	var (
		healthEndpoint  = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishEndpoint = fmt.Sprintf("http://localhost:%d/publish", port)
	)

	mockServer := testingutils.NewMockServer()
	mockServer.Start(t, tokenEndpoint, eventsEndpoint, eventsHTTP400Endpoint)
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
	cfg := testingutils.NewEnvConfig(
		emsCEURL,
		authURL,
		testingutils.WithPort(port),
	)
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(emsCEURL, client)

	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	legacyTransformer := legacy.NewTransformer("beb.ns", "event.type.prefix")
	opts := &options.Options{MaxRequestSize: 65536}
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t, healthEndpoint)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()
			resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
			if err != nil {
				t.Errorf("failed to send event with error: %v", err)
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

	port, err := generatePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	var (
		requestTimeout     = time.Nanosecond  // short request timeout
		serverResponseTime = time.Millisecond // long server response time
		healthEndpoint     = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishEndpoint    = fmt.Sprintf("http://localhost:%d/publish", port)
	)
	mockServer := testingutils.NewMockServer(testingutils.WithResponseTime(serverResponseTime))
	mockServer.Start(t, tokenEndpoint, eventsEndpoint, "")
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)
	cfg := testingutils.NewEnvConfig(emsCEURL,
		authURL,
		testingutils.WithPort(port),
		testingutils.WithRequestTimeout(requestTimeout),
	)
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(emsCEURL, client)

	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)

	legacyTransformer := legacy.NewTransformer("beb.ns", "event.type.prefix")
	opts := &options.Options{MaxRequestSize: 65536}
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
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

func TestHandlerForLegacyEvents(t *testing.T) {
	t.Parallel()
	port, err := generatePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	var (
		healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/app/v1/events", port)
		bebNs                 = "/beb.namespace"
		eventTypePrefix       = "event.type.prefix"
	)
	testCases := []struct {
		name           string
		targetEndpoint string
		provideMessage func() (string, http.Header)
		wantStatusCode int
		wantResponse   legacyapi.PublishEventResponses
	}{
		{
			name: "Send a legacy event successfully with event-id",
			provideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithEventId, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusOK,
			wantResponse: legacyapi.PublishEventResponses{
				Ok: &legacyapi.PublishResponse{
					EventID: "8945ec08-256b-11eb-9928-acde48001122",
					Status:  "",
					Reason:  "",
				},
			},
		},
		{
			name: "Send a legacy event successfully without event-id",
			provideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithoutEventId, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusOK,
			wantResponse: legacyapi.PublishEventResponses{
				Ok: &legacyapi.PublishResponse{
					EventID: "",
					Status:  "",
					Reason:  "",
				},
			},
		},
		{
			name: "Send a legacy event with invalid event id",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithInvalidEventId, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: getInvalidValidationErrorFor("event-id"),
			},
		},
		{
			name: "Send a legacy event without event time",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithoutEventTime, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationErrorFor("event-time"),
			},
		},
		{
			name: "Send a legacy event without event type",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithoutEventType, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationErrorFor("event-type"),
			},
		},
		{
			name: "Send a legacy event with invalid event time",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithInvalidEventTime, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: getInvalidValidationErrorFor("event-time"),
			},
		},
		{
			name: "Send a legacy event without event version",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithWithoutEventVersion, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationErrorFor("event-type-version"),
			},
		},
		{
			name: "Send a legacy event without data field",
			provideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithoutData, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: getMissingFieldValidationErrorFor("data"),
			},
		},
	}

	mockServer := testingutils.NewMockServer()
	mockServer.Start(t, tokenEndpoint, eventsEndpoint, eventsHTTP400Endpoint)
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bebCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)

	cfg := testingutils.NewEnvConfig(
		bebCEURL,
		authURL,
		testingutils.WithPort(port),
		testingutils.WithBEBNamespace(bebNs),
		testingutils.WithEventTypePrefix(eventTypePrefix),
	)
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(bebCEURL, client)
	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	legacyTransformer := legacy.NewTransformer(cfg.BEBNamespace, cfg.EventTypePrefix)
	opts := &options.Options{MaxRequestSize: 65536}
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t, healthEndpoint)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()

			resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
			if err != nil {
				t.Errorf("Failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("Test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}

			if testCase.wantStatusCode == http.StatusOK {
				validateOkResponse(t, *resp, &testCase.wantResponse)
			} else {
				validateErrorResponse(t, *resp, &testCase.wantResponse)
			}
		})
	}
}

func TestHandlerForBEBFailures(t *testing.T) {
	t.Parallel()
	port, err := generatePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	var (
		healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/app/v1/events", port)
		publishEndpoint       = fmt.Sprintf("http://localhost:%d/publish", port)
		bebNs                 = "/beb.namespace"
		eventTypePrefix       = "event.type.prefix"
	)
	mockServer := testingutils.NewMockServer()
	mockServer.Start(t, tokenEndpoint, eventsEndpoint, eventsHTTP400Endpoint)
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	beb400CEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsHTTP400Endpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)

	cfg := testingutils.NewEnvConfig(
		beb400CEURL,
		authURL,
		testingutils.WithPort(port),
		testingutils.WithBEBNamespace(bebNs),
		testingutils.WithEventTypePrefix(eventTypePrefix),
	)
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(beb400CEURL, client)
	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	legacyTransformer := legacy.NewTransformer(cfg.BEBNamespace, cfg.EventTypePrefix)
	opts := &options.Options{MaxRequestSize: 65536}
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t, healthEndpoint)

	testCases := []struct {
		name           string
		targetEndpoint string
		provideMessage func() (string, http.Header)
		endPoint       string
		wantStatusCode int
		wantResponse   legacyapi.PublishEventResponses
	}{
		{
			name: "Send a legacy event with event-id",
			provideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithEventId, testingutils.GetApplicationJSONHeaders()
			},
			endPoint:       publishLegacyEndpoint,
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: &legacyapi.Error{
					Status:  400,
					Message: "invalid request"},
			},
		},
		{
			name: "Binary CloudEvent is valid with required headers",
			provideMessage: func() (string, http.Header) {
				return testingutils.BinaryCloudEventPayload, testingutils.GetBinaryMessageHeaders()
			},
			endPoint:       publishEndpoint,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()
			_ = legacyapi.PublishEventResponses{}
			resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
			if err != nil {
				t.Errorf("failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}

			if testCase.endPoint == publishLegacyEndpoint {
				validateErrorResponse(t, *resp, &testCase.wantResponse)
			}
		})
	}
}

func TestHandlerForHugeRequests(t *testing.T) {
	t.Parallel()
	port, err := generatePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	var (
		healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", port)
		publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/app/v1/events", port)
		bebNs                 = "/beb.namespace"
		eventTypePrefix       = "event.type.prefix"
	)
	mockServer := testingutils.NewMockServer()
	mockServer.Start(t, tokenEndpoint, eventsEndpoint, eventsHTTP400Endpoint)
	defer mockServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	beb400CEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsHTTP400Endpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), tokenEndpoint)

	cfg := testingutils.NewEnvConfig(
		beb400CEURL,
		authURL,
		testingutils.WithPort(port),
		testingutils.WithBEBNamespace(bebNs),
		testingutils.WithEventTypePrefix(eventTypePrefix),
	)
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewHttpMessageSender(beb400CEURL, client)
	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	legacyTransformer := legacy.NewTransformer(cfg.BEBNamespace, cfg.EventTypePrefix)

	// Limiting the accepted size of the request to 2 Bytes
	opts := &options.Options{MaxRequestSize: 2}
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	waitForHandlerToStart(t, healthEndpoint)

	testCases := []struct {
		name           string
		targetEndpoint string
		provideMessage func() (string, http.Header)
		endPoint       string
		wantStatusCode int
	}{
		{
			name: "Should fail with HTTP 413 with a request which is larger than 2 Bytes as the maximum accepted size is 2 Bytes",
			provideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithEventId, testingutils.GetApplicationJSONHeaders()
			},
			endPoint:       publishLegacyEndpoint,
			wantStatusCode: http.StatusRequestEntityTooLarge,
		},
		{
			name: "Should accept a request which is lesser than 2 Bytes as the maximum accepted size is 2 Bytes",
			provideMessage: func() (string, http.Header) {
				return "{}", testingutils.GetBinaryMessageHeaders()
			},
			endPoint:       publishEndpoint,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()
			_ = legacyapi.PublishEventResponses{}
			resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
			if err != nil {
				t.Errorf("failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}
		})
	}
}

// getMissingFieldValidationErrorFor generates an Error message for a missing field
func getMissingFieldValidationErrorFor(field string) *legacyapi.Error {
	return &legacyapi.Error{
		Status:  400,
		Type:    "validation_violation",
		Message: "Missing field",
		Details: []legacyapi.ErrorDetail{
			{
				Field:    field,
				Type:     "missing_field",
				Message:  "Missing field",
				MoreInfo: "",
			},
		},
	}
}

// isValidEventID checks whether EventID is valid or not
func isValidEventID(id string) bool {
	return regexp.MustCompile(legacy.AllowedEventIDChars).MatchString(id)
}

// getInvalidValidationErrorFor generates an Error message for an invalid field
func getInvalidValidationErrorFor(field string) *legacyapi.Error {
	return &legacyapi.Error{
		Status:  400,
		Type:    "validation_violation",
		Message: "Invalid field",
		Details: []legacyapi.ErrorDetail{
			{
				Field:    field,
				Type:     "invalid_field",
				Message:  "Invalid field",
				MoreInfo: "",
			},
		},
	}
}

func TestIsARequestWithLegacyEvent(t *testing.T) {
	testCases := []struct {
		inputURI     string
		wantedResult bool
	}{
		{
			inputURI:     "/app/v1/events",
			wantedResult: true,
		},
		{
			inputURI:     "///app/v1/events",
			wantedResult: true,
		},
		{
			inputURI:     "///app/v1//events",
			wantedResult: false,
		},
		{
			inputURI:     "///app/v1/foo/events",
			wantedResult: false,
		},
	}

	for _, tc := range testCases {
		got := isARequestWithLegacyEvent(tc.inputURI)
		if tc.wantedResult != got {
			t.Errorf("incorrect result with inputURI: %s, wanted: %v, got: %v", tc.inputURI, tc.wantedResult, got)
		}
	}
}

// validateErrorResponse validates Error Response
func validateErrorResponse(t *testing.T, resp http.Response, tcWantResponse *legacyapi.PublishEventResponses) {
	legacyResponse := legacyapi.PublishEventResponses{}
	legacyError := legacyapi.Error{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	t.Logf("response body : %s", string(bodyBytes))
	if err = json.Unmarshal(bodyBytes, &legacyError); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	legacyResponse.Error = &legacyError
	if !reflect.DeepEqual(tcWantResponse.Error, legacyResponse.Error) {
		t.Fatalf("Invalid error, want: %v, got: %v", tcWantResponse.Error, legacyResponse.Error)
	}
}

// validateOkResponse validates Ok Response
func validateOkResponse(t *testing.T, resp http.Response, tcWantResponse *legacyapi.PublishEventResponses) {
	legacyOkResponse := legacyapi.PublishResponse{}
	legacyResponse := legacyapi.PublishEventResponses{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	t.Logf("response body : %s", string(bodyBytes))
	if err = json.Unmarshal(bodyBytes, &legacyOkResponse); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	legacyResponse.Ok = &legacyOkResponse
	if err = resp.Body.Close(); err != nil {
		t.Fatalf("failed to close body: %v", err)
	}

	if tcWantResponse.Ok.EventID != "" && tcWantResponse.Ok.EventID != legacyResponse.Ok.EventID {
		t.Errorf("invalid event-id, want: %v, got: %v", tcWantResponse.Ok.EventID, legacyResponse.Ok.EventID)
	}

	if tcWantResponse.Ok.EventID == "" && !isValidEventID(legacyResponse.Ok.EventID) {
		t.Errorf("should match regex: [%s] Not a valid event-id: %v ", legacy.AllowedEventIDChars, legacyResponse.Ok.EventID)
	}
	if tcWantResponse.Ok.Reason != legacyResponse.Ok.Reason {
		t.Errorf("invalid reason, want: %v, got: %v", tcWantResponse.Ok.Reason, legacyResponse.Ok.Reason)
	}
	if tcWantResponse.Ok.Status != legacyResponse.Ok.Status {
		t.Errorf("invalid status, want: %v, got: %v", tcWantResponse.Ok.Status, legacyResponse.Ok.Status)
	}
}

// generatePort generates a random 5 digit port
func generatePort() (int, error) {
	max := 4
	// Add 4 as prefix to make it 5 digits but less than 65535
	add4AsPrefix := "4"
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return 0, err
	}
	if err != nil {
		return 0, err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	num, err := strconv.Atoi(fmt.Sprintf("%s%s", add4AsPrefix, string(b)))
	if err != nil {
		return 0, err
	}

	return num, nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9'}
