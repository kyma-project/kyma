//go:build unit
// +build unit

package jetstreamv2

import (
	"context"
	"errors"
	"fmt"
	"testing"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	jetstreamv2mocks "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2/mocks"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	subtesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// //////////////////////////////////////////////////////////////////////
// JetStream.Initialize()
// /////////////////////////////////////////////////////////////////////

func Test_Initialize_HappyPath(t *testing.T) {
	// SuT: JetStream
	// UoW: Initialize()
	// test kind: state + interaction test

	// given
	connClosedHandler := func(message *nats.Conn) {}
	config := fixtureNatsConfig()
	streamInfo := fixtureNatsConfigAsStreamInfo()
	ceClient := &ceClientStub{}
	jetStreamContext := &jetStreamContextStub{streamInfo: streamInfo}

	connectionMock := &jetstreamv2mocks.ConnectionInterface{}
	connectionMock.On("SetClosedHandler", mock.Anything)
	connectionMock.On("SetReconnectHandler", mock.Anything)
	connectionMock.On("JetStream").Return(jetStreamContext, nil)

	js := JetStream{
		config:            config,
		connectionBuilder: &connectionBuilderStub{conn: connectionMock},
		ceClientFactory:   ceClientFactoryStub{client: ceClient},
		logger:            &loggerStub{t},
	}

	// when
	err := js.Initialize(connClosedHandler)

	// then
	assert.NoError(t, err)

	// state checks
	assert.Equal(t, js.connection, connectionMock)
	assert.Equal(t, js.jsCtx, jetStreamContext)
	assert.Equal(t, js.ceClient, ceClient)

	// ensure connection handlers are set
	connectionMock.AssertExpectations(t)
}

func Test_Initialize_ForInitNATSConn_Error(t *testing.T) {
	// SuT: JetStream
	// UoW: Initialize()
	// test kind: value test (checking the error code)
	// given
	connClosedHandler := func(message *nats.Conn) {}
	config := fixtureNatsConfig()
	givenErr := errors.New("build failed")
	jetStream := JetStream{
		// needs to be uninitialized so that the JetStream struct will do the initialization
		connection:        nil,
		config:            config,
		connectionBuilder: &connectionBuilderStub{err: givenErr},
	}

	// when
	err := jetStream.Initialize(connClosedHandler)

	// then
	assert.ErrorIs(t, err, givenErr)
}

func Test_Initialize_ForInitJSContext_BuildError(t *testing.T) {
	// SuT: JetStream
	// UoW: Initialize()
	// test kind: value test

	// given
	connClosedHandler := func(message *nats.Conn) {}
	config := fixtureNatsConfig()

	jetStream := JetStream{
		connectionBuilder: &connectionBuilderStub{conn: connectionStub{isConnected: false,
			jetStreamContextError: errors.New("build error")}},
		config: config,
	}

	// when
	err := jetStream.Initialize(connClosedHandler)

	// then
	assert.ErrorIs(t, err, ErrContext)
}

func Test_Initialize_ForInitCloudEventCloud_ReturnsError(t *testing.T) {
	// SuT: JetStream
	// UoW: Initialize()
	// test kind: value test

	// given
	connClosedHandler := func(message *nats.Conn) {}
	config := fixtureNatsConfig()

	connection := connectionStub{isConnected: false, jetStreamContext: &jetStreamContextStub{}}
	jetStream := JetStream{
		logger:            &loggerStub{t},
		connectionBuilder: &connectionBuilderStub{conn: connection},
		config:            config,
		ceClientFactory:   ceClientFactoryStub{err: errors.New("ceClient create error")},
	}

	// when
	err := jetStream.Initialize(connClosedHandler)

	// then
	assert.ErrorIs(t, err, ErrCEClient)
}

// Test_Initialize_ForEnsureStreamExists_ReturnsError ensures that an error from ensureStreamExists is propagated.
func Test_Initialize_ForEnsureStreamExists_ReturnsError(t *testing.T) {
	// SuT: JetStream
	// UoW: Initialize()
	// test kind: value test

	// given
	connClosedHandler := func(message *nats.Conn) {}
	config := fixtureNatsConfig()
	jsCtx := &jetStreamContextStub{streamInfoError: errors.New("stream info error")}
	connection := connectionStub{isConnected: false, jetStreamContext: jsCtx}
	jetStream := JetStream{
		connectionBuilder: &connectionBuilderStub{conn: connection},
		config:            config,
		ceClientFactory:   ceClientFactoryStub{},
	}

	// when
	err := jetStream.Initialize(connClosedHandler)

	assert.ErrorIs(t, err, ErrUnknown)
}

// Test_Initialize_ForEnsureCorrectStreamConfiguration_ReturnsError ensures that an error
// from ensureCorrectStreamConfiguration is propagated.
func Test_Initialize_ForEnsureCorrectStreamConfiguration_ReturnsError(t *testing.T) {
	// SuT: JetStream
	// UoW: Initialize()
	// test kind: value test

	// given
	connClosedHandler := func(message *nats.Conn) {}

	config := backendnats.Config{
		JSStreamName: "some-name",

		// new config has interest retention policy
		JSStreamRetentionPolicy: RetentionPolicyInterest,

		JSStreamStorageType:   StorageTypeMemory,
		JSStreamDiscardPolicy: DiscardPolicyNew,
	}
	// stream config still has limits policy
	streamConfigOld := nats.StreamConfig{Name: "some-name", Retention: nats.LimitsPolicy}

	jetStreamContext := &jetStreamContextStub{
		streamInfo:            &nats.StreamInfo{Config: streamConfigOld},
		updateStreamInfoError: errors.New("error while updating stream"),
	}
	connection := connectionStub{isConnected: false, jetStreamContext: jetStreamContext}
	jetStream := JetStream{
		logger:            loggerStub{t},
		connectionBuilder: &connectionBuilderStub{conn: connection},
		config:            config,
		ceClientFactory:   ceClientFactoryStub{},
		jsCtx:             jetStreamContext,
	}

	// when
	err := jetStream.Initialize(connClosedHandler)

	// then
	assert.ErrorIs(t, err, ErrUpdateStreamConfig)
}

// //////////////////////////////////////////////////////////////////////
// JetStream.ensureStreamExists()
// /////////////////////////////////////////////////////////////////////

// Test_EnsureStreamExists_HappyPath tests that if the stream exists, it's information are returned.
func Test_EnsureStreamExists_HappyPath(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureStreamExists()
	// test kind: interaction test

	// given
	jsCtx := jetstreamv2mocks.JetStreamContext{}
	config := fixtureNatsConfig()
	givenStreamInfo := fixtureNatsConfigAsStreamInfo()
	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  &jsCtx,
	}
	jsCtx.On("StreamInfo", config.JSStreamName).Return(givenStreamInfo, nil)

	// when
	streamInfo, streamConfig, err := js.ensureStreamExists()

	// then
	assert.NoError(t, err)
	assert.Equal(t, givenStreamInfo, streamInfo)
	assert.NotNil(t, streamConfig)
	jsCtx.AssertExpectations(t)
}

