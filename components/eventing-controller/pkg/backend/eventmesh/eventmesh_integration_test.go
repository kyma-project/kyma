package eventmesh

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/stretchr/testify/require"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbebv1 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	PublisherManagerMock "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	controllertestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
)

func Test_getProcessedEventTypes(t *testing.T) {

	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com",
		maxSubscriptionNameLength)

	// cases
	testCases := []struct {
		name                    string
		givenSubscription       *eventingv1alpha2.Subscription
		givenEventTypePrefix    string
		wantProcessedEventTypes []backendutils.EventTypeInfo
		wantError               bool
	}{
		{
			name: "success if the given subscription has already cleaned source and event types",
			givenSubscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Types: []string{
						"order.created.v1",
					},
					Source: "test",
				},
			},
			givenEventTypePrefix: controllertestingv2.EventMeshPrefix,
			wantProcessedEventTypes: []backendutils.EventTypeInfo{
				{
					OriginalType:  "order.created.v1",
					CleanType:     "order.created.v1",
					ProcessedType: fmt.Sprintf("%s.test.order.created.v1", controllertestingv2.EventMeshPrefix),
				},
			},
			wantError: false,
		},
		{
			name: "success if the given subscription has uncleaned source and event types",
			givenSubscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Types: []string{
						"Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
					},
					Source: "test-Ä.Segment2",
				},
			},
			givenEventTypePrefix: controllertestingv2.EventMeshPrefix,
			wantProcessedEventTypes: []backendutils.EventTypeInfo{
				{
					OriginalType: "Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
					CleanType:    "Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
					ProcessedType: fmt.Sprintf("%s.testSegment2.Segment1Segment2Segment3Segment4Part1Part2"+
						".Segment5Part1Part2.v1", controllertestingv2.EventMeshPrefix),
				},
			},
			wantError: false,
		},
		{
			name: "success if the given subscription has duplicate event types",
			givenSubscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Types: []string{
						"order.created.v1",
						"order.created.v1",
					},
					Source: "test",
				},
			},
			givenEventTypePrefix: controllertestingv2.EventMeshPrefix,
			wantProcessedEventTypes: []backendutils.EventTypeInfo{
				{
					OriginalType:  "order.created.v1",
					CleanType:     "order.created.v1",
					ProcessedType: fmt.Sprintf("%s.test.order.created.v1", controllertestingv2.EventMeshPrefix),
				},
			},
			wantError: false,
		},
		{
			name: "should not clean or process event type if the given subscription has matchingType=EXACT",
			givenSubscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Types: []string{
						"test1.test2.test3.order.created.v1",
					},
					Source:       "test",
					TypeMatching: eventingv1alpha2.TypeMatchingExact,
				},
			},
			givenEventTypePrefix: controllertestingv2.EventMeshPrefix,
			wantProcessedEventTypes: []backendutils.EventTypeInfo{
				{
					OriginalType:  "test1.test2.test3.order.created.v1",
					CleanType:     "test1.test2.test3.order.created.v1",
					ProcessedType: "test1.test2.test3.order.created.v1",
				},
			},
			wantError: false,
		},
		{
			name: "should fail if the given subscription types and EventMeshPrefix " +
				"exceeds the EventMesh segments limit",
			givenSubscription: &eventingv1alpha2.Subscription{
				Spec: eventingv1alpha2.SubscriptionSpec{
					Types: []string{
						"order.created.v1",
					},
					Source: "test",
				},
			},
			givenEventTypePrefix:    controllertestingv2.InvalidEventMeshPrefix,
			wantProcessedEventTypes: nil,
			wantError:               true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			eventMesh := NewEventMesh(&backendbebv1.OAuth2ClientCredentials{}, nameMapper, defaultLogger)
			emCleaner := cleaner.NewEventMeshCleaner(defaultLogger)
			err = eventMesh.Initialize(env.Config{EventTypePrefix: tc.givenEventTypePrefix})
			require.NoError(t, err)

			// when
			eventTypeInfos, err := eventMesh.getProcessedEventTypes(tc.givenSubscription, emCleaner)

			// then
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantProcessedEventTypes, eventTypeInfos)
			}
		})
	}

}

