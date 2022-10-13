package jetstreamv2

import (
	"sync"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	backenderrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2/errors"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/mocks"
	subtesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test_SyncNATSConsumersAndSubscriptions_ForEmptyTypes tests the subscription without any EventType.
func Test_SyncNATSConsumersAndSubscriptions_ForEmptyTypes(t *testing.T) {
	// given
	callback := func(m *nats.Msg) {}
	subWithOneType := NewSubscriptionWithEmptyTypes()

	// when
	js := JetStream{}
	err := js.syncNATSConsumersAndSubscriptions(subWithOneType, callback)

	// then
	assert.NoError(t, err)
}

// Test_GetOrCreateConsumer tests the behaviour of the getOrCreateConsumer function.
func Test_GetOrCreateConsumer(t *testing.T) {
	existingConsumer := &nats.ConsumerInfo{Name: "ExistingConsumer", Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}
	newConsumer := &nats.ConsumerInfo{Name: "NewConsumer", Config: nats.ConsumerConfig{MaxAckPending: 20}}
	sub := NewSubscriptionWithOneType()

	testCases := []struct {
		name             string
		jetStreamContext *jetStreamContextStub
		wantError        error
		wantConsumerInfo *nats.ConsumerInfo
	}{
		{
			name: "existing consumer should be returned",
			jetStreamContext: &jetStreamContextStub{
				consumerInfoError: nil,
				consumerInfo:      existingConsumer,
			},
			wantConsumerInfo: existingConsumer,
			wantError:        nil,
		},
		{
			name: "new consumer should created when there is no existing one",
			jetStreamContext: &jetStreamContextStub{
				consumerInfoError: nats.ErrConsumerNotFound,
				consumerInfo:      nil,

				addConsumer:      newConsumer,
				addConsumerError: nil,
			},
			wantConsumerInfo: newConsumer,
			wantError:        nil,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			js := JetStream{
				subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
				metricsCollector: metrics.NewCollector(),
				jsCtx:            *testCase.jetStreamContext,
				cleaner:          &cleaner.JetStreamCleaner{},
			}
			eventType := sub.Status.Types[0]
			jsSubject := js.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
			jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

			maxInFlight, err := sub.GetMaxInFlightMessages()
			assert.NoError(t, err)

			// when
			consumerInfo, err := js.getOrCreateConsumer(jsSubject, jsSubKey, maxInFlight)

			// then
			assert.Equal(t, tc.wantConsumerInfo, consumerInfo)
			assert.ErrorIs(t, tc.wantError, err)
		})
	}
}

// Test_SyncNATSConsumersAndSubscriptions_ForGetMaxInFlight test for valid/invalid maxInFlight values in the subscription.
func Test_SyncNATSConsumersAndSubscriptions_ForGetMaxInFlight(t *testing.T) {
	// given
	callback := func(m *nats.Msg) {}

	testCases := []struct {
		name             string
		jetStreamContext *jetStreamContextStub
		givenMaxInFlight string
		wantErr          error
	}{
		{
			name:             "invalid maxInFlight should return an error",
			givenMaxInFlight: "nonInt",
			jetStreamContext: &jetStreamContextStub{},
			wantErr:          &backenderrors.FailedToReadConfigError{},
		},
		{
			name:             "invalid maxInFlight should return an error",
			givenMaxInFlight: "20",
			jetStreamContext: &jetStreamContextStub{
				consumerInfoError: nil,
				consumerInfo:      &nats.ConsumerInfo{},
				subscribe:         &nats.Subscription{},
			},
			wantErr: nil,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			sub := subtesting.NewSubscription("test", "test",
				subtesting.WithStatusTypes([]v1alpha2.EventType{
					{
						OriginalType: subtesting.CloudEventType,
						CleanType:    subtesting.CloudEventType,
					},
				}),
				subtesting.WithMaxInFlightMessages(testCase.givenMaxInFlight),
			)
			// given

			js := JetStream{
				subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
				metricsCollector: metrics.NewCollector(),
				jsCtx:            *testCase.jetStreamContext,
				cleaner:          &cleaner.JetStreamCleaner{},
			}
			err := js.syncNATSConsumersAndSubscriptions(sub, callback)
			require.ErrorIs(t, err, testCase.wantErr)
		})
	}

}

