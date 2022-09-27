//go:build !skip
// +build !skip

package jetstreamv2

import (
	"reflect"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	cleaner2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/stretchr/testify/require"
)

// maxJetStreamConsumerNameLength is the maximum preferred length for the JetStream consumer names
// as per https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming
const maxJetStreamConsumerNameLength = 32

func TestGetCleanEventTypes(t *testing.T) {
	t.Parallel()
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	cleaner := cleaner2.NewJetStreamCleaner(defaultLogger)
	testCases := []struct {
		name              string
		givenSubscription *eventingv1alpha2.Subscription
		wantEventTypes    []eventingv1alpha2.EventType
		wantError         bool
	}{
		{
			name: "Should throw an error if the source is empty",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithEventType(evtestingv2.OrderCreatedUncleanEvent),
			),
			wantEventTypes: []eventingv1alpha2.EventType{},
			wantError:      true,
		},
		{
			name: "Should throw an error if the eventType is empty",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithEventSource(evtestingv2.EventSourceUnclean),
			),
			wantEventTypes: []eventingv1alpha2.EventType{},
			wantError:      true,
		},
		{
			name: "Should not clean eventTypes if the typeMatching is set to Exact",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithNotCleanEventSourceAndType(),
				evtestingv2.WithTypeMatchingExact(),
			),
			wantEventTypes: []eventingv1alpha2.EventType{
				{
					OriginalType: evtestingv2.OrderCreatedUncleanEvent,
					CleanType:    evtestingv2.OrderCreatedUncleanEvent,
				},
			},
			wantError: false,
		},
		{
			name: "Should clean eventTypes if the typeMatching is set to Standard",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithNotCleanEventSourceAndType(),
				evtestingv2.WithTypeMatchingStandard(),
			),
			wantEventTypes: []eventingv1alpha2.EventType{
				{
					OriginalType: evtestingv2.OrderCreatedUncleanEvent,
					CleanType:    evtestingv2.OrderCreatedCleanEvent,
				},
			},
			wantError: false,
		},
		{
			name: "Should clean multiple eventTypes",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithNotCleanEventSourceAndType(),
				evtestingv2.WithEventType(evtestingv2.OrderCreatedV1Event),
				evtestingv2.WithTypeMatchingStandard(),
			),
			wantEventTypes: []eventingv1alpha2.EventType{
				{
					OriginalType: evtestingv2.OrderCreatedUncleanEvent,
					CleanType:    evtestingv2.OrderCreatedCleanEvent,
				},
				{
					OriginalType: evtestingv2.OrderCreatedV1Event,
					CleanType:    evtestingv2.OrderCreatedV1Event,
				},
			},
			wantError: false,
		},
		{
			name: "Should throw an error for zero length - BadSubject",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithEventType(""),
				evtestingv2.WithTypeMatchingStandard(),
			),
			wantEventTypes: []eventingv1alpha2.EventType{},
			wantError:      true,
		},
		{
			name: "Should throw an error for less than two segments - BadSubject",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithEventType("order"),
				evtestingv2.WithTypeMatchingStandard(),
			),
			wantEventTypes: []eventingv1alpha2.EventType{},
			wantError:      true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eventTypes, err := getCleanEventTypes(tc.givenSubscription, cleaner)
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantEventTypes, eventTypes)
		})
	}
}

