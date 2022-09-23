package beb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/beb/mock"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	legacyapi "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

const (
	// mock server endpoints
	defaultEventsEndpoint        = "/events"
	defaultEventsHTTP400Endpoint = "/events400"

	// request size
	smallRequestSize = 2
	bigRequestSize   = 65536

	publishEndpointFormat       = "http://localhost:%d/publish"
	publishLegacyEndpointFormat = "http://localhost:%d/%s/v1/events"
	subscribedEndpointFormat    = "http://localhost:%d/%s/v1/events/subscribed"
)

func TestHandlerForCloudEvents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                           string
		givenEventTypePrefix           string
		givenEventType                 string
		givenApplicationNameToCreate   string
		givenApplicationNameToValidate string
	}{
		// not-clean event-types
		{
			name:                           "With prefix and clean application name and not-clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefix,
			givenEventType:                 testingutils.CloudEventTypeNotClean,
			givenApplicationNameToCreate:   testingutils.ApplicationName,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With empty prefix and clean application name and not-clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefixEmpty,
			givenEventType:                 testingutils.CloudEventTypeNotClean,
			givenApplicationNameToCreate:   testingutils.ApplicationName,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefix,
			givenEventType:                 testingutils.CloudEventTypeNotClean,
			givenApplicationNameToCreate:   testingutils.ApplicationNameNotClean,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With empty prefix and not-clean application name and not-clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefixEmpty,
			givenEventType:                 testingutils.CloudEventTypeNotClean,
			givenApplicationNameToCreate:   testingutils.ApplicationNameNotClean,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		// clean event-types
		{
			name:                           "With prefix and clean application name and clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefix,
			givenEventType:                 testingutils.CloudEventType,
			givenApplicationNameToCreate:   testingutils.ApplicationName,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With empty prefix and clean application name and clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefixEmpty,
			givenEventType:                 testingutils.CloudEventType,
			givenApplicationNameToCreate:   testingutils.ApplicationName,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With prefix and not-clean application name and clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefix,
			givenEventType:                 testingutils.CloudEventType,
			givenApplicationNameToCreate:   testingutils.ApplicationNameNotClean,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With empty prefix and not-clean application name and clean event-type",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefixEmpty,
			givenEventType:                 testingutils.CloudEventType,
			givenApplicationNameToCreate:   testingutils.ApplicationNameNotClean,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestSize        = bigRequestSize
				eventsEndpoint     = defaultEventsEndpoint
				requestTimeout     = time.Second
				serverResponseTime = time.Nanosecond
			)

			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, tc.givenEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(tc.givenApplicationNameToCreate, tc.givenApplicationNameToValidate),
			)
			defer handlerMock.Close()
			publishEndpoint := fmt.Sprintf(publishEndpointFormat, handlerMock.GetPort())

			for _, testCase := range handlertest.TestCasesForCloudEvents {
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					body, headers := testCase.ProvideMessage(tc.givenEventType)
					resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
					require.NoError(t, err)
					require.NoError(t, resp.Body.Close())
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestHandlerForLegacyEvents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                           string
		givenEventTypePrefix           string
		givenApplicationNameToCreate   string
		givenApplicationNameToValidate string
	}{
		{
			name:                           "With prefix and clean application name",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefix,
			givenApplicationNameToCreate:   testingutils.ApplicationName,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With empty prefix and clean application name",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationNameToCreate:   testingutils.ApplicationName,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With prefix and not-clean application name",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefix,
			givenApplicationNameToCreate:   testingutils.ApplicationNameNotClean,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
		{
			name:                           "With empty prefix and not-clean application name",
			givenEventTypePrefix:           testingutils.MessagingEventTypePrefixEmpty,
			givenApplicationNameToCreate:   testingutils.ApplicationNameNotClean,
			givenApplicationNameToValidate: testingutils.ApplicationName,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestSize        = bigRequestSize
				eventsEndpoint     = defaultEventsEndpoint
				requestTimeout     = time.Second
				serverResponseTime = time.Nanosecond
			)

			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, tc.givenEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(tc.givenApplicationNameToCreate, tc.givenApplicationNameToValidate),
			)
			defer handlerMock.Close()
			publishLegacyEndpoint := fmt.Sprintf(publishLegacyEndpointFormat, handlerMock.GetPort(), tc.givenApplicationNameToCreate)

			for _, testCase := range handlertest.TestCasesForLegacyEvents {
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					body, headers := testCase.ProvideMessage()
					resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
					require.NoError(t, err)
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)
					if testCase.WantStatusCode == http.StatusOK {
						handlertest.ValidateLegacyOkResponse(t, *resp, &testCase.WantResponse)
					} else {
						handlertest.ValidateLegacyErrorResponse(t, *resp, &testCase.WantResponse)
					}
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestHandlerForBEBFailures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		givenEventTypePrefix string
	}{
		{
			name:                 "With prefix",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
		},
		{
			name:                 "With empty prefix",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestSize        = bigRequestSize
				applicationName    = testingutils.ApplicationName
				eventsEndpoint     = defaultEventsHTTP400Endpoint
				requestTimeout     = time.Second
				serverResponseTime = time.Nanosecond
			)

			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, tc.givenEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(applicationName, applicationName),
			)
			defer handlerMock.Close()
			publishEndpoint := fmt.Sprintf(publishEndpointFormat, handlerMock.GetPort())
			publishLegacyEndpoint := fmt.Sprintf(publishLegacyEndpointFormat, handlerMock.GetPort(), applicationName)

			innerTestCases := []struct {
				name           string
				provideMessage func() (string, http.Header)
				givenEndpoint  string
				wantStatusCode int
				wantResponse   legacyapi.PublishEventResponses
			}{
				{
					name: "Send a legacy event with event-id",
					provideMessage: func() (string, http.Header) {
						builder := testingutils.NewLegacyEventBuilder()
						return builder.Build()
					},
					givenEndpoint:  publishLegacyEndpoint,
					wantStatusCode: http.StatusBadRequest,
					wantResponse: legacyapi.PublishEventResponses{
						Error: &legacyapi.Error{
							Status:  http.StatusBadRequest,
							Message: "invalid request"},
					},
				},
				{
					name: "Binary CloudEvent is valid with required headers",
					provideMessage: func() (string, http.Header) {
						return fmt.Sprintf(`"%s"`, testingutils.EventData), testingutils.GetBinaryMessageHeaders()
					},
					givenEndpoint:  publishEndpoint,
					wantStatusCode: http.StatusBadRequest,
				},
			}
			for _, testCase := range innerTestCases {
				testCase := testCase
				t.Run(testCase.name, func(t *testing.T) {
					body, headers := testCase.provideMessage()
					resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
					require.NoError(t, err)
					require.Equal(t, testCase.wantStatusCode, resp.StatusCode)
					if testCase.givenEndpoint == publishLegacyEndpoint {
						handlertest.ValidateLegacyErrorResponse(t, *resp, &testCase.wantResponse)
					}
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestHandlerForHugeRequests(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		givenEventTypePrefix string
	}{
		{
			name:                 "With prefix",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
		},
		{
			name:                 "With empty prefix",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestSize        = smallRequestSize
				applicationName    = testingutils.ApplicationName
				eventsEndpoint     = defaultEventsHTTP400Endpoint
				requestTimeout     = time.Second
				serverResponseTime = time.Nanosecond
			)

			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, tc.givenEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(applicationName, applicationName),
			)
			defer handlerMock.Close()
			publishLegacyEndpoint := fmt.Sprintf(publishLegacyEndpointFormat, handlerMock.GetPort(), applicationName)

			innerTestCases := []struct {
				name           string
				provideMessage func() (string, http.Header)
				givenEndpoint  string
				wantStatusCode int
			}{
				{
					name: "Should fail with HTTP 413 with a request which is larger than 2 Bytes as the maximum accepted size is 2 Bytes",
					provideMessage: func() (string, http.Header) {
						builder := testingutils.NewLegacyEventBuilder()
						return builder.Build()
					},
					givenEndpoint:  publishLegacyEndpoint,
					wantStatusCode: http.StatusRequestEntityTooLarge,
				},
				{
					name: "Should accept a request which is lesser than 2 Bytes as the maximum accepted size is 2 Bytes",
					provideMessage: func() (string, http.Header) {
						return "{}", testingutils.GetBinaryMessageHeaders()
					},
					givenEndpoint:  handler.PublishEndpoint,
					wantStatusCode: http.StatusBadRequest,
				},
			}
			for _, testCase := range innerTestCases {
				testCase := testCase
				t.Run(testCase.name, func(t *testing.T) {
					body, headers := testCase.provideMessage()
					resp, err := testingutils.SendEvent(publishLegacyEndpoint, body, headers)
					require.NoError(t, err)
					require.Equal(t, testCase.wantStatusCode, resp.StatusCode)
					if testingutils.Is2XX(resp.StatusCode) {
						metricstest.EnsureMetricLatency(t, handlerMock.GetMetricsCollector())
					}
				})
			}
		})
	}
}