// Test_SyncNATSConsumersAndSubscriptions_ForBindInvalidSubscriptions tests the binding behaviour in the syncNATSConsumersAndSubscriptions function.
func Test_SyncNATSConsumersAndSubscriptions_ForBindInvalidSubscriptions(t *testing.T) {
	sub := NewSubscriptionWithOneType()
	validSubMock := &mocks.Subscriber{}
	validSubMock.On("IsValid").Return(true)
	invalidSubMock := &mocks.Subscriber{}
	invalidSubMock.On("IsValid").Return(false)

	testCases := []struct {
		name       string
		wantError  error
		givenMocks func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier)
	}{
		{
			name: "Bind invalid NATS Subscription should result into error",
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: invalidSubMock,
				}
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					// mock the expected calls
					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(&nats.ConsumerInfo{}, nil)
					jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).Return(nil, nats.ErrJetStreamNotEnabled)
				}
			},
			wantError: &backenderrors.FailedToSubscribeOnNATSError{OriginalError: nats.ErrJetStreamNotEnabled},
		},
		{
			name: "Bind invalid NATS Subscription should succeed",
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: invalidSubMock,
				}
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					// mock the expected calls
					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(&nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}, nil)
					jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).Return(&nats.Subscription{}, nil)
				}
			},
			wantError: nil,
		},
		{
			name: "Skip binding for when no invalid NATS Subscriptions",
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: validSubMock,
				}
				// mock the expected calls
				jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(&nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}, nil)
			},
			wantError: nil,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			callback := func(m *nats.Msg) {}
			js, jsCtxMock := setupJetStreamBackend()

			// setup the mocks
			eventType := sub.Status.Types[0]
			jsSubject := js.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
			jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
			tc.givenMocks(sub, js, jsCtxMock, jsSubKey)

			// when
			err := js.syncNATSConsumersAndSubscriptions(sub, callback)

			// then
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			}
			jsCtxMock.AssertExpectations(t)
		})
	}

}

// Test_SyncSyncConsumerMaxInFlight tests the behaviour of the syncConsumerMaxInFlight function.
func Test_SyncSyncConsumerMaxInFlight(t *testing.T) {
	testCases := []struct {
		name                       string
		givenSubMaxInFlight        int
		givenConsumerMaxAckPending int
		givenMocks                 func(jsBackend *JetStream, jsCtx *mocks.JetStreamContext)
	}{
		{
			name:                       "up-to-date consumer shouldn't be updated",
			givenSubMaxInFlight:        DefaultMaxInFlights,
			givenConsumerMaxAckPending: DefaultMaxInFlights,
			// no updateConsumer calls expected
			givenMocks: func(jsBackend *JetStream, jsCtx *mocks.JetStreamContext) {},
		},
		{
			name:                       "up-to-date consumer should be updated",
			givenSubMaxInFlight:        10,
			givenConsumerMaxAckPending: 20,
			givenMocks: func(jsBackend *JetStream, jsCtx *mocks.JetStreamContext) {
				jsCtx.On("UpdateConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(&nats.ConsumerInfo{
					Config: nats.ConsumerConfig{MaxAckPending: 20},
				}, nil)
			},
		},
		{
			name:                       "up-to-date consumer should be updated",
			givenSubMaxInFlight:        10,
			givenConsumerMaxAckPending: 20,
			givenMocks: func(jsBackend *JetStream, jsCtx *mocks.JetStreamContext) {
				jsCtx.On("UpdateConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(&nats.ConsumerInfo{
					Config: nats.ConsumerConfig{MaxAckPending: 20},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			js, jsCtxMock := setupJetStreamBackend()

			// setup the mocks
			consumer := nats.ConsumerInfo{
				Name:   "name",
				Config: nats.ConsumerConfig{MaxAckPending: tc.givenConsumerMaxAckPending},
			}
			tc.givenMocks(js, jsCtxMock)

			// when
			err := js.syncConsumerMaxInFlight(consumer, tc.givenSubMaxInFlight)

			// then
			assert.NoError(t, err)
			jsCtxMock.AssertExpectations(t)
		})
	}
}