// TestSubscriptionSubjectIdentifierEqual checks the equality of two SubscriptionSubjectIdentifier instances and their consumer names.
func TestSubscriptionSubjectIdentifierEqual(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		givenIdentifier1 SubscriptionSubjectIdentifier
		givenIdentifier2 SubscriptionSubjectIdentifier
		wantEqual        bool
	}{
		// instances are equal
		{
			name: "should be equal if the two instances are identical",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: true,
		},
		// instances are not equal
		{
			name: "should not be equal if only name is different",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-2", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
		{
			name: "should not be equal if only namespace is different",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-2"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
		{
			name: "should not be equal if only subject is different",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v2",
			),
			wantEqual: false,
		},
		// possible naming collisions
		{
			name: "should not be equal if subject is the same but name and namespace are swapped",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("ns-1", "sub-1"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
		{
			name: "should not be equal if subject is the same but name and namespace are only equal if joined together",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1", "ns-1"), // evaluates to "sub-1ns-1" when joined
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub-1n", "s-1"), // evaluates to "sub-1ns-1" when joined
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotInstanceEqual := reflect.DeepEqual(tc.givenIdentifier1, tc.givenIdentifier2)
			require.Equal(t, tc.wantEqual, gotInstanceEqual)

			gotConsumerNameEqual := tc.givenIdentifier1.ConsumerName() == tc.givenIdentifier2.ConsumerName()
			require.Equal(t, tc.wantEqual, gotConsumerNameEqual)
		})
	}
}

// TestSubscriptionSubjectIdentifierConsumerNameLength checks that the SubscriptionSubjectIdentifier consumer name
// length is equal to the recommended length by JetStream.
func TestSubscriptionSubjectIdentifierConsumerNameLength(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                   string
		givenIdentifier        SubscriptionSubjectIdentifier
		wantConsumerNameLength int
	}{
		{
			name: "short string values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub", "ns"),
				"app.event.operation.v1",
			),
			wantConsumerNameLength: maxJetStreamConsumerNameLength,
		},
		{
			name: "long string values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("some-test-subscription", "some-test-namespace"),
				"some.test.prefix.some-test-application.some-test-event-name.some-test-operation.some-test-version",
			),
			wantConsumerNameLength: maxJetStreamConsumerNameLength,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantConsumerNameLength, len(tc.givenIdentifier.ConsumerName()))
		})
	}
}

// TestSubscriptionSubjectIdentifierNamespacedName checks the syntax of the SubscriptionSubjectIdentifier namespaced name.
func TestSubscriptionSubjectIdentifierNamespacedName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name               string
		givenIdentifier    SubscriptionSubjectIdentifier
		wantNamespacedName string
	}{
		{
			name: "short name and namespace values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("sub", "ns"),
				"app.event.operation.v1",
			),
			wantNamespacedName: "ns/sub",
		},
		{
			name: "long name and namespace values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtestingv2.NewSubscription("some-test-subscription", "some-test-namespace"),
				"app.event.operation.v1",
			),
			wantNamespacedName: "some-test-namespace/some-test-subscription",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantNamespacedName, tc.givenIdentifier.NamespacedName())
		})
	}
}

// TestJetStream_isJsSubAssociatedWithKymaSub tests the isJsSubAssociatedWithKymaSub method.
func TestJetStream_isJsSubAssociatedWithKymaSub(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create subscription 1 and its JetStream subscription
	cleanSubject1 := "subOne"
	sub1 := evtestingv2.NewSubscription(cleanSubject1, "foo", evtestingv2.WithNotCleanEventSourceAndType())
	jsSub1Key := NewSubscriptionSubjectIdentifier(sub1, cleanSubject1)

	// create subscription 2 and its JetStream subscription
	cleanSubject2 := "subOneTwo"
	sub2 := evtestingv2.NewSubscription(cleanSubject2, "foo", evtestingv2.WithNotCleanEventSourceAndType())
	jsSub2Key := NewSubscriptionSubjectIdentifier(sub2, cleanSubject2)

	testCases := []struct {
		name            string
		givenJSSubKey   SubscriptionSubjectIdentifier
		givenKymaSubKey *eventingv1alpha2.Subscription
		wantResult      bool
	}{
		{
			name:            "",
			givenJSSubKey:   jsSub1Key,
			givenKymaSubKey: sub1,
			wantResult:      true,
		},
		{
			name:            "",
			givenJSSubKey:   jsSub2Key,
			givenKymaSubKey: sub2,
			wantResult:      true,
		},
		{
			name:            "",
			givenJSSubKey:   jsSub1Key,
			givenKymaSubKey: sub2,
			wantResult:      false,
		},
		{
			name:            "",
			givenJSSubKey:   jsSub2Key,
			givenKymaSubKey: sub1,
			wantResult:      false,
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			gotResult := isJsSubAssociatedWithKymaSub(tC.givenJSSubKey, tC.givenKymaSubKey)
			require.Equal(t, tC.wantResult, gotResult)
		})
	}
}