func Test_handleKymaSubModified(t *testing.T) {
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com",
		maxSubscriptionNameLength)

	// cases
	testCases := []struct {
		name                      string
		givenKymaSub              *eventingv1alpha2.Subscription
		givenEventMeshSub         *types.Subscription
		givenClientDeleteResponse *types.DeleteResponse
		wantIsModified            bool
		wantError                 bool
	}{
		{
			name: "should not delete EventMesh sub if Kyma sub was not modified",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Backend: eventingv1alpha2.Backend{
						Ev2hash: int64(-9219276050977208880),
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentMode",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientDeleteResponse: &types.DeleteResponse{
				Response: types.Response{
					StatusCode: http.StatusNoContent,
					Message:    "",
				},
			},
			wantIsModified: false,
			wantError:      false,
		},
		{
			name: "should delete EventMesh sub if Kyma sub was modified",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Backend: eventingv1alpha2.Backend{
						Ev2hash: int64(-9219276050977208880),
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentModeModified",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientDeleteResponse: &types.DeleteResponse{
				Response: types.Response{
					StatusCode: http.StatusNoContent,
					Message:    "",
				},
			},
			wantIsModified: true,
			wantError:      false,
		},
		{
			name: "fail if delete EventMesh sub return error",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Backend: eventingv1alpha2.Backend{
						Ev2hash: int64(-9219276050977208880),
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentModeModified",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientDeleteResponse: &types.DeleteResponse{
				Response: types.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "",
				},
			},
			wantIsModified: true,
			wantError:      true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			eventMesh := NewEventMesh(&backendbebv1.OAuth2ClientCredentials{}, nameMapper, defaultLogger)
			// Set a mock client interface for EventMesh
			mockClient := new(PublisherManagerMock.PublisherManager)
			mockClient.On("Delete", tc.givenEventMeshSub.Name).Return(tc.givenClientDeleteResponse, nil)
			eventMesh.client = mockClient

			// when
			isModified, err := eventMesh.handleKymaSubModified(tc.givenEventMeshSub, tc.givenKymaSub)

			// then
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantIsModified, isModified)
			}
		})
	}
}

func Test_handleEventMeshSubModified(t *testing.T) {
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com",
		maxSubscriptionNameLength)

	// cases
	testCases := []struct {
		name                      string
		givenKymaSub              *eventingv1alpha2.Subscription
		givenEventMeshSub         *types.Subscription
		givenClientDeleteResponse *types.DeleteResponse
		wantIsModified            bool
		wantError                 bool
	}{
		{
			name: "should not delete EventMesh sub if EventMesh sub was not modified",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Backend: eventingv1alpha2.Backend{
						Emshash: int64(-9219276050977208880),
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentMode",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientDeleteResponse: &types.DeleteResponse{
				Response: types.Response{
					StatusCode: http.StatusNoContent,
					Message:    "",
				},
			},
			wantIsModified: false,
			wantError:      false,
		},
		{
			name: "should delete EventMesh sub if EventMesh sub was modified",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Backend: eventingv1alpha2.Backend{
						Emshash: int64(-9219276050977208880),
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentModeModified",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientDeleteResponse: &types.DeleteResponse{
				Response: types.Response{
					StatusCode: http.StatusNoContent,
					Message:    "",
				},
			},
			wantIsModified: true,
			wantError:      false,
		},
		{
			name: "should fail if delete EventMesh sub return error",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Backend: eventingv1alpha2.Backend{
						Emshash: int64(-9219276050977208880),
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentModeModified",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientDeleteResponse: &types.DeleteResponse{
				Response: types.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "",
				},
			},
			wantIsModified: true,
			wantError:      true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			eventMesh := NewEventMesh(&backendbebv1.OAuth2ClientCredentials{}, nameMapper, defaultLogger)
			// Set a mock client interface for EventMesh
			mockClient := new(PublisherManagerMock.PublisherManager)
			mockClient.On("Delete", tc.givenEventMeshSub.Name).Return(tc.givenClientDeleteResponse, nil)
			eventMesh.client = mockClient

			// when
			isModified, err := eventMesh.handleEventMeshSubModified(tc.givenEventMeshSub, tc.givenKymaSub)

			// then
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantIsModified, isModified)
			}
		})
	}
}

