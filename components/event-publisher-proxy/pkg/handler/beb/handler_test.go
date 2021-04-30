package beb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

const (
	// mock server endpoints
	defaultTokenEndpoint         = "/token"
	defaultEventsEndpoint        = "/events"
	defaultEventsHTTP400Endpoint = "/events400"

	// request size
	smallRequestSize = 2
	bigRequestSize   = 65536
)

func TestHandlerForCloudEvents(t *testing.T) {
	t.Parallel()

	exec := func(t *testing.T, applicationName, expectedApplicationName string) {
		var (
			port               = testingutils.GeneratePortOrDie()
			requestSize        = bigRequestSize
			healthEndpoint     = fmt.Sprintf("http://localhost:%d/healthz", port)
			bebNs              = testingutils.MessagingNamespace
			eventTypePrefix    = testingutils.MessagingEventTypePrefix
			eventsEndpoint     = defaultEventsEndpoint
			requestTimeout     = time.Second
			serverResponseTime = time.Nanosecond
			publishEndpoint    = fmt.Sprintf("http://localhost:%d/publish", port)
		)

		cancel, mockServer, collector := setupTestResources(t, port, requestSize, applicationName,
			expectedApplicationName, healthEndpoint, bebNs, eventTypePrefix, eventsEndpoint,
			requestTimeout, serverResponseTime)
		defer cancel()
		defer mockServer.Close()

		for _, testCase := range handlertest.TestCasesForCloudEvents {
			t.Run(testCase.Name, func(tt *testing.T) {
				body, headers := testCase.ProvideMessage()
				resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
				if err != nil {
					tt.Fatalf("failed to send event with error: %v", err)
				}
				_ = resp.Body.Close()
				if testCase.WantStatusCode != resp.StatusCode {
					tt.Fatalf("Test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
				}
				if testingutils.Is2XX(resp.StatusCode) {
					metricstest.EnsureMetricLatency(t, collector)
				}
			})
		}
	}

	// make sure not to change the cloudevent, even if its event-type contains none-alphanumeric characters
	exec(t, testingutils.ApplicationName, testingutils.ApplicationNameNotClean)
	exec(t, testingutils.ApplicationNameNotClean, testingutils.ApplicationNameNotClean)
}

func TestHandlerForLegacyEvents(t *testing.T) {
	t.Parallel()

	exec := func(t *testing.T, applicationName, expectedApplicationName string) {
		var (
			port                  = testingutils.GeneratePortOrDie()
			requestSize           = bigRequestSize
			healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", port)
			bebNs                 = testingutils.MessagingNamespace
			eventTypePrefix       = testingutils.MessagingEventTypePrefix
			eventsEndpoint        = defaultEventsEndpoint
			requestTimeout        = time.Second
			serverResponseTime    = time.Nanosecond
			publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/%s/v1/events", port, applicationName)
		)

		cancel, mockServer, collector := setupTestResources(t, port, requestSize, applicationName,
			expectedApplicationName, healthEndpoint, bebNs, eventTypePrefix, eventsEndpoint,
			requestTimeout, serverResponseTime)
		defer cancel()
		defer mockServer.Close()

		for _, testCase := range handlertest.TestCasesForLegacyEvents {
			t.Run(testCase.Name, func(t *testing.T) {
				body, headers := testCase.ProvideMessage()

				resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
				if err != nil {
					t.Fatalf("Failed to send event with error: %v", err)
				}

				if testCase.WantStatusCode != resp.StatusCode {
					t.Fatalf("Test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
				}

				if testCase.WantStatusCode == http.StatusOK {
					handlertest.ValidateOkResponse(t, *resp, &testCase.WantResponse)
				} else {
					handlertest.ValidateErrorResponse(t, *resp, &testCase.WantResponse)
				}

				if testingutils.Is2XX(resp.StatusCode) {
					metricstest.EnsureMetricLatency(t, collector)
				}
			})
		}
	}

	// make sure to clean the legacy event, so that its event-type is free from none-alphanumeric characters
	exec(t, testingutils.ApplicationName, testingutils.ApplicationName)
	exec(t, testingutils.ApplicationNameNotClean, testingutils.ApplicationName)
}

func TestHandlerForBEBFailures(t *testing.T) {
	t.Parallel()

	var (
		port                  = testingutils.GeneratePortOrDie()
		requestSize           = bigRequestSize
		applicationName       = testingutils.ApplicationName
		healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", port)
		bebNs                 = testingutils.MessagingNamespace
		eventTypePrefix       = testingutils.MessagingEventTypePrefix
		eventsEndpoint        = defaultEventsHTTP400Endpoint
		requestTimeout        = time.Second
		serverResponseTime    = time.Nanosecond
		publishEndpoint       = fmt.Sprintf("http://localhost:%d/publish", port)
		publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/%s/v1/events", port, applicationName)
	)

	cancel, mockServer, collector := setupTestResources(t, port, requestSize, applicationName, applicationName,
		healthEndpoint, bebNs, eventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime)
	defer cancel()
	defer mockServer.Close()

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
				return fmt.Sprintf(`"%s"`, testingutils.CloudEventData), testingutils.GetBinaryMessageHeaders()
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
				t.Fatalf("failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Fatalf("test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}

			if testCase.endPoint == publishLegacyEndpoint {
				handlertest.ValidateErrorResponse(t, *resp, &testCase.wantResponse)
			}

			if testingutils.Is2XX(resp.StatusCode) {
				metricstest.EnsureMetricLatency(t, collector)
			}
		})
	}
}

