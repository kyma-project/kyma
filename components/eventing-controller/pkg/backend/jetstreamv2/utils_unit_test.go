package jetstreamv2

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/stretchr/testify/require"
)

// maxJetStreamConsumerNameLength is the maximum preferred length for the JetStream consumer names
// as per https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming
const (
	subName                        = "subName"
	subNamespace                   = "subNamespace"
	maxJetStreamConsumerNameLength = 32
)

func TestToJetStreamStorageType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		givenStorageType string
		wantStorageType  nats.StorageType
		wantError        bool
	}{
		{
			name:             "memory storage type",
			givenStorageType: StorageTypeMemory,
			wantStorageType:  nats.MemoryStorage,
			wantError:        false,
		},
		{
			name:             "file storage type",
			givenStorageType: StorageTypeFile,
			wantStorageType:  nats.FileStorage,
			wantError:        false,
		},
		{
			name:             "invalid storage type",
			givenStorageType: "",
			wantStorageType:  nats.MemoryStorage,
			wantError:        true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			storageType, err := toJetStreamStorageType(tc.givenStorageType)
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantStorageType, storageType)
		})
	}
}

func TestToJetStreamRetentionPolicy(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                 string
		givenRetentionPolicy string
		wantSRetentionPolicy nats.RetentionPolicy
		wantError            bool
	}{
		{
			name:                 "retention policy limits",
			givenRetentionPolicy: RetentionPolicyLimits,
			wantSRetentionPolicy: nats.LimitsPolicy,
			wantError:            false,
		},
		{
			name:                 "retention policy interest",
			givenRetentionPolicy: RetentionPolicyInterest,
			wantSRetentionPolicy: nats.InterestPolicy,
			wantError:            false,
		},
		{
			name:                 "invalid retention policy",
			givenRetentionPolicy: "",
			wantSRetentionPolicy: nats.LimitsPolicy,
			wantError:            true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			retentionPolicy, err := toJetStreamRetentionPolicy(tc.givenRetentionPolicy)
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantSRetentionPolicy, retentionPolicy)
		})
	}
}

func TestGetStreamConfig(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		givenNatsConfig  env.NatsConfig
		wantStreamConfig *nats.StreamConfig
		wantError        bool
	}{
		{
			name: "Should throw an error if storage type is invalid",
			givenNatsConfig: env.NatsConfig{
				JSStreamStorageType: "invalid",
			},
			wantStreamConfig: nil,
			wantError:        true,
		},
		{
			name: "Should throw an error if retention policy is invalid",
			givenNatsConfig: env.NatsConfig{
				JSStreamRetentionPolicy: "invalid",
			},
			wantStreamConfig: nil,
			wantError:        true,
		},
		{
			name: "Should return valid StreamConfig",
			givenNatsConfig: env.NatsConfig{
				JSStreamName:            DefaultStreamName,
				JSStreamStorageType:     StorageTypeMemory,
				JSStreamRetentionPolicy: RetentionPolicyLimits,
				JSStreamReplicas:        3,
				JSStreamMaxMessages:     -1,
				JSStreamMaxBytes:        -1,
			},
			wantStreamConfig: &nats.StreamConfig{
				Name:      DefaultStreamName,
				Storage:   nats.MemoryStorage,
				Replicas:  3,
				Retention: nats.LimitsPolicy,
				MaxMsgs:   -1,
				MaxBytes:  -1,
				Subjects:  []string{fmt.Sprintf("%s.>", env.JetStreamSubjectPrefix)},
			},
			wantError: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			streamConfig, err := getStreamConfig(tc.givenNatsConfig)
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantStreamConfig, streamConfig)
		})
	}
}

func TestCreateKeyPrefix(t *testing.T) {
	// given
	sub := evtestingv2.NewSubscription(subName, subNamespace)
	// when
	keyPrefix := createKeyPrefix(sub)
	// then
	require.Equal(t, keyPrefix, fmt.Sprintf("%s/%s", subNamespace, subName))
}

func TestGetCleanEventTypesFromStatus(t *testing.T) {
	// given
	sub := evtestingv2.NewSubscription(subName, subNamespace)
	sub.Status.Types = []eventingv1alpha2.EventType{
		{
			OriginalType: evtestingv2.OrderCreatedUncleanEvent,
			CleanType:    evtestingv2.OrderCreatedCleanEvent,
		},
		{
			OriginalType: evtestingv2.OrderCreatedEventTypeNotClean,
			CleanType:    evtestingv2.OrderCreatedEventType,
		},
	}
	// when
	cleanTypes := GetCleanEventTypesFromEventTypes(sub.Status.Types)
	// then
	require.Equal(t, cleanTypes, []string{evtestingv2.OrderCreatedCleanEvent, evtestingv2.OrderCreatedEventType})
}

func TestGetCleanEventTypes(t *testing.T) {
	t.Parallel()
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	jscleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	testCases := []struct {
		name              string
		givenSubscription *eventingv1alpha2.Subscription
		wantEventTypes    []eventingv1alpha2.EventType
		wantError         bool
	}{
		{
			name: "Should throw an error if the eventType is empty",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithEventSource(evtestingv2.EventSourceUnclean),
			),
			wantEventTypes: []eventingv1alpha2.EventType{},
			wantError:      true,
		},
		{
			name: "Should throw an error if the eventType is empty",
			givenSubscription: evtestingv2.NewSubscription("sub", "test",
				evtestingv2.WithEventSource(evtestingv2.EventSourceUnclean),
				evtestingv2.WithEventType(""),
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
			eventTypes, getCleanTypesErr := GetCleanEventTypes(tc.givenSubscription, jscleaner)
			require.Equal(t, tc.wantError, getCleanTypesErr != nil)
			require.Equal(t, tc.wantEventTypes, eventTypes)
		})
	}
}

