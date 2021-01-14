package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	"github.com/sirupsen/logrus"
)

const (
	// mock server endpoints
	tokenEndpoint         = "/token"
	eventsEndpoint        = "/events"
	eventsHTTP400Endpoint = "/events400"
)

var (
	testCasesForCloudEvents = []struct {
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
	testCasesForLegacyEvents = []struct {
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
				Error: testingutils.GetInvalidValidationErrorFor("event-id"),
			},
		},
		{
			name: "Send a legacy event without event time",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithoutEventTime, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: testingutils.GetMissingFieldValidationErrorFor("event-time"),
			},
		},
		{
			name: "Send a legacy event without event type",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithoutEventType, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: testingutils.GetMissingFieldValidationErrorFor("event-type"),
			},
		},
		{
			name: "Send a legacy event with invalid event time",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithInvalidEventTime, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: testingutils.GetInvalidValidationErrorFor("event-time"),
			},
		},
		{
			name: "Send a legacy event without event version",
			provideMessage: func() (string, http.Header) {
				return testingutils.LegacyEventPayloadWithWithoutEventVersion, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: testingutils.GetMissingFieldValidationErrorFor("event-type-version"),
			},
		},
		{
			name: "Send a legacy event without data field",
			provideMessage: func() (string, http.Header) {
				return testingutils.ValidLegacyEventPayloadWithoutData, testingutils.GetApplicationJSONHeaders()
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse: legacyapi.PublishEventResponses{
				Error: testingutils.GetMissingFieldValidationErrorFor("data"),
			},
		},
	}
	testCasesForSubscribedEndpoit = []struct {
		name               string
		appName            string
		inputSubscriptions []eventingv1alpha1.Subscription
		wantStatusCode     int
		wantResponse       subscribed.Events
	}{
		{
			name:           "Send a request with a valid application name",
			appName:        "valid-app",
			wantStatusCode: http.StatusOK,
			wantResponse: subscribed.Events{
				EventsInfo: []subscribed.Event{{
					Name:    "order.created",
					Version: "v1",
				}},
			},
		}, {
			name:           "Send a request with an invalid application name",
			appName:        "invalid-app",
			wantStatusCode: http.StatusOK,
			wantResponse: subscribed.Events{
				EventsInfo: []subscribed.Event{},
			},
		},
	}
)

func TestHandler(t *testing.T) {
	t.Parallel()

	port, err := testingutils.GeneratePort()
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
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, nil, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	testingutils.WaitForHandlerToStart(t, healthEndpoint)

	for _, testCase := range testCasesForCloudEvents {
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

	port, err := testingutils.GeneratePort()
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
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, nil, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	testingutils.WaitForHandlerToStart(t, healthEndpoint)

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

func TestHandlerForLegacyEvents(t *testing.T) {
	t.Parallel()
	port, err := testingutils.GeneratePort()
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
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, nil, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	testingutils.WaitForHandlerToStart(t, healthEndpoint)

	for _, testCase := range testCasesForLegacyEvents {
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
				testingutils.ValidateOkResponse(t, *resp, &testCase.wantResponse)
			} else {
				testingutils.ValidateErrorResponse(t, *resp, &testCase.wantResponse)
			}
		})
	}
}

func TestHandlerForBEBFailures(t *testing.T) {
	t.Parallel()
	port, err := testingutils.GeneratePort()
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
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, nil, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	testingutils.WaitForHandlerToStart(t, healthEndpoint)

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
				testingutils.ValidateErrorResponse(t, *resp, &testCase.wantResponse)
			}
		})
	}
}

func TestHandlerForHugeRequests(t *testing.T) {
	t.Parallel()
	port, err := testingutils.GeneratePort()
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
	handler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, nil, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()

	testingutils.WaitForHandlerToStart(t, healthEndpoint)

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

func TestHandlerForSubscribedEndpoint(t *testing.T) {
	t.Parallel()
	port, err := testingutils.GeneratePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	var (
		subscribedEndpointFormat = "http://localhost:%d/%s/v1/events/subscribed"
		healthEndpoint           = fmt.Sprintf("http://localhost:%d/healthz", port)
		bebNs                    = "/beb.namespace"
		eventTypePrefix          = "event.type.prefix"
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := testingutils.NewEnvConfig(
		// BEB details are not needed in this test
		"",
		"",
		testingutils.WithPort(port),
		testingutils.WithBEBNamespace(bebNs),
		testingutils.WithEventTypePrefix(eventTypePrefix),
	)
	recv := receiver.NewHttpMessageReceiver(cfg.Port)
	opts := &options.Options{MaxRequestSize: 65536}
	legacyTransformer := legacy.NewTransformer(cfg.BEBNamespace, cfg.EventTypePrefix)

	// Setting up fake informers
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 to scheme: %v", err)
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add eventing v1alpha1 to scheme: %v", err)
	}
	subscription := testingutils.NewSubscription()

	subUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(subscription)
	if err != nil {
		t.Fatalf("failed to convert subscription to unstructured obj: %v", err)
	}
	// Creating unstructured subscriptions
	subUnstructured := &unstructured.Unstructured{
		Object: subUnstructuredMap,
	}
	// Setting Kind information in unstructured subscription
	subscriptionGVK := schema.GroupVersionKind{
		Group:   subscribed.GVR.Group,
		Version: subscribed.GVR.Version,
		Kind:    "Subscription",
	}
	subUnstructured.SetGroupVersionKind(subscriptionGVK)
	// Configuring fake dynamic client
	dynamicTestClient := dynamicfake.NewSimpleDynamicClient(scheme, subUnstructured)

	dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicTestClient,
		10*time.Second,
		v1.NamespaceAll,
		nil,
	)
	genericInf := dFilteredSharedInfFactory.ForResource(subscribed.GVR)
	t.Logf("Waiting for cache to resync")
	subscribed.WaitForCacheSyncOrDie(ctx, dFilteredSharedInfFactory)
	t.Logf("Informers resynced successfully")
	subLister := genericInf.Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             cfg,
		Logger:             logrus.New(),
	}

	handler := NewHandler(recv, nil, cfg.RequestTimeout, legacyTransformer, opts, subscribedProcessor, logrus.New())
	go func() {
		if err := handler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()
	testingutils.WaitForHandlerToStart(t, healthEndpoint)

	for _, testCase := range testCasesForSubscribedEndpoit {
		t.Run(testCase.name, func(t *testing.T) {
			subscribedURL := fmt.Sprintf(subscribedEndpointFormat, port, testCase.appName)
			resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
			if err != nil {
				t.Errorf("failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Errorf("test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			respBodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("failed to convert body to bytes: %v", err)
			}
			gotEventsResponse := subscribed.Events{}
			err = json.Unmarshal(respBodyBytes, &gotEventsResponse)
			if err != nil {
				t.Errorf("failed to unmarshal body bytes to events response: %v", err)
			}
			if !reflect.DeepEqual(testCase.wantResponse, gotEventsResponse) {
				t.Errorf("incorrect response, wanted: %v, got: %v", testCase.wantResponse, gotEventsResponse)
			}
		})
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