func TestHandlerForHugeRequests(t *testing.T) {
	t.Parallel()

	var (
		port                  = testingutils.GeneratePortOrDie()
		requestSize           = smallRequestSize
		applicationName       = testingutils.ApplicationName
		healthEndpoint        = fmt.Sprintf("http://localhost:%d/healthz", port)
		bebNs                 = testingutils.MessagingNamespace
		eventTypePrefix       = testingutils.MessagingEventTypePrefix
		eventsEndpoint        = defaultEventsHTTP400Endpoint
		requestTimeout        = time.Second
		serverResponseTime    = time.Nanosecond
		publishLegacyEndpoint = fmt.Sprintf("http://localhost:%d/%s/v1/events", port, applicationName)
	)

	cancel, mockServer, collector := setupTestResources(t, port, requestSize, applicationName, applicationName,
		healthEndpoint, bebNs, eventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime)
	defer cancel()
	defer mockServer.Close()

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
			endPoint:       handler.PublishEndpoint,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body, headers := testCase.provideMessage()
			_ = legacyapi.PublishEventResponses{}
			resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
			if err != nil {
				t.Fatalf("failed to send event with error: %v", err)
			}

			if testCase.wantStatusCode != resp.StatusCode {
				t.Fatalf("test failed, want status code:%d but got:%d", testCase.wantStatusCode, resp.StatusCode)
			}

			if testingutils.Is2XX(resp.StatusCode) {
				metricstest.EnsureMetricLatency(t, collector)
			}
		})
	}
}

func TestHandlerForSubscribedEndpoint(t *testing.T) {
	t.Parallel()

	var (
		port                     = testingutils.GeneratePortOrDie()
		requestSize              = smallRequestSize
		applicationName          = testingutils.ApplicationName
		healthEndpoint           = fmt.Sprintf("http://localhost:%d/healthz", port)
		bebNs                    = testingutils.MessagingNamespace
		eventTypePrefix          = testingutils.MessagingEventTypePrefix
		eventsEndpoint           = defaultEventsHTTP400Endpoint
		requestTimeout           = time.Second
		serverResponseTime       = time.Nanosecond
		subscribedEndpointFormat = "http://localhost:%d/%s/v1/events/subscribed"
	)

	cancel, mockServer, _ := setupTestResources(t, port, requestSize, applicationName, applicationName, healthEndpoint,
		bebNs, eventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime)
	defer cancel()
	defer mockServer.Close()

	for _, testCase := range handlertest.TestCasesForSubscribedEndpoint {
		t.Run(testCase.Name, func(t *testing.T) {
			subscribedURL := fmt.Sprintf(subscribedEndpointFormat, port, testCase.AppName)
			resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
			if err != nil {
				t.Fatalf("failed to send event with error: %v", err)
			}

			if testCase.WantStatusCode != resp.StatusCode {
				t.Fatalf("test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			respBodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to convert body to bytes: %v", err)
			}
			gotEventsResponse := subscribed.Events{}
			err = json.Unmarshal(respBodyBytes, &gotEventsResponse)
			if err != nil {
				t.Fatalf("failed to unmarshal body bytes to events response: %v", err)
			}
			if !reflect.DeepEqual(testCase.WantResponse, gotEventsResponse) {
				t.Fatalf("incorrect response, wanted: %v, got: %v", testCase.WantResponse, gotEventsResponse)
			}
		})
	}
}