// Test_EnsureStreamExists_ForStreamInfo_ReturnsErrStreamNotFound ensures that a stream is created
// if it does not exist already.
func Test_EnsureStreamExists_ForStreamInfo_ReturnsErrStreamNotFound(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureStreamExists()
	// test kind: interaction test

	// given
	jsCtx := jetstreamv2mocks.JetStreamContext{}
	config := fixtureNatsConfig()
	streamConfig, err := convertNatsConfigToStreamConfig(config)
	require.NoError(t, err)

	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  &jsCtx,
	}
	jsCtx.On("StreamInfo", config.JSStreamName).Return(nil, nats.ErrStreamNotFound)
	jsCtx.On("AddStream", streamConfig).Return(nil, nil)

	// when
	_, _, err = js.ensureStreamExists()

	// then
	assert.NoError(t, err)
	jsCtx.AssertExpectations(t)
}

func Test_EnsureStreamExists_ForConvertNatsConfigToStreamConfig_ReturnsError(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureStreamExists()
	// test kind: value test

	// given
	config := backendnats.Config{
		JSStreamName:        "not-empty",
		JSStreamStorageType: "invalid field",
	}
	js := JetStream{
		logger: &loggerStub{t},
		config: config,
	}

	// when
	_, _, err := js.ensureStreamExists()

	// then
	assert.ErrorIs(t, err, ErrConfig)
}