func TestGetBackendJetStreamTypes(t *testing.T) {
	t.Parallel()
	jsCleaner := cleaner.NewJetStreamCleaner(nil)
	defaultSub := evtestingv2.NewSubscription(subName, subNamespace)
	js := NewJetStream(env.NatsConfig{}, nil, jsCleaner, env.DefaultSubscriptionConfig{}, nil)
	testCases := []struct {
		name              string
		givenSubscription *eventingv1alpha2.Subscription
		givenJSSubjects   []string
		wantJSTypes       []eventingv1alpha2.JetStreamTypes
	}{
		{
			name:              "Should be nil is there are no jsSubjects",
			givenSubscription: defaultSub,
			givenJSSubjects:   []string{},
			wantJSTypes:       nil,
		},
		{
			name: "one type and one jsSubject",
			givenSubscription: evtestingv2.NewSubscription(subName, subNamespace,
				evtestingv2.WithSource(evtestingv2.EventSourceUnclean),
				evtestingv2.WithEventType(evtestingv2.OrderCreatedUncleanEvent)),
			givenJSSubjects: js.GetJetStreamSubjects(evtestingv2.EventSourceUnclean,
				[]string{evtestingv2.OrderCreatedCleanEvent},
				eventingv1alpha2.TypeMatchingStandard),
			wantJSTypes: []eventingv1alpha2.JetStreamTypes{
				{
					OriginalType: evtestingv2.OrderCreatedUncleanEvent,
					ConsumerName: "828ed501e743dfc43e2f23cfc14b0232",
				},
			},
		},
		{
			name: "two types and two jsSubjects",
			givenSubscription: evtestingv2.NewSubscription(subName, subNamespace,
				evtestingv2.WithSource(evtestingv2.EventSourceUnclean),
				evtestingv2.WithEventType(evtestingv2.OrderCreatedCleanEvent),
				evtestingv2.WithEventType(evtestingv2.OrderCreatedV1Event)),
			givenJSSubjects: js.GetJetStreamSubjects(evtestingv2.EventSourceUnclean,
				[]string{evtestingv2.OrderCreatedCleanEvent, evtestingv2.OrderCreatedV1Event},
				eventingv1alpha2.TypeMatchingStandard),
			wantJSTypes: []eventingv1alpha2.JetStreamTypes{
				{
					OriginalType: evtestingv2.OrderCreatedCleanEvent,
					ConsumerName: "828ed501e743dfc43e2f23cfc14b0232",
				},
				{
					OriginalType: evtestingv2.OrderCreatedV1Event,
					ConsumerName: "ec2f903b07de7a974cf97c3d61fb043f",
				},
			},
		},
		{
			name: "two types and two jsSubjects with type matching exact",
			givenSubscription: evtestingv2.NewSubscription(subName, subNamespace,
				evtestingv2.WithSource(evtestingv2.EventSourceUnclean),
				evtestingv2.WithEventType(evtestingv2.OrderCreatedCleanEvent),
				evtestingv2.WithEventType(evtestingv2.OrderCreatedV1Event)),
			givenJSSubjects: js.GetJetStreamSubjects(evtestingv2.EventSourceUnclean,
				[]string{evtestingv2.OrderCreatedCleanEvent, evtestingv2.OrderCreatedV1Event},
				eventingv1alpha2.TypeMatchingExact),
			wantJSTypes: []eventingv1alpha2.JetStreamTypes{
				{
					OriginalType: evtestingv2.OrderCreatedCleanEvent,
					ConsumerName: "015e691825a7813383a419a53d8c5ea0",
				},
				{
					OriginalType: evtestingv2.OrderCreatedV1Event,
					ConsumerName: "15b59df6dc97f232718e05d7087c7a50",
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jsTypes := GetBackendJetStreamTypes(tc.givenSubscription, tc.givenJSSubjects)
			require.Equal(t, tc.wantJSTypes, jsTypes)
		})
	}
}

func TestCleanEventType(t *testing.T) {
	t.Parallel()
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	testCases := []struct {
		name           string
		givenEventType string
		wantEventType  string
		wantError      bool
	}{
		{
			name:           "Should throw an error if the event type is of length zero",
			givenEventType: "",
			wantEventType:  "",
			wantError:      true,
		},
		{
			name:           "Should throw an error if segments are less then two",
			givenEventType: "onesegment",
			wantEventType:  "",
			wantError:      true,
		},
		{
			name:           "Should return valid cleaned eventType",
			givenEventType: evtestingv2.OrderCreatedUncleanEvent,
			wantEventType:  evtestingv2.OrderCreatedCleanEvent,
			wantError:      false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cleanEventType, getTypesErr := getCleanEventType(tc.givenEventType, jsCleaner)
			require.Equal(t, tc.wantError, getTypesErr != nil)
			require.Equal(t, tc.wantEventType, cleanEventType)
		})
	}
}

// TestSubscriptionSubjectIdentifierEqual checks the equality of two
// SubscriptionSubjectIdentifier instances and their consumer names.
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

// TestSubscriptionSubjectIdentifierNamespacedName checks
// the syntax of the SubscriptionSubjectIdentifier namespaced name.
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
	t.Parallel()
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
			t.Parallel()
			gotResult := isJsSubAssociatedWithKymaSub(tC.givenJSSubKey, tC.givenKymaSubKey)
			require.Equal(t, tC.wantResult, gotResult)
		})
	}
}