func Test_handleCreateEventMeshSub(t *testing.T) {
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com",
		maxSubscriptionNameLength)

	// cases
	testCases := []struct {
		name                      string
		givenKymaSub              *eventingv1alpha2.Subscription
		givenEventMeshSub         *types.Subscription
		givenClientCreateResponse *types.CreateResponse
		wantError                 bool
	}{
		{
			name: "should be able create EventMesh sub",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Types: []eventingv1alpha2.EventType{
						{
							OriginalType: "test1",
							CleanType:    "test1",
						},
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentMode",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientCreateResponse: &types.CreateResponse{
				Response: types.Response{
					StatusCode: http.StatusAccepted,
					Message:    "",
				},
			},
			wantError: false,
		},
		{
			name: "should fail to create EventMesh sub if server returns error",
			givenKymaSub: &eventingv1alpha2.Subscription{
				Status: eventingv1alpha2.SubscriptionStatus{
					Types: []eventingv1alpha2.EventType{
						{
							OriginalType: "test1",
							CleanType:    "test1",
						},
					},
				},
			},
			givenEventMeshSub: &types.Subscription{
				Name:            "Name1",
				ContentMode:     "ContentModeModified",
				ExemptHandshake: true,
				Qos:             types.QosAtLeastOnce,
				WebhookURL:      "www.kyma-project.io",
			},
			givenClientCreateResponse: &types.CreateResponse{
				Response: types.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "",
				},
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			eventMesh := NewEventMesh(&backendbebv1.OAuth2ClientCredentials{}, nameMapper, defaultLogger)
			// Set a mock client interface for EventMesh
			mockClient := new(PublisherManagerMock.PublisherManager)
			mockClient.On("Create", tc.givenEventMeshSub).Return(tc.givenClientCreateResponse, nil)
			mockClient.On("Get", tc.givenEventMeshSub.Name).Return(tc.givenEventMeshSub, &types.Response{
				StatusCode: http.StatusOK,
				Message:    "",
			}, nil)
			eventMesh.client = mockClient

			// when
			_, err := eventMesh.handleCreateEventMeshSub(tc.givenEventMeshSub, tc.givenKymaSub)

			// then
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Empty(t, tc.givenKymaSub.Status.Types)
			}
		})
	}
}