func Test_EnsureStreamExists_ForStreamAdd_ReturnsError(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureStreamExists()
	// test kind: value test

	// given
	jsCtx := jetStreamContextStub{
		// if the stream cannot be retrieved, a new stream will be added
		streamInfoError: nats.ErrStreamNotFound,
		addStreamError:  errors.New("error while adding stream"),
	}
	config := fixtureNatsConfig()

	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  &jsCtx,
	}

	// when
	_, _, err := js.ensureStreamExists()

	// then
	assert.ErrorIs(t, err, ErrAddStream)
}

func Test_EnsureStreamExists_ForStreamInfo_ReturnsRandomError(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureStreamExists()
	// test kind: value test

	// given
	config := fixtureNatsConfig()
	natsError := errors.New("some nats error other than ErrStreamNotFound")
	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  &jetStreamContextStub{streamInfoError: natsError},
	}

	// when
	_, _, err := js.ensureStreamExists()

	// then
	assert.ErrorIs(t, err, ErrUnknown)
}

// //////////////////////////////////////////////////////////////////////
// JetStream.ensureCorrectStreamConfiguration()
// /////////////////////////////////////////////////////////////////////

// Test_EnsureStreamExists_HappyPath ensures that if there is no change in the stream configuration,
// then there is no update performed on the JetStream side.
func Test_EnsureCorrectStreamConfiguration_HappyPath(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureCorrectStreamConfiguration()
	// test kind: interaction
	// given
	config := fixtureNatsConfig()
	jsCtx := &jetstreamv2mocks.JetStreamContext{}
	jsCtx.On("UpdateStream").Return(nil, nil)
	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  jsCtx,
	}

	// both use the same config
	streamConfig := nats.StreamConfig{Name: "some-name"}
	streamInfo := &nats.StreamInfo{Config: streamConfig}

	// when
	err := js.ensureCorrectStreamConfiguration(streamInfo, &streamConfig)

	// then
	assert.NoError(t, err)
	// There should be no updateConsumer of the steam config (because there is no change)
	jsCtx.AssertNumberOfCalls(t, "UpdateStream", 0)
}

// Test_EnsureCorrectStreamConfiguration_StreamConfigChanged_HappyPath ensures that if there is a change
// in the stream configuration, the new configuration is updated on the stream.
func Test_EnsureCorrectStreamConfiguration_StreamConfigChanged_HappyPath(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureCorrectStreamConfiguration()
	// test kind: interaction
	// given
	config := fixtureNatsConfig()
	jsCtx := &jetstreamv2mocks.JetStreamContext{}
	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  jsCtx,
	}

	streamConfigOld := nats.StreamConfig{Name: "some-name", Retention: nats.LimitsPolicy}
	streamConfigNew := &nats.StreamConfig{Name: "some-name", Retention: nats.InterestPolicy}

	// ensure that the new config is updated on jetstream side (first argument)
	jsCtx.On("UpdateStream", streamConfigNew).Return(nil, nil)

	// simulate that stream has old config
	streamInfo := &nats.StreamInfo{Config: streamConfigOld}

	// when
	err := js.ensureCorrectStreamConfiguration(streamInfo, streamConfigNew)

	// then
	assert.NoError(t, err)
	// There should an updateConsumer since there is a new steam config (change of retention)
	jsCtx.AssertExpectations(t)
}

// Test_EnsureCorrectStreamConfiguration_StreamConfigChanged_ForUpdateStream_ReturnsError ensures that the correct
// error is returned for the case that UpdateStream returns an error.
func Test_EnsureCorrectStreamConfiguration_StreamConfigChanged_ForUpdateStream_ReturnsError(t *testing.T) {
	// SuT: JetStream
	// UoW: ensureCorrectStreamConfiguration()
	// test kind: interaction
	// given
	config := fixtureNatsConfig()
	streamConfigOld := nats.StreamConfig{Name: "some-name", Retention: nats.LimitsPolicy}
	streamConfigNew := &nats.StreamConfig{Name: "some-name", Retention: nats.InterestPolicy}

	js := JetStream{
		logger: &loggerStub{t},
		config: config,
		jsCtx:  &jetStreamContextStub{updateStreamInfoError: fmt.Errorf("update error")},
	}

	// simulate that stream has old config
	oldStreamInfo := &nats.StreamInfo{Config: streamConfigOld}

	// when
	err := js.ensureCorrectStreamConfiguration(oldStreamInfo, streamConfigNew)

	// then
	assert.ErrorIs(t, err, ErrUpdateStreamConfig)
}

