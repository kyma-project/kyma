package eventmesh

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/stretchr/testify/require"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbebv1 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	controllertestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
)

func Test_GetProcessedEventTypes(t *testing.T) {

	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com", MaxEventMeshSubscriptionNameLength)

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
					OriginalType:  "Segment1.Segment2.Segment3.Segment4-Part1-Part2-Ä.Segment5-Part1-Part2-Ä.v1",
					CleanType:     "Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1",
					ProcessedType: fmt.Sprintf("%s.testSegment2.Segment1Segment2Segment3Segment4Part1Part2.Segment5Part1Part2.v1", controllertestingv2.EventMeshPrefix),
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
					TypeMatching: eventingv1alpha2.EXACT,
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
			name: "should fail if the given subscription types and EventMeshPrefix exceeds the EventMesh segments limit",
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
			cleaner := cleaner.NewEventMeshCleaner(defaultLogger)
			err = eventMesh.Initialize(env.Config{EventTypePrefix: tc.givenEventTypePrefix})
			require.NoError(t, err)

			// when
			eventTypeInfos, err := eventMesh.GetProcessedEventTypes(tc.givenSubscription, cleaner)

			// then
			require.Equal(t, tc.wantError, err != nil)
			if !tc.wantError {
				require.Equal(t, tc.wantProcessedEventTypes, eventTypeInfos)
			}
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

	nameMapper := backendutils.NewBEBSubscriptionNameMapper("mydomain.com", MaxEventMeshSubscriptionNameLength)
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

	// when
	subscription := fixtureValidSubscription("some-name", "some-namespace")
	subscription.Status.Backend.Emshash = 0
	subscription.Status.Backend.Ev2hash = 0

	apiRule := controllertestingv2.NewAPIRule(subscription,
		controllertestingv2.WithPath(),
		controllertestingv2.WithService("foo-svc", "foo-host"),
	)

	// then
	changed, err := eventMesh.SyncSubscription(subscription, cleaner.NewEventMeshCleaner(defaultLogger), apiRule)
	require.NoError(t, err)
	require.True(t, changed)
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