func Test_handleKymaSubStatusUpdate(t *testing.T) {
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// cases
	testCases := []struct {
		name               string
		givenKymaSub       *eventingv1alpha2.Subscription
		givenEventMeshSub  *types.Subscription
		givenTypeInfos     []backendutils.EventTypeInfo
		wantEventTypes     []eventingv1alpha2.EventType
		wantEventMeshTypes []eventingv1alpha2.EventMeshTypes
	}{
		{
			name:         "should be able create EventMesh sub",
			givenKymaSub: &eventingv1alpha2.Subscription{},
			givenEventMeshSub: &types.Subscription{
				Name:                     "Name1",
				ContentMode:              "ContentMode",
				ExemptHandshake:          true,
				Qos:                      types.QosAtLeastOnce,
				WebhookURL:               "www.kyma-project.io",
				SubscriptionStatusReason: "test-reason",
			},
			givenTypeInfos: []backendutils.EventTypeInfo{
				{
					OriginalType:  "test1",
					CleanType:     "test2",
					ProcessedType: "test3",
				},
				{
					OriginalType:  "order1",
					CleanType:     "order2",
					ProcessedType: "order3",
				},
			},
			wantEventTypes: []eventingv1alpha2.EventType{
				{
					OriginalType: "test1",
					CleanType:    "test2",
				},
				{
					OriginalType: "order1",
					CleanType:    "order2",
				},
			},
			wantEventMeshTypes: []eventingv1alpha2.EventMeshTypes{
				{
					OriginalType:  "test1",
					EventMeshType: "test3",
				},
				{
					OriginalType:  "order1",
					EventMeshType: "order3",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// given
			eventMesh := NewEventMesh(nil, nil, defaultLogger)

			// when
			isChanged, _ := eventMesh.handleKymaSubStatusUpdate(tc.givenEventMeshSub, tc.givenEventMeshSub, tc.givenKymaSub, tc.givenTypeInfos)

			// then
			require.Equal(t, isChanged, true)
			require.Equal(t, tc.givenKymaSub.Status.Types, tc.wantEventTypes)
			require.Equal(t, tc.givenKymaSub.Status.Backend.EmsTypes, tc.wantEventMeshTypes)
			require.Equal(t, tc.givenKymaSub.Status.Backend.EmsSubscriptionStatus.StatusReason,
				tc.givenEventMeshSub.SubscriptionStatusReason)
		})
	}
}

func Test_SyncSubscription(t *testing.T) {
	credentials := &backendbebv1.OAuth2ClientCredentials{
		ClientID:     "foo-client-id",
		ClientSecret: "foo-client-secret",
	}
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com",
		maxSubscriptionNameLength)
	eventMesh := NewEventMesh(credentials, nameMapper, defaultLogger)

	// start BEB Mock
	eventMeshMock := startEventMeshMock()
	envConf := env.Config{
		BEBAPIURL:                eventMeshMock.MessagingURL,
		ClientID:                 "client-id",
		ClientSecret:             "client-secret",
		TokenEndpoint:            eventMeshMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookTokenEndpoint:     "webhook-token-endpoint",
		Domain:                   "domain.com",
		EventTypePrefix:          controllertestingv2.EventTypePrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      string(types.QosAtLeastOnce),
	}

	err = eventMesh.Initialize(envConf)
	require.NoError(t, err)

	subscription := fixtureValidSubscription("some-name", "some-namespace")
	subscription.Status.Backend.Emshash = 0
	subscription.Status.Backend.Ev2hash = 0

	apiRule := controllertestingv2.NewAPIRule(subscription,
		controllertestingv2.WithPath(),
		controllertestingv2.WithService("foo-svc", "foo-host"),
	)

	// cases - reconcile same subscription multiple times
	testCases := []struct {
		name           string
		givenEventType string
		wantIsChanged  bool
	}{
		{
			name:           "should be able to sync first time",
			givenEventType: controllertestingv2.OrderCreatedEventTypeNotClean,
			wantIsChanged:  true,
		},
		{
			name:           "should be able to sync second time with same subscription",
			givenEventType: controllertestingv2.OrderCreatedEventTypeNotClean,
			wantIsChanged:  false,
		},
		{
			name:           "should be able to sync third time with modified subscription",
			givenEventType: controllertestingv2.OrderCreatedV2Event,
			wantIsChanged:  true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// when
			subscription.Spec.Types[0] = tc.givenEventType
			changed, err := eventMesh.SyncSubscription(subscription, cleaner.NewEventMeshCleaner(defaultLogger), apiRule)
			require.NoError(t, err)
			require.Equal(t, tc.wantIsChanged, changed)
		})
	}

	// cleanup
	eventMeshMock.Stop()
}

// fixtureValidSubscription returns a valid subscription.
func fixtureValidSubscription(name, namespace string) *eventingv1alpha2.Subscription {
	return controllertestingv2.NewSubscription(
		name, namespace,
		controllertestingv2.WithSinkURL("https://webhook.xxx.com"),
		controllertestingv2.WithDefaultSource(),
		controllertestingv2.WithEventType(controllertestingv2.OrderCreatedEventTypeNotClean),
		controllertestingv2.WithWebhookAuthForBEB(),
	)
}

func startEventMeshMock() *controllertesting.BEBMock {
	eventMesh := controllertesting.NewBEBMock()
	eventMesh.Start()
	return eventMesh
}