// Test_SyncConsumersAndSubscriptions_ForEmptyTypes tests the subscription without any EventType.
func Test_SyncConsumersAndSubscriptions_ForEmptyTypes(t *testing.T) {
	// given
	callback := func(m *nats.Msg) {}
	subWithOneType := NewSubscriptionWithEmptyTypes()

	// when
	js := JetStream{}
	err := js.syncConsumerAndSubscription(subWithOneType, callback)

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
				jsCtx:         testCase.jetStreamContext,
				cleaner:       &cleaner.JetStreamCleaner{},
				config:        backendnats.Config{},
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

// Test_SyncConsumersAndSubscriptions_ForBindInvalidSubscriptions tests the
// binding behaviour in the syncConsumerAndSubscription function.
func Test_SyncConsumersAndSubscriptions_ForBindInvalidSubscriptions(t *testing.T) {
	// pre-requisites
	subWithOneType := NewSubscriptionWithOneType()
	validSubscriber := &subscriberStub{isValid: true}
	invalidSubscriber := &subscriberStub{isValid: false}

	testCases := []struct {
		name  string
		mocks func(sub *v1alpha2.Subscription,
			jsBackend *JetStream,
			jsCtx *jetstreamv2mocks.JetStreamContext,
			jsSubKey SubscriptionSubjectIdentifier,
		)
	}{
		{
			name: "Bind invalid NATS Subscription should succeed",
			mocks: func(sub *v1alpha2.Subscription,
				jsBackend *JetStream,
				jsCtx *jetstreamv2mocks.JetStreamContext,
				jsSubKey SubscriptionSubjectIdentifier,
			) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: invalidSubscriber,
				}
				eventType := sub.Status.Types[0]
				jsSubject := jsBackend.GetJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
				// mock the expected calls
				config := jsBackend.GetConfig()
				jsCtx.On("ConsumerInfo", config.JSStreamName, jsSubKey.ConsumerName()).
					Return(&nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}, nil)
				jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).
					Return(&nats.Subscription{}, nil)
			},
		},
		{
			name: "Skip binding for when no invalid NATS Subscriptions",
			mocks: func(sub *v1alpha2.Subscription,
				jsBackend *JetStream,
				jsCtx *jetstreamv2mocks.JetStreamContext,
				jsSubKey SubscriptionSubjectIdentifier,
			) {
				// inject the subscriptions map
				jsBackend.subscriptions = map[SubscriptionSubjectIdentifier]Subscriber{
					jsSubKey: validSubscriber,
				}
				// mock the expected calls
				config := jsBackend.GetConfig()
				jsCtx.On("ConsumerInfo", config.JSStreamName, jsSubKey.ConsumerName()).
					Return(&nats.ConsumerInfo{Config: nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights}}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			callback := func(m *nats.Msg) {}
			jsCtxMock := &jetstreamv2mocks.JetStreamContext{}
			js := &JetStream{
				jsCtx:   jsCtxMock,
				cleaner: &cleaner.JetStreamCleaner{},
			}

			// setup the jetstreamv2mocks
			eventType := subWithOneType.Status.Types[0]
			jsSubject := js.GetJetStreamSubject(subWithOneType.Spec.Source,
				eventType.CleanType,
				subWithOneType.Spec.TypeMatching,
			)
			jsSubKey := NewSubscriptionSubjectIdentifier(subWithOneType, jsSubject)
			tc.mocks(subWithOneType, js, jsCtxMock, jsSubKey)

			// when
			assert.NoError(t, js.syncConsumerAndSubscription(subWithOneType, callback))

			// then
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
		givenjetstreamv2mocks      func(jsBackend *JetStream,
			jsCtx *jetstreamv2mocks.JetStreamContext,
			consumerConfigToUpdate *nats.ConsumerConfig,
		)
		wantConfigToUpdate *nats.ConsumerConfig
	}{
		{
			name:                       "up-to-date consumer shouldn't be updated",
			givenSubMaxInFlight:        DefaultMaxInFlights,
			givenConsumerMaxAckPending: DefaultMaxInFlights,
			// no updateConsumer calls expected
			givenjetstreamv2mocks: func(jsBackend *JetStream,
				jsCtx *jetstreamv2mocks.JetStreamContext,
				consumerConfigToUpdate *nats.ConsumerConfig) {
			},
			wantConfigToUpdate: nil,
		},
		{
			name:                       "non-up-to-date consumer should be updated with the expected MaxAckPending value",
			givenSubMaxInFlight:        10,
			givenConsumerMaxAckPending: 20,
			givenjetstreamv2mocks: func(jsBackend *JetStream,
				jsCtx *jetstreamv2mocks.JetStreamContext,
				consumerConfigToUpdate *nats.ConsumerConfig,
			) {
				jsCtx.On("UpdateConsumer", jsBackend.GetConfig().JSStreamName, consumerConfigToUpdate).Return(&nats.ConsumerInfo{
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
			jsCtxMock := &jetstreamv2mocks.JetStreamContext{}
			js := &JetStream{
				jsCtx: jsCtxMock,
			}
			sub := subtesting.NewSubscription("test", "test",
				subtesting.WithMaxInFlight(tc.givenSubMaxInFlight),
			)

			// setup the jetstreamv2mocks
			consumer := nats.ConsumerInfo{
				Name:   "name",
				Config: nats.ConsumerConfig{MaxAckPending: tc.givenConsumerMaxAckPending},
			}
			tc.givenjetstreamv2mocks(js, jsCtxMock, tc.wantConfigToUpdate)

			// when
			err := js.syncConsumerMaxInFlight(sub, consumer)

			// then
			assert.NoError(t, err)
			jsCtxMock.AssertExpectations(t)
		})
	}
}

// Test_SyncConsumersAndSubscriptions_ForErrors test the syncConsumerAndSubscription for right error handling.
func Test_SyncConsumersAndSubscriptions_ForErrors(t *testing.T) {
	// pre-requisites
	subWithOneType := NewSubscriptionWithOneType()
	js := &JetStream{cleaner: &cleaner.JetStreamCleaner{}}
	eventType := subWithOneType.Status.Types[0]
	jsSubject := js.GetJetStreamSubject(subWithOneType.Spec.Source,
		eventType.CleanType,
		subWithOneType.Spec.TypeMatching,
	)
	jsSubKey := NewSubscriptionSubjectIdentifier(subWithOneType, jsSubject)
	invalidSubscriber := &subscriberStub{isValid: false}

	testCases := []struct {
		name             string
		jetStreamContext *jetStreamContextStub
		jsBackend        *JetStream
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
			name: "Subscribe call should result into error when a NATS subscription is invalid",
			jsBackend: &JetStream{subscriptions: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: invalidSubscriber,
			}},
			jetStreamContext: &jetStreamContextStub{
				consumerInfo:      &nats.ConsumerInfo{},
				consumerInfoError: nil,

				subscribeError: ErrFailedSubscribe,
			},
			wantError: ErrFailedSubscribe,
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

				updateConsumer:      nil,
				updateConsumerError: nats.ErrJetStreamNotEnabled,
			},
			wantError: ErrUpdateConsumer,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			callback := func(m *nats.Msg) {}
			js := JetStream{
				subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
				metricsCollector: metrics.NewCollector(),
				jsCtx:            testCase.jetStreamContext,
				cleaner:          &cleaner.JetStreamCleaner{},
			}

			// when
			err := js.syncConsumerAndSubscription(subWithOneType, callback)

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
	err := js.syncConsumerAndSubscription(subWithType, callback)

	// then
	require.ErrorIs(t, err, ErrMissingSubscription)
}

// Test_DeleteSubscriptionFromJetStream test the behaviour of the deleteSubscriptionFromJetStream function.
func Test_DeleteSubscriptionFromJetStream(t *testing.T) {
	// pre-requisites
	jsBackend := &JetStream{
		cleaner: &cleaner.JetStreamCleaner{},
	}
	subWithOneType := NewSubscriptionWithOneType()
	jsSubject := jsBackend.GetJetStreamSubject(subWithOneType.Spec.Source,
		subWithOneType.Status.Types[0].CleanType,
		subWithOneType.Spec.TypeMatching,
	)
	jsSubKey := NewSubscriptionSubjectIdentifier(subWithOneType, jsSubject)

	testCases := []struct {
		name                  string
		jsBackend             *JetStream
		jsCtx                 *jetStreamContextStub
		subscriber            Subscriber
		givenSubscriptionsMap map[SubscriptionSubjectIdentifier]Subscriber
		wantError             error
		wantSubscriptionMap   map[SubscriptionSubjectIdentifier]Subscriber
	}{
		{
			name: "Valid subscriber returns error on Unsubscribe",
			givenSubscriptionsMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: &subscriberStub{
					isValid:          true,
					unsubscribeError: nats.ErrConnectionNotTLS,
				},
			},
			wantError: ErrFailedUnsubscribe,
		},
		{
			name: "Valid unsubscribe with err on consumer delete",
			givenSubscriptionsMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: &subscriberStub{
					isValid:          true,
					unsubscribeError: nil,
				},
			},
			jsCtx: &jetStreamContextStub{
				deleteConsumerErr: nats.ErrJetStreamNotEnabled,
			},
			wantError: ErrDeleteConsumer,
		},
		{
			name: "ErrConsumerNotFound error should be ignored",
			givenSubscriptionsMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: &subscriberStub{
					isValid:          true,
					unsubscribeError: nil,
				},
			},
			jsCtx: &jetStreamContextStub{
				deleteConsumerErr: nats.ErrConsumerNotFound,
			},
			wantSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{},
			wantError:           nil,
		},
		{
			name: "Invalid subscriber doesn't try to unsubscribe",
			givenSubscriptionsMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: &subscriberStub{
					isValid:          false,
					unsubscribeError: nats.ErrConnectionNotTLS,
				},
			},
			jsCtx:               &jetStreamContextStub{},
			wantSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{},
			wantError:           nil,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			// inject the given subscriber
			jsBackend.jsCtx = testCase.jsCtx
			jsBackend.subscriptions = testCase.givenSubscriptionsMap

			// when
			resultErr := jsBackend.deleteSubscriptionFromJetStream(jsBackend.subscriptions[jsSubKey], jsSubKey)

			// then
			assert.ErrorIs(t, resultErr, testCase.wantError)
			if testCase.wantError == nil {
				assert.Equal(t, testCase.wantSubscriptionMap, jsBackend.subscriptions)
			}
		})
	}
}

