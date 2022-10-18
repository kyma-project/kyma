package jetstreamv2 //nolint:nolintlint,testpackage

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
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

// Test_SyncConsumersAndSubscriptions_ForEmptyTypes tests the subscription without any EventType.
func Test_SyncConsumersAndSubscriptions_ForEmptyTypes(t *testing.T) {
	// given
	callback := func(m *nats.Msg) {}
	subWithOneType := NewSubscriptionWithEmptyTypes()

	// when
	js := JetStream{}
	err := js.syncConsumersAndSubscriptions(subWithOneType, callback)

	// then
	assert.NoError(t, err)
}

// Test_GetOrCreateConsumer tests the behaviour of the getOrCreateConsumer function.
func Test_GetOrCreateConsumer(t *testing.T) {
	// pre-requisites
	existingConsumer := &nats.ConsumerInfo{
		Name: "ExistingConsumer",
		Config: nats.ConsumerConfig{
			MaxAckPending: DefaultMaxInFlights,
		},
	}
	newConsumer := &nats.ConsumerInfo{Name: "NewConsumer", Config: nats.ConsumerConfig{MaxAckPending: 20}}

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
				subscriptions: make(map[SubscriptionSubjectIdentifier]Subscriber),
				jsCtx:         *testCase.jetStreamContext,
				cleaner:       &cleaner.JetStreamCleaner{},
			}
			sub := NewSubscriptionWithOneType()
			eventType := sub.Status.Types[0]

			// when
			consumerInfo, err := js.getOrCreateConsumer(sub, eventType)

			// then
			assert.Equal(t, tc.wantConsumerInfo, consumerInfo)
			assert.ErrorIs(t, tc.wantError, err)
		})
	}
}

// Test_SyncConsumersAndSubscriptions_ForGetMaxInFlight test for valid/invalid maxInFlight values in the subscription.
func Test_SyncConsumersAndSubscriptions_ForGetMaxInFlight(t *testing.T) {
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
			wantErr:          ErrInvalidMaxInFlight,
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
			// given
			callback := func(m *nats.Msg) {}
			sub := NewSubscriptionWithOneType()
			sub.Spec.Config = map[string]string{
				v1alpha2.MaxInFlightMessages: testCase.givenMaxInFlight,
			}
			js := JetStream{
				subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
				metricsCollector: metrics.NewCollector(),
				jsCtx:            *testCase.jetStreamContext,
				cleaner:          &cleaner.JetStreamCleaner{},
			}
			// when
			err := js.syncConsumersAndSubscriptions(sub, callback)

			// then
			require.ErrorIs(t, err, testCase.wantErr)
		})
	}
}

// Test_SyncConsumersAndSubscriptions_ForBindInvalidSubscriptions tests the
// binding behaviour in the syncConsumersAndSubscriptions function.
func Test_SyncConsumersAndSubscriptions_ForBindInvalidSubscriptions(t *testing.T) {
	// pre-requisites
	subWithOneType := NewSubscriptionWithOneType()
	validSubscriber := &subscriberStub{isValid: true}
	invalidSubscriber := &subscriberStub{isValid: false}

	testCases := []struct {
		name       string
		wantError  error
		givenMocks func(sub *v1alpha2.Subscription,
			jsBackend *JetStream,
			jsCtx *mocks.JetStreamContext,
			jsSubKey SubscriptionSubjectIdentifier,
		)
	}{
		{
			name: "Bind invalid NATS Subscription should result into error",
			givenMocks: func(sub *v1alpha2.Subscription,
				jsBackend *JetStream,
				jsCtx *mocks.JetStreamContext,
				jsSubKey SubscriptionSubjectIdentifier,
			) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: invalidSubscriber,
				}
				eventType := sub.Status.Types[0]
				jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
				// mock the expected calls
				jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(&nats.ConsumerInfo{}, nil)
				jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).
					Return(nil, nats.ErrJetStreamNotEnabled)
			},
			wantError: ErrFailedSubscribe,
		},
		{
			name: "Bind invalid NATS Subscription should succeed",
			givenMocks: func(sub *v1alpha2.Subscription,
				jsBackend *JetStream,
				jsCtx *mocks.JetStreamContext,
				jsSubKey SubscriptionSubjectIdentifier,
			) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: invalidSubscriber,
				}
				eventType := sub.Status.Types[0]
				jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
				// mock the expected calls
				jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).
					Return(&nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}, nil)
				jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).
					Return(&nats.Subscription{}, nil)
			},
			wantError: nil,
		},
		{
			name: "Skip binding for when no invalid NATS Subscriptions",
			givenMocks: func(sub *v1alpha2.Subscription,
				jsBackend *JetStream,
				jsCtx *mocks.JetStreamContext,
				jsSubKey SubscriptionSubjectIdentifier,
			) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: validSubscriber,
				}
				// mock the expected calls
				jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).
					Return(&nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}, nil)
			},
			wantError: nil,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			callback := func(m *nats.Msg) {}
			jsCtxMock := &mocks.JetStreamContext{}
			js := &JetStream{
				jsCtx:   jsCtxMock,
				cleaner: &cleaner.JetStreamCleaner{},
			}

			// setup the mocks
			eventType := subWithOneType.Status.Types[0]
			jsSubject := js.getJetStreamSubject(subWithOneType.Spec.Source,
				eventType.CleanType,
				subWithOneType.Spec.TypeMatching,
			)
			jsSubKey := NewSubscriptionSubjectIdentifier(subWithOneType, jsSubject)
			tc.givenMocks(subWithOneType, js, jsCtxMock, jsSubKey)

			// when
			err := js.syncConsumersAndSubscriptions(subWithOneType, callback)

			// then
			assert.ErrorIs(t, err, tc.wantError)
			jsCtxMock.AssertExpectations(t)
		})
	}
}

