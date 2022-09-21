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

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats/mock"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestHandlerForCloudEvents(t *testing.T) {
	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenApplicationName string
		givenEventType       string
		wantEventType        string
	}{
		// not-clean event-types
		{
			name:                 "With prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
		{
			name:                 "With empty prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
		// clean event-types
		{
			name:                 "With prefix and clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventType,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With prefix and not-clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventType,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypePrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
		{
			name:                 "With empty prefix and not-clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypePrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(tc.givenApplicationName),
			)
			defer handlerMock.Stop()

			// run the tests for publishing cloudevents
			publishEndpoint := fmt.Sprintf("http://localhost:%d/publish", handlerMock.GetNATSConfig().Port)
			for _, testCase := range handlertest.TestCasesForCloudEvents {
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					// connect to nats
					connection, err := testingutils.ConnectToNATSServer(handlerMock.GetNATSURL())
					assert.Nil(t, err)
					assert.NotNil(t, connection)
					defer connection.Close()

					// validator to check NATS received events
					notify := make(chan bool)
					defer close(notify)
					validator := testingutils.ValidateNATSSubjectOrFail(t, tc.wantEventType, notify)
					testingutils.SubscribeToEventOrFail(t, connection, tc.wantEventType, validator)

					body, headers := testCase.ProvideMessage(tc.wantEventType)
					resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
					assert.NoError(t, err)
					assert.NoError(t, resp.Body.Close())
					assert.Equal(t, testCase.WantStatusCode, resp.StatusCode)
					if testingutils.IsNot4XX(resp.StatusCode) {
						metricstest.EnsureMetricEventTypePublished(t, handlerMock.GetMetricsCollector())
					}
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
						assert.NoError(t, testingutils.WaitForChannelOrTimeout(notify, time.Second*3))
					}
				})
			}
		})
	}
}

func TestHandlerForLegacyEvents(t *testing.T) {
	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenApplicationName string
		givenEventType       string
		wantEventType        string
	}{
		// not-clean event-types
		{
			name:                 "With prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
		{
			name:                 "With prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotClean,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
		// clean event-types
		{
			name:                 "With prefix and clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventType,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationName,
			givenEventType:       testingutils.CloudEventTypePrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
		{
			name:                 "With prefix and not-clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventType,
			wantEventType:        testingutils.CloudEventType,
		},
		{
			name:                 "With empty prefix and not-clean application name and clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationName: testingutils.ApplicationNameNotClean,
			givenEventType:       testingutils.CloudEventTypePrefixEmpty,
			wantEventType:        testingutils.CloudEventTypePrefixEmpty,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handlerMock := mock.StartOrDie(ctx, t,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(tc.givenApplicationName),
			)
			defer handlerMock.Stop()

			// run the tests for publishing legacy events
			publishLegacyEndpoint := fmt.Sprintf("http://localhost:%d/%s/v1/events", handlerMock.GetNATSConfig().Port, tc.givenApplicationName)
			for _, testCase := range handlertest.TestCasesForLegacyEvents {
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					// connect to nats
					connection, err := testingutils.ConnectToNATSServer(handlerMock.GetNATSURL())
					assert.Nil(t, err)
					assert.NotNil(t, connection)
					defer connection.Close()

					// publish a message to NATS and validate it
					notify := make(chan bool)
					defer close(notify)
					validator := testingutils.ValidateNATSSubjectOrFail(t, tc.wantEventType, notify)
					testingutils.SubscribeToEventOrFail(t, connection, tc.wantEventType, validator)

					body, headers := testCase.ProvideMessage()
					resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
					require.NoError(t, err)
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)

					if testCase.WantStatusCode == http.StatusOK {
						handlertest.ValidateLegacyOkResponse(t, *resp, &testCase.WantResponse)
					} else {
						handlertest.ValidateLegacyErrorResponse(t, *resp, &testCase.WantResponse)
					}
					if testingutils.IsNot4XX(resp.StatusCode) {
						metricstest.EnsureMetricEventTypePublished(t, handlerMock.GetMetricsCollector())
					}
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
						assert.NoError(t, testingutils.WaitForChannelOrTimeout(notify, time.Second*10))
					}
				})
			}
		})
	}
}

func TestHandlerForSubscribedEndpoint(t *testing.T) {
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
		tc := tc
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
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithSubscription(scheme, subscription),
				mock.WithApplication(testingutils.ApplicationName),
			)
			defer handlerMock.Stop()

			for _, testCase := range handlertest.TestCasesForSubscribedEndpoint {
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					subscribedURL := fmt.Sprintf(subscribedEndpointFormat, handlerMock.GetNATSConfig().Port, testCase.AppName)
					resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
					require.NoError(t, err)
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)

					respBodyBytes, err := ioutil.ReadAll(resp.Body)
					require.NoError(t, err)
					require.NoError(t, resp.Body.Close())

					gotEventsResponse := subscribed.Events{}
					err = json.Unmarshal(respBodyBytes, &gotEventsResponse)
					require.NoError(t, err)
					require.Equal(t, testCase.WantResponse, gotEventsResponse)
				})
			}
		})
	}
}