func TestHandlerForSubscribedEndpoint(t *testing.T) {
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestSize              = smallRequestSize
				eventsEndpoint           = defaultEventsHTTP400Endpoint
				requestTimeout           = time.Second
				serverResponseTime       = time.Nanosecond
				subscribedEndpointFormat = subscribedEndpointFormat
			)

			scheme := runtime.NewScheme()
			require.NoError(t, corev1.AddToScheme(scheme))
			require.NoError(t, eventingv1alpha1.AddToScheme(scheme))
			subscription := testingutils.NewSubscription(testingutils.SubscriptionWithFilter(testingutils.MessagingNamespace, tc.givenEventType))

			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, tc.givenEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime,
				mock.WithSubscription(scheme, subscription),
			)
			defer handlerMock.Close()

			for _, testCase := range handlertest.TestCasesForSubscribedEndpoint {
				testCase := testCase
				t.Run(testCase.Name, func(t *testing.T) {
					subscribedURL := fmt.Sprintf(subscribedEndpointFormat, handlerMock.GetPort(), testCase.AppName)
					resp, err := testingutils.QuerySubscribedEndpoint(subscribedURL)
					require.NoError(t, err)
					require.Equal(t, testCase.WantStatusCode, resp.StatusCode)
					defer func() { require.NoError(t, resp.Body.Close()) }()
					respBodyBytes, err := io.ReadAll(resp.Body)
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

func TestHandlerTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		givenEventTypePrefix string
		givenEventType       string
	}{
		// not-clean event-types
		{
			name:                 "With prefix and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefix,
			givenEventType:       testingutils.CloudEventTypeNotClean,
		},
		{
			name:                 "With empty prefix and not-clean event-type",
			givenEventTypePrefix: testingutils.MessagingEventTypePrefixEmpty,
			givenEventType:       testingutils.CloudEventTypeNotCleanPrefixEmpty,
		},
		// clean event-types
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
			t.Parallel()

			var (
				requestSize        = bigRequestSize
				applicationName    = testingutils.ApplicationName
				eventsEndpoint     = defaultEventsHTTP400Endpoint
				requestTimeout     = time.Nanosecond  // short request timeout
				serverResponseTime = time.Millisecond // long server response time
			)

			handlerMock := mock.StartOrDie(context.TODO(), t, requestSize, tc.givenEventTypePrefix, eventsEndpoint, requestTimeout, serverResponseTime,
				mock.WithEventTypePrefix(tc.givenEventTypePrefix),
				mock.WithApplication(applicationName, applicationName),
			)
			defer handlerMock.Close()
			publishEndpoint := fmt.Sprintf(publishEndpointFormat, handlerMock.GetPort())

			builder := testingutils.NewCloudEventBuilder(
				testingutils.WithCloudEventType(tc.givenEventType),
			)
			body, headers := builder.BuildStructured()
			resp, err := testingutils.SendEvent(publishEndpoint, body, headers)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			metricstest.EnsureMetricErrors(t, handlerMock.GetMetricsCollector())
		})
	}
}
