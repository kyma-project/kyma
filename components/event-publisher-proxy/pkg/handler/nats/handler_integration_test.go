package nats_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithApplication(tc.givenApplicationName),
			)
			defer handlerMock.ShutdownNatsServerAndWait()

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
					assert.NoError(t, err)
					_ = resp.Body.Close()
					assert.Equal(t, testCase.WantStatusCode, resp.StatusCode)
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestNatsHandlerForLegacyEvents(t *testing.T) {
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithApplication(tc.givenApplicationName),
			)
			defer handlerMock.ShutdownNatsServerAndWait()

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
					require.NoError(t, err)
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)

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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// setup test environment
			scheme := runtime.NewScheme()
			require.NoError(t, corev1.AddToScheme(scheme))
			require.NoError(t, eventingv1alpha1.AddToScheme(scheme))

			subscribedEndpointFormat := "http://localhost:%d/%s/v1/events/subscribed"
			subscription := testingutils.NewSubscription(
				testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, tc.givenEventType),
			)
			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithSubscription(scheme, subscription, tc.givenEventTypePrefix),
				mock.WithApplication(testingutils.ApplicationName),
			)
			defer handlerMock.ShutdownNatsServerAndWait()

			// nolint:scopelint
			// run the tests for subscribed endpoint
			for _, testCase := range handlertest.TestCasesForSubscribedEndpoint {
				t.Run(testCase.Name, func(t *testing.T) {
					subscribedURL := fmt.Sprintf(subscribedEndpointFormat, handlerMock.GetNatsConfig().Port, testCase.AppName)
					resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
					require.NoError(t, err)
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)
					defer func() { _ = resp.Body.Close() }()

					respBodyBytes, err := ioutil.ReadAll(resp.Body)
					require.NoError(t, err)

					gotEventsResponse := subscribed.Events{}
					err = json.Unmarshal(respBodyBytes, &gotEventsResponse)
					require.NoError(t, err)
					require.Equal(t, testCase.WantResponse, gotEventsResponse)
				})
			}
		})
	}
}
