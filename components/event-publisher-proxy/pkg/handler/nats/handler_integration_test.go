package nats_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats/mock"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func TestNatsHandlerForCloudEvents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenApplicationName string
		givenEventType       string
		wantNatsSubject      string
	}{
		{
			name:                 "With prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantNatsSubject:      testingutils.CloudEventTypeNotClean,
		},
		{
			name:                 "With prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantNatsSubject:      testingutils.CloudEventTypeNotClean,
		},
		{
			name:                 "With empty prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantNatsSubject:      testingutils.CloudEventTypeNotClean,
		},
		{
			name:                 "With empty prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantNatsSubject:      testingutils.CloudEventTypeNotClean,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithApplication(tc.givenApplicationName),
			)

			// setup test environment
			publishEndpoint := fmt.Sprintf("http://localhost:%d/publish", handlerMock.GetNatsConfig().Port)
			subscription := testingutils.NewSubscription(
				testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, tc.givenEventType),
			)

			// prepare event type from subscription
			assert.NotNil(t, subscription.Spec.Filter)
			assert.NotEmpty(t, subscription.Spec.Filter.Filters)
			eventTypeToSubscribe := subscription.Spec.Filter.Filters[0].EventType.Value

			// connect to nats
			connection, err := pkgnats.Connect(handlerMock.GetNatsURL(), true, 3, time.Second)
			assert.Nil(t, err)
			assert.NotNil(t, connection)

			// publish a message to NATS and validate it
			validator := testingutils.ValidateNatsSubjectOrFail(t, tc.wantNatsSubject)
			testingutils.SubscribeToEventOrFail(t, connection, eventTypeToSubscribe, validator)

			// nolint:scopelint
			// run the tests for publishing cloudevents
			for _, testCase := range handlertest.TestCasesForCloudEvents {
				t.Run(testCase.Name, func(t *testing.T) {
					body, headers := testCase.ProvideMessage()
					resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
					if err != nil {
						t.Errorf("Failed to send event with error: %v", err)
					}
					_ = resp.Body.Close()
					if testCase.WantStatusCode != resp.StatusCode {
						t.Errorf("Test failed, want status code:%d but got:%d", testCase.WantStatusCode, resp.StatusCode)
					}
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestNatsHandlerForLegacyEvents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenApplicationName string
		givenEventType       string
		wantNatsSubject      string
	}{
		{
			name:                 "With prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantNatsSubject:      testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantNatsSubject:      testingutils.CloudEventType,
		},
		{
			name:                 "With prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantNatsSubject:      testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantNatsSubject:      testingutils.CloudEventType,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithApplication(tc.givenApplicationName),
			)

			// setup test environment
			publishLegacyEndpoint := fmt.Sprintf("http://localhost:%d/%s/v1/events", handlerMock.GetNatsConfig().Port, tc.givenApplicationName)
			subscription := testingutils.NewSubscription(testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, tc.givenEventType))

			// prepare event type from subscription
			assert.NotNil(t, subscription.Spec.Filter)
			assert.NotEmpty(t, subscription.Spec.Filter.Filters)
			eventTypeToSubscribe := subscription.Spec.Filter.Filters[0].EventType.Value

			// connect to nats
			connection, err := pkgnats.Connect(handlerMock.GetNatsURL(), true, 3, time.Second)
			assert.Nil(t, err)
			assert.NotNil(t, connection)

			// publish a message to NATS and validate it
			validator := testingutils.ValidateNatsSubjectOrFail(t, tc.wantNatsSubject)
			testingutils.SubscribeToEventOrFail(t, connection, eventTypeToSubscribe, validator)

			// nolint:scopelint
			// run the tests for publishing legacy events
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
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestNatsHandlerForSubscribedEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenEventType       string
	}{
		{
			name:                 "With prefix and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenEventType:       testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenEventType:       testingutils.CloudEventTypePrefixEmpty,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// setup test environment
			scheme := runtime.NewScheme()
			if err := corev1.AddToScheme(scheme); err != nil {
				require.NoError(t, err)
			}
			if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
				require.NoError(t, err)
			}

			subscribedEndpointFormat := "http://localhost:%d/%s/v1/events/subscribed"
			subscription := testingutils.NewSubscription(
				testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, tc.givenEventType),
			)
			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithSubscription(scheme, subscription, tc.givenEventTypePrefix),
				mock.WithApplication(testingutils.ApplicationName),
			)

			// nolint:scopelint
			// run the tests for subscribed endpoint
			for _, testCase := range handlertest.TestCasesForSubscribedEndpoint {
				t.Run(testCase.Name, func(t *testing.T) {
					subscribedURL := fmt.Sprintf(subscribedEndpointFormat, handlerMock.GetNatsConfig().Port, testCase.AppName)
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
						t.Errorf("failed to convert body to bytes: %v", err)
					}
					gotEventsResponse := subscribed.Events{}
					err = json.Unmarshal(respBodyBytes, &gotEventsResponse)
					if err != nil {
						t.Errorf("failed to unmarshal body bytes to events response: %v", err)
					}
					if !reflect.DeepEqual(testCase.WantResponse, gotEventsResponse) {
						t.Errorf("incorrect response, wanted: %v, got: %v", testCase.WantResponse, gotEventsResponse)
					}
				})
			}
		})
	}
}