// Test_DeleteInvalidConsumers tests the behaviour of the DeleteInvalidConsumers function.
func Test_DeleteInvalidConsumers(t *testing.T) {
	// pre-requisites
	jsBackend := &JetStream{
		cleaner: &cleaner.JetStreamCleaner{},
	}

	subs := NewSubscriptionsWithMultipleTypes()
	givenConsumers := NewConsumers(subs, jsBackend)

	danglingConsumer := &nats.ConsumerInfo{
		Name:      "dangling-invalid-consumer",
		Config:    nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights},
		PushBound: false,
	}
	// add a dangling consumer which should be deleted
	givenConsumersWithDangling := givenConsumers
	givenConsumersWithDangling = append(givenConsumersWithDangling, danglingConsumer)

	testCases := []struct {
		name               string
		givenSubscriptions []v1alpha2.Subscription
		jetStreamContext   *jetStreamContextStub
		wantConsumers      []*nats.ConsumerInfo
		wantError          error
	}{
		{
			name:               "no consumer should be deleted",
			givenSubscriptions: subs,
			jetStreamContext: &jetStreamContextStub{
				consumers: givenConsumers,
			},
			wantConsumers: givenConsumers,
			wantError:     nil,
		},
		{
			name:               "a dangling invalid consumer should be deleted",
			givenSubscriptions: subs,
			jetStreamContext: &jetStreamContextStub{
				consumers: givenConsumersWithDangling,
			},
			wantConsumers: givenConsumers,
			wantError:     nil,
		},
		{
			name:               "no consumer should be deleted",
			givenSubscriptions: subs,
			jetStreamContext: &jetStreamContextStub{
				consumers:         givenConsumersWithDangling,
				deleteConsumerErr: nats.ErrConnectionNotTLS,
			},
			wantError: ErrDeleteConsumer,
		},
		{
			name:               "all consumers must be deleted if there is no subscription resource",
			givenSubscriptions: []v1alpha2.Subscription{},
			jetStreamContext: &jetStreamContextStub{
				consumers: givenConsumers,
			},
			wantConsumers: []*nats.ConsumerInfo{},
			wantError:     nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			jsBackend.jsCtx = tc.jetStreamContext

			// when
			err := jsBackend.DeleteInvalidConsumers(tc.givenSubscriptions)

			// then
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				cons := jsBackend.jsCtx.Consumers("")
				actualConsumers := []*nats.ConsumerInfo{}
				for con := range cons {
					actualConsumers = append(actualConsumers, con)
				}
				assert.Equal(t, len(tc.wantConsumers), len(actualConsumers))
				assert.Equal(t, tc.wantConsumers, actualConsumers)
			}
		})
	}
}