// Test_SyncConsumersAndSubscriptions_ForSyncConsumerMaxInFlight tests
// the behaviour of the syncConsumerMaxInFlight function.
func Test_SyncConsumersAndSubscriptions_ForSyncConsumerMaxInFlight(t *testing.T) {
	testCases := []struct {
		name                       string
		givenSubMaxInFlight        int
		givenConsumerMaxAckPending int
		givenMocks                 func(jsBackend *JetStream,
			jsCtx *mocks.JetStreamContext,
			consumerConfigToUpdate *nats.ConsumerConfig,
		)
		wantConfigToUpdate *nats.ConsumerConfig
	}{
		{
			name:                       "up-to-date consumer shouldn't be updated",
			givenSubMaxInFlight:        DefaultMaxInFlights,
			givenConsumerMaxAckPending: DefaultMaxInFlights,
			// no updateConsumer calls expected
			givenMocks: func(jsBackend *JetStream, jsCtx *mocks.JetStreamContext, consumerConfigToUpdate *nats.ConsumerConfig) {
			},
			wantConfigToUpdate: nil,
		},
		{
			name:                       "non-up-to-date consumer should be updated with the expected MaxAckPending value",
			givenSubMaxInFlight:        10,
			givenConsumerMaxAckPending: 20,
			givenMocks: func(jsBackend *JetStream, jsCtx *mocks.JetStreamContext, consumerConfigToUpdate *nats.ConsumerConfig) {
				jsCtx.On("UpdateConsumer", jsBackend.Config.JSStreamName, consumerConfigToUpdate).Return(&nats.ConsumerInfo{
					Config: *consumerConfigToUpdate,
				}, nil)
			},
			wantConfigToUpdate: &nats.ConsumerConfig{MaxAckPending: 10},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			jsCtxMock := &mocks.JetStreamContext{}
			js := &JetStream{
				jsCtx: jsCtxMock,
			}
			sub := subtesting.NewSubscription("test", "test",
				subtesting.WithMaxInFlightMessages(fmt.Sprint(tc.givenSubMaxInFlight)),
			)

			// setup the mocks
			consumer := nats.ConsumerInfo{
				Name:   "name",
				Config: nats.ConsumerConfig{MaxAckPending: tc.givenConsumerMaxAckPending},
			}
			tc.givenMocks(js, jsCtxMock, tc.wantConfigToUpdate)

			// when
			err := js.syncConsumerMaxInFlight(sub, consumer)

			// then
			assert.NoError(t, err)
			jsCtxMock.AssertExpectations(t)
		})
	}
}

// Test_SyncConsumersAndSubscriptions_ForErrors test the syncConsumersAndSubscriptions for right error handling.
func Test_SyncConsumersAndSubscriptions_ForErrors(t *testing.T) {
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
			wantError: ErrGetConsumer,
		},
		{
			name: "AddConsumer's error should be propagated",
			jetStreamContext: &jetStreamContextStub{
				consumerInfo:      nil,
				consumerInfoError: nats.ErrConsumerNotFound,

				addConsumer:      nil,
				addConsumerError: nats.ErrJetStreamNotEnabledForAccount,
			},
			wantError: ErrAddConsumer,
		},
		{
			name: "Subscribe call on createNATSSubscription error should be propagated",
			jetStreamContext: &jetStreamContextStub{
				consumerInfo:      &nats.ConsumerInfo{},
				consumerInfoError: nil,

				subscribe:      nil,
				subscribeError: nats.ErrJetStreamNotEnabled,
			},
			wantError: ErrFailedSubscribe,
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
			wantError: ErrUpdateConsumer,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			callback := func(m *nats.Msg) {}
			subWithOneType := NewSubscriptionWithOneType()
			js := JetStream{
				subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
				metricsCollector: metrics.NewCollector(),
				jsCtx:            *testCase.jetStreamContext,
				cleaner:          &cleaner.JetStreamCleaner{},
			}

			// when
			err := js.syncConsumersAndSubscriptions(subWithOneType, callback)

			// then
			assert.ErrorIs(t, err, testCase.wantError)
		})
	}
}

// Test_SyncConsumersAndSubscriptions_ForBoundConsumerWithoutSubscription tests the scenario
// then the consumer reports, that it is bound to a NATS subscription even though it is not.
func Test_SyncConsumersAndSubscriptions_ForBoundConsumerWithoutSubscription(t *testing.T) {
	// given
	js := &JetStream{
		jsCtx: &jetStreamContextStub{
			consumerInfoError: nil,
			consumerInfo:      &nats.ConsumerInfo{PushBound: true},
		},
		cleaner: &cleaner.JetStreamCleaner{},
	}
	subWithType := NewSubscriptionWithOneType()
	callback := func(m *nats.Msg) {}

	// when
	err := js.syncConsumersAndSubscriptions(subWithType, callback)

	// then
	require.ErrorIs(t, err, ErrMissingSubscription)
}

// HELPER FUNCTIONS

func NewSubscriptionWithEmptyTypes() *v1alpha2.Subscription {
	return subtesting.NewSubscription("test", "test",
		subtesting.WithStatusTypes(nil),
	)
}

func NewSubscriptionWithOneType() *v1alpha2.Subscription {
	return subtesting.NewSubscription("test", "test",
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