// Test_SyncNATSConsumers_ForErrors test the syncNATSConsumersAndSubscriptions for right error handling.
func Test_SyncNATSConsumers_ForErrors(t *testing.T) {
	// given
	callback := func(m *nats.Msg) {}
	subWithOneType := NewSubscriptionWithOneType()
	jsCleaner := &cleaner.JetStreamCleaner{}

	testCases := []struct {
		name             string
		jetStreamContext *jetStreamContextStub
		wantError        error
	}{
		{
			name: "ConsumerInfo's not found error should be ignored",
			jetStreamContext: &jetStreamContextStub{
				consumerInfoError: nats.ErrConsumerNotFound,
				consumerInfo:      nil,

				addConsumer: &nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}},

				subscribe: &nats.Subscription{},
			},
			wantError: nil,
		},
		{
			name: "ConsumerInfo's error should be propagated",
			jetStreamContext: &jetStreamContextStub{
				consumerInfoError: nats.ErrStreamNotFound,
				consumerInfo:      nil,
			},
			wantError: &backenderrors.FailedToFetchConsumerInfoError{OriginalError: nats.ErrStreamNotFound},
		},
		{
			name: "AddConsumer's error should be propagated",
			jetStreamContext: &jetStreamContextStub{
				consumerInfo:      nil,
				consumerInfoError: nats.ErrConsumerNotFound,

				addConsumer:      nil,
				addConsumerError: nats.ErrJetStreamNotEnabledForAccount,
			},
			wantError: &backenderrors.FailedToAddConsumerError{OriginalError: nats.ErrJetStreamNotEnabledForAccount},
		},
		{
			name: "Subscribe call on createNATSSubscription error should be propagated",
			jetStreamContext: &jetStreamContextStub{
				consumerInfo:      &nats.ConsumerInfo{},
				consumerInfoError: nil,

				subscribe:      nil,
				subscribeError: nats.ErrJetStreamNotEnabled,
			},
			wantError: &backenderrors.FailedToSubscribeOnNATSError{OriginalError: nats.ErrJetStreamNotEnabled},
		},
		{
			name: "UpdateConsumer call error should be propagated",
			jetStreamContext: &jetStreamContextStub{
				consumerInfo:      &nats.ConsumerInfo{},
				consumerInfoError: nil,

				subscribe:      &nats.Subscription{},
				subscribeError: nil,

				update:      nil,
				updateError: nats.ErrJetStreamNotEnabled,
			},
			wantError: &backenderrors.FailedToUpdateConsumerInfoError{OriginalError: nats.ErrJetStreamNotEnabled},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			js := JetStream{
				subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
				metricsCollector: metrics.NewCollector(),
				jsCtx:            *testCase.jetStreamContext,
				cleaner:          jsCleaner,
			}

			// when
			err := js.syncNATSConsumersAndSubscriptions(subWithOneType, callback)

			// then
			assert.ErrorIs(t, err, testCase.wantError)

		})
	}
}

// Test_CheckNATSSubscriptionsCount tests the behaviour of the checkNATSSubscriptionsCount function.
func Test_CheckNATSSubscriptionsCount(t *testing.T) {
	// given
	subWithType := NewSubscriptionWithOneType()
	jsBackend := &JetStream{
		subscriptions: make(map[SubscriptionSubjectIdentifier]Subscriber),
		cleaner:       &cleaner.JetStreamCleaner{},
	}

	testCases := []struct {
		name                 string
		givenSubscription    *v1alpha2.Subscription
		givenSubscriptionMap map[SubscriptionSubjectIdentifier]Subscriber
		wantErr              error
	}{
		{
			name:                 "empty subscriptions map with subscription with no types should result in not error",
			givenSubscription:    NewSubscriptionWithEmptyTypes(),
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{},
			wantErr:              nil,
		},
		{
			name:              "if the subscriptions map contains all the NATS Subscriptions, no error is expected",
			givenSubscription: subWithType,
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{
				NewSubscriptionSubjectIdentifier(subWithType, "kyma./default/kyma/id.prefix.testapp1023.order.created.v1"): &nats.Subscription{},
			},
			wantErr: nil,
		},
		{
			name:                 "unexpected empty subscriptions map should result into an error",
			givenSubscription:    subWithType,
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{},
			wantErr:              &backenderrors.MissingSubscriptionError{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {

			// inject the fake subscription map
			jsBackend.subscriptions = testCase.givenSubscriptionMap

			// when
			err := jsBackend.checkNATSSubscriptionsCount(testCase.givenSubscription)

			// then
			if testCase.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, testCase.wantErr)
			}
		})
	}

}

// HELPER FUNCTION

// sets up the setupJetStreamBackend with the mocks instead of running a full-fledged NATS test server.
func setupJetStreamBackend() (*JetStream, *mocks.JetStreamContext) {
	jsCtx := &mocks.JetStreamContext{}
	natsConfig := env.NatsConfig{}
	metricsCollector := metrics.NewCollector()

	return &JetStream{
		Config:           natsConfig,
		Conn:             nil,
		jsCtx:            jsCtx,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
		sinks:            sync.Map{},
		metricsCollector: metricsCollector,
		cleaner:          &cleaner.JetStreamCleaner{},
	}, jsCtx
}

func NewSubscription(name, namespace string, opts ...subtesting.SubscriptionOpt) *v1alpha2.Subscription {
	newSub := &v1alpha2.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha2.SubscriptionSpec{},
	}
	for _, o := range opts {
		o(newSub)
	}
	return newSub
}

func NewSubscriptionWithEmptyTypes() *v1alpha2.Subscription {
	return NewSubscription("test", "test",
		subtesting.WithStatusTypes(nil),
	)
}

func NewSubscriptionWithOneType() *v1alpha2.Subscription {
	return NewSubscription("test", "test",
		subtesting.WithSourceAndType(subtesting.EventSource, subtesting.CloudEventType),
		subtesting.WithTypeMatchingStandard(),
		subtesting.WithMaxInFlight(env.DefaultMaxInFlight),
		subtesting.WithStatusTypes([]v1alpha2.EventType{
			{
				OriginalType: subtesting.CloudEventType,
				CleanType:    subtesting.CloudEventType,
			},
		}),
	)
}