// Test_DeleteInvalidConsumersFromJetStream test the behaviour of the deleteSubscriptionFromJetStream function.
func Test_IsConsumerUsedByKymaSub(t *testing.T) {
	// pre-requisites
	jsBackend := &JetStream{
		cleaner: &cleaner.JetStreamCleaner{},
	}

	subs := NewSubscriptionsWithMultipleTypes()

	givenConName1 := computeConsumerName(&subs[0], jsBackend.getJetStreamSubject(subs[0].Spec.Source,
		subs[0].Status.Types[0].CleanType,
		subs[0].Spec.TypeMatching,
	))
	givenConName2 := computeConsumerName(&subs[1], jsBackend.getJetStreamSubject(subs[1].Spec.Source,
		subs[1].Status.Types[1].CleanType,
		subs[1].Spec.TypeMatching,
	))
	givenNotExistentConName := "non-existing-consumer-name"

	// creating a consumer with existing v1 subject, but different namespace
	sub3 := subs[0]
	sub3.Namespace = "test3"
	invalidDanglingConsumer := computeConsumerName(&sub3, jsBackend.getJetStreamSubject(sub3.Spec.Source,
		sub3.Status.Types[0].CleanType,
		sub3.Spec.TypeMatching,
	))

	testCases := []struct {
		name               string
		givenConsumerName  string
		givenSubscriptions []v1alpha2.Subscription
		wantResult         bool
		wantError          error
	}{
		{
			name:               "Consumer name should be found in subscriptions types",
			givenConsumerName:  givenConName1,
			givenSubscriptions: subs,
			wantResult:         true,
		},
		{
			name:               "Consumer name should be found in the second subscriptions types",
			givenConsumerName:  givenConName2,
			givenSubscriptions: subs,
			wantResult:         true,
		},
		{
			name:               "Consumer should be not found in subscriptions types",
			givenConsumerName:  givenNotExistentConName,
			givenSubscriptions: subs,
			wantResult:         false,
		},
		{
			name:               "Consumer with the same subject but different namespace should not be found",
			givenConsumerName:  invalidDanglingConsumer,
			givenSubscriptions: subs,
			wantResult:         false,
		},
		{
			name:               "should return false if there is no subscription resource",
			givenConsumerName:  givenConName1,
			givenSubscriptions: []v1alpha2.Subscription{},
			wantResult:         false,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// when
			result := jsBackend.isConsumerUsedByKymaSub(tc.givenConsumerName, tc.givenSubscriptions)

			// then
			if testCase.wantError == nil {
				assert.Equal(t, testCase.wantResult, result)
			}
		})
	}
}