func TestHandlerTimeout(t *testing.T) {
	t.Parallel()

	var (
		port               = testingutils.GeneratePortOrDie()
		requestSize        = bigRequestSize
		applicationName    = testingutils.ApplicationName
		healthEndpoint     = fmt.Sprintf("http://localhost:%d/healthz", port)
		bebNs              = testingutils.MessagingNamespace
		eventTypePrefix    = testingutils.MessagingEventTypePrefix
		eventsEndpoint     = defaultEventsHTTP400Endpoint
		requestTimeout     = time.Nanosecond  // short request timeout
		serverResponseTime = time.Millisecond // long server response time
		publishEndpoint    = fmt.Sprintf("http://localhost:%d/publish", port)
	)

	cancel, mockServer, collector := setupTestResources(t, port, requestSize, applicationName, applicationName,
		healthEndpoint, bebNs, eventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime)
	defer cancel()
	defer mockServer.Close()

	body, headers := testingutils.StructuredCloudEventPayload, testingutils.GetStructuredMessageHeaders()
	resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
	if err != nil {
		t.Fatalf("Failed to send event with error: %v", err)
	}
	_ = resp.Body.Close()
	if http.StatusInternalServerError != resp.StatusCode {
		t.Fatalf("Test failed, want status code:%d but got:%d", http.StatusInternalServerError, resp.StatusCode)
	}
	metricstest.EnsureMetricErrors(t, collector)
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
		got := handler.IsARequestWithLegacyEvent(tc.inputURI)
		if tc.wantedResult != got {
			t.Errorf("incorrect result with inputURI: %s, wanted: %v, got: %v", tc.inputURI, tc.wantedResult, got)
		}
	}
}

func setupTestResources(t *testing.T, port, maxRequestSize int, applicationName, expectedApplicationName,
	healthEndpoint, bebNs, eventTypePrefix, eventsEndpoint string, requestTimeout,
	serverResponseTime time.Duration) (context.CancelFunc, *testingutils.MockServer, *metrics.Collector) {
	validator := validateApplicationName(expectedApplicationName)
	mockServer := testingutils.NewMockServer(testingutils.WithResponseTime(serverResponseTime), testingutils.WithValidator(validator))
	mockServer.Start(t, defaultTokenEndpoint, defaultEventsEndpoint, defaultEventsHTTP400Endpoint)

	ctx, cancel := context.WithCancel(context.Background())

	emsCEURL := fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint)
	authURL := fmt.Sprintf("%s%s", mockServer.URL(), defaultTokenEndpoint)
	cfg := testingutils.NewEnvConfig(
		emsCEURL,
		authURL,
		testingutils.WithPort(port),
		testingutils.WithBEBNamespace(bebNs),
		testingutils.WithRequestTimeout(requestTimeout),
		testingutils.WithEventTypePrefix(eventTypePrefix),
	)
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	msgSender := sender.NewBebMessageSender(emsCEURL, client)
	msgReceiver := receiver.NewHttpMessageReceiver(cfg.Port)
	opts := &options.Options{MaxRequestSize: int64(maxRequestSize)}
	appLister := handlertest.NewApplicationListerOrDie(ctx, applicationName)
	legacyTransformer := legacy.NewTransformer(cfg.BEBNamespace, cfg.EventTypePrefix, appLister)

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
	informers.WaitForCacheSyncOrDie(ctx, dFilteredSharedInfFactory)
	t.Logf("Informers resynced successfully")
	subLister := genericInf.Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             cfg,
		Logger:             logrus.New(),
	}

	collector := metrics.NewCollector()
	msgHandler := NewHandler(msgReceiver, msgSender, cfg.RequestTimeout, legacyTransformer, opts, subscribedProcessor, logrus.New(), collector)
	go func() {
		if err := msgHandler.Start(ctx); err != nil {
			t.Errorf("failed to start handler with error: %v", err)
		}
	}()
	testingutils.WaitForHandlerToStart(t, healthEndpoint)

	return cancel, mockServer, collector
}

func validateApplicationName(appName string) testingutils.Validator {
	return func(r *http.Request) error {
		if r.Header.Get(cev2http.ContentType) == cev2event.ApplicationCloudEventsJSON {
			// CE structured mode
			message := cev2http.NewMessageFromHttpRequest(r)
			defer func() { _ = message.Finish(nil) }()
			var event *cev2event.Event
			var err error
			if event, err = binding.ToEvent(context.Background(), message); err != nil {
				return err
			}
			if strings.Contains(event.Type(), appName) {
				return nil
			}
			return errors.New(fmt.Sprintf("expected the event-type [%s] to contain the application name [%s] for structured mode", event.Type(), appName))
		}
		// default CE binary mode
		for k, v := range r.Header {
			if strings.ToLower(k) != strings.ToLower(testingutils.CeTypeHeader) {
				continue
			}
			if strings.Contains(v[0], appName) {
				return nil
			}
			return errors.New(fmt.Sprintf("expected the event-type [%s] to contain the application name [%s] for binary mode", v[0], appName))
		}
		return errors.New("event-type header is not found")
	}
}