////////////////////////////////////////////////////////////////////////
// test helpers
///////////////////////////////////////////////////////////////////////

func NewSubscriptionWithEmptyTypes() *v1alpha2.Subscription {
	return subtesting.NewSubscription("test", "test",
		subtesting.WithStatusTypes(nil),
	)
}

func NewSubscriptionWithOneType() *v1alpha2.Subscription {
	return subtesting.NewSubscription("test", "test",
		subtesting.WithSourceAndType(subtesting.EventSource, subtesting.CloudEventType),
		subtesting.WithTypeMatchingStandard(),
		subtesting.WithMaxInFlight(DefaultMaxInFlights),
		subtesting.WithStatusTypes([]v1alpha2.EventType{
			{
				OriginalType: subtesting.CloudEventType,
				CleanType:    subtesting.CloudEventType,
			},
		}),
	)
}

func NewSubscriptionsWithMultipleTypes() []v1alpha2.Subscription {
	sub1 := subtesting.NewSubscription("test1", "test1",
		subtesting.WithSourceAndType(subtesting.EventSourceClean, subtesting.OrderCreatedV1Event),
		subtesting.WithTypeMatchingStandard(),
		subtesting.WithMaxInFlight(DefaultMaxInFlights),
		subtesting.WithStatusTypes([]v1alpha2.EventType{
			{
				OriginalType: subtesting.OrderCreatedV1Event,
				CleanType:    subtesting.OrderCreatedV1Event,
			},
		}),
	)
	sub2 := subtesting.NewSubscription("test2", "test2",
		subtesting.WithSourceAndType(subtesting.EventSourceClean, subtesting.OrderCreatedV1Event),
		subtesting.WithSourceAndType(subtesting.EventSourceClean, subtesting.OrderCreatedV1EventNotClean),
		subtesting.WithTypeMatchingStandard(),
		subtesting.WithMaxInFlight(DefaultMaxInFlights),
		subtesting.WithStatusTypes([]v1alpha2.EventType{
			{
				OriginalType: subtesting.OrderCreatedV1Event,
				CleanType:    subtesting.OrderCreatedV1Event,
			},
			{
				OriginalType: subtesting.OrderCreatedV1Event,
				CleanType:    subtesting.OrderCreatedV1Event,
			},
		}),
	)
	return []v1alpha2.Subscription{*sub1, *sub2}
}

func NewConsumers(subs []v1alpha2.Subscription, jsBackend *JetStream) []*nats.ConsumerInfo {
	sub1 := subs[0]
	sub2 := subs[1]

	// existing subscription type with existing consumers
	// namespace: test1, sub: test1, type: v1
	return []*nats.ConsumerInfo{
		{
			Name: computeConsumerName(
				&sub1,
				jsBackend.getJetStreamSubject(sub1.Spec.Source,
					sub1.Status.Types[0].CleanType,
					sub1.Spec.TypeMatching,
				)),
			Config:    nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights},
			PushBound: false,
		},
		{
			Name: computeConsumerName(
				&sub2,
				jsBackend.getJetStreamSubject(sub2.Spec.Source,
					sub2.Status.Types[0].CleanType,
					sub2.Spec.TypeMatching,
				)),
			Config:    nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights},
			PushBound: false,
		},
		{
			Name: computeConsumerName(
				&sub2,
				jsBackend.getJetStreamSubject(sub2.Spec.Source,
					sub2.Status.Types[1].CleanType,
					sub2.Spec.TypeMatching,
				)),
			Config:    nats.ConsumerConfig{MaxAckPending: DefaultMaxInFlights},
			PushBound: false,
		},
	}
}

////////////////////////////////////////////////////////////////////////
// test fixtures
///////////////////////////////////////////////////////////////////////

func fixtureNatsConfig() backendnats.Config {
	return backendnats.Config{
		JSStreamName:            "not-empty",
		JSStreamStorageType:     StorageTypeMemory,
		JSStreamRetentionPolicy: RetentionPolicyLimits,
		JSStreamDiscardPolicy:   DiscardPolicyNew,
	}
}

// fixtureNatsConfigAsStreamInfo is the converted backendnats.Config to nats.StreamInfo.
// Use this fixture to make streamIsConfiguredCorrectly happy.
func fixtureNatsConfigAsStreamInfo() *nats.StreamInfo {
	return &nats.StreamInfo{}
}

////////////////////////////////////////////////////////////////////////
// stubs
///////////////////////////////////////////////////////////////////////

type ceClientStub struct{}

func (c ceClientStub) Send(_ context.Context, _ cev2event.Event) cev2protocol.Result {
	panic("implement me")
}

func (c ceClientStub) Request(_ context.Context, _ cev2event.Event) (*cev2event.Event, cev2protocol.Result) {
	panic("implement me")
}

func (c ceClientStub) StartReceiver(_ context.Context, _ interface{}) error {
	panic("implement me")
}

type ceClientFactoryStub struct {
	client cev2.Client
	err    error
}

func (c ceClientFactoryStub) NewHTTP(_ ...http.Option) (cev2.Client, error) {
	return c.client, c.err
}

type loggerStub struct {
	t *testing.T
}

func (l loggerStub) WithContext() *zap.SugaredLogger {
	return zaptest.NewLogger(l.t).Sugar()
}

func (l loggerStub) WithTracing(_ context.Context) *zap.SugaredLogger {
	return zaptest.NewLogger(l.t).Sugar()
}

type connectionBuilderStub struct {
	conn Connection
	err  error
}

func (c *connectionBuilderStub) Build() (Connection, error) {
	return c.conn, c.err
}

type connectionStub struct {
	isConnected           bool
	jetStreamContext      nats.JetStreamContext
	jetStreamContextError error
}

func (c connectionStub) JetStream(_ ...nats.JSOpt) (nats.JetStreamContext, error) {
	return c.jetStreamContext, c.jetStreamContextError
}

func (c connectionStub) IsConnected() bool {
	return c.isConnected
}

func (c connectionStub) SetClosedHandler(_ nats.ConnHandler) {}

func (c connectionStub) SetReconnectHandler(_ nats.ConnHandler) {}
