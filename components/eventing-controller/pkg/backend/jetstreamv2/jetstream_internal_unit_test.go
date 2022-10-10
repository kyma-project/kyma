package jetstreamv2

import (
	backenderrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2/errors"
	"sync"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/mocks"
	subtesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test_SyncNATSConsumers tests creation/binding of the consumer and NATS Subscriptions.
func Test_SyncNATSConsumers(t *testing.T) {
	testEnv := getTestEnvironmentWithMock(t)
	callback := func(m *nats.Msg) {}
	subWithOneType := subtesting.NewSubscriptionWithOneType()
	_, jsSubKey, defaultConsumerInfo, _ := generateConsumerInfra(testEnv.jsBackend, subWithOneType, subWithOneType.Status.Types[0])
	subscriberMock := &mocks.Subscriber{}

	testCases := []struct {
		name                 string
		givenSubscription    *v1alpha2.Subscription
		givenSubscriptionMap map[SubscriptionSubjectIdentifier]Subscriber
		givenMocks           func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier)
		givenAssertions      func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier)
		wantErr              error
	}{
		{
			name:              "No types on subscription should result in no action",
			givenSubscription: subtesting.NewSubscriptionWithEmptyTypes(),
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				jsCtx.AssertNotCalled(t, "ConsumerInfo", mock.Anything, mock.Anything)
			},
		},
		{
			name:              "Failed to get Consumer info should give expected error",
			givenSubscription: subWithOneType,
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(nil, nats.ErrStreamNameRequired)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
				}
			},
			wantErr: nats.ErrStreamNameRequired,
		},
		{
			name:              "Creating a consumer/subscribing should be successful",
			givenSubscription: subWithOneType,
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					_, jsSubKey, newConsumer, natsSub := generateConsumerInfra(jsBackend, sub, eventType)
					subscribeArgs := getSubOptsAsVariadicSlice(jsBackend, sub, eventType)

					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(nil, nats.ErrConsumerNotFound)
					jsCtx.On("AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(newConsumer, nil)
					jsCtx.On("Subscribe", subscribeArgs...).Return(natsSub, nil)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscribeArgs := getSubOptsAsVariadicSlice(jsBackend, sub, eventType)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
					jsCtx.AssertCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
					jsCtx.AssertCalled(t, "Subscribe", subscribeArgs...)
				}
			},
		},
		{
			name:              "Adding a consumer should result in an error",
			givenSubscription: subWithOneType,
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(nil, nats.ErrConsumerNotFound)
					jsCtx.On("AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(nil, nats.ErrJetStreamNotEnabled)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscribeArgs := getSubOptsAsVariadicSlice(jsBackend, sub, eventType)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
					jsCtx.AssertCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
					jsCtx.AssertNotCalled(t, "Subscribe", subscribeArgs...)
				}
			},
			wantErr: nats.ErrJetStreamNotEnabled,
		},
		{
			name:              "Creating a new NATS Subscription should result in a Stream not found error",
			givenSubscription: subWithOneType,
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					_, jsSubKey, newConsumer, _ := generateConsumerInfra(jsBackend, sub, eventType)
					subscribeArgs := getSubOptsAsVariadicSlice(jsBackend, sub, eventType)

					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(nil, nats.ErrConsumerNotFound)
					jsCtx.On("AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(newConsumer, nil)
					jsCtx.On("Subscribe", subscribeArgs...).Return(nil, nats.ErrStreamNotFound)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscribeArgs := getSubOptsAsVariadicSlice(jsBackend, sub, eventType)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
					jsCtx.AssertCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
					jsCtx.AssertCalled(t, "Subscribe", subscribeArgs...)
				}
			},
			wantErr: &backenderrors.FailedToSubscribeOnNATSError{OriginalError: nats.ErrStreamNotFound},
		},
		{
			name:              "Bind invalid NATS Subscription",
			givenSubscription: subWithOneType,
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject, jsSubKey, newConsumer, natsSub := generateConsumerInfra(jsBackend, sub, eventType)

					// inject invalid subscriber
					sub := &mocks.Subscriber{}
					sub.On("IsValid").Return(false)
					jsBackend.subscriptions[jsSubKey] = sub

					// mock the expected calls
					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(newConsumer, nil)
					jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).Return(natsSub, nil)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
					jsCtx.AssertNotCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
					jsCtx.AssertCalled(t, "Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn"))

				}
			},
		},
		{
			name:              "Bind invalid NATS Subscription should result into error",
			givenSubscription: subWithOneType,
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject, jsSubKey, newConsumer, natsSub := generateConsumerInfra(jsBackend, sub, eventType)

					// inject invalid subscriber
					sub := &mocks.Subscriber{}
					sub.On("IsValid").Return(false)
					jsBackend.subscriptions[jsSubKey] = sub

					// mock the expected calls
					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(newConsumer, nil)
					jsCtx.On("Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn")).Return(natsSub, nats.ErrJetStreamNotEnabled)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
					jsCtx.AssertNotCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
					jsCtx.AssertCalled(t, "Subscribe", jsSubject, mock.AnythingOfType("nats.MsgHandler"), mock.AnythingOfType("nats.subOptFn"))

				}
			},
			wantErr: &backenderrors.FailedToSubscribeOnNATSError{OriginalError: nats.ErrJetStreamNotEnabled},
		},
		{
			name:              "Update maxInFlight on the consumer side, when the subscription config changes",
			givenSubscription: subWithOneType,
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: subscriberMock,
			},
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscriberMock.On("IsValid").Return(true)

					subWithOneType.Spec.Config[v1alpha2.MaxInFlightMessages] = "20"
					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(defaultConsumerInfo, nil)
					jsCtx.On("UpdateConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(defaultConsumerInfo, nil)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())

					defaultConsumerInfo.Config.MaxAckPending = 20
					jsCtx.AssertCalled(t, "UpdateConsumer", jsBackend.Config.JSStreamName, &defaultConsumerInfo.Config)

					jsCtx.AssertNotCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
				}
			},
		},
		{
			name:              "Update consumer should result into an error",
			givenSubscription: subWithOneType,
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: subscriberMock,
			},
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscriberMock.On("IsValid").Return(true)

					defaultConsumerInfo.Config.MaxAckPending = defaultMaxInFlights
					subWithOneType.Spec.Config[v1alpha2.MaxInFlightMessages] = "20"
					jsCtx.On("ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName()).Return(defaultConsumerInfo, nil)
					jsCtx.On("UpdateConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig")).Return(nil, nats.ErrJetStreamNotEnabled)
				}
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				// check if the expected methods got called
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

					jsCtx.AssertCalled(t, "ConsumerInfo", jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())

					defaultConsumerInfo.Config.MaxAckPending = 20
					jsCtx.AssertCalled(t, "UpdateConsumer", jsBackend.Config.JSStreamName, &defaultConsumerInfo.Config)
					jsCtx.AssertNotCalled(t, "AddConsumer", jsBackend.Config.JSStreamName, mock.AnythingOfType("*nats.ConsumerConfig"))
				}
			},
			wantErr: nats.ErrJetStreamNotEnabled,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			testEnv := getTestEnvironmentWithMock(t)
			if testCase.givenSubscriptionMap != nil {
				testEnv.jsBackend.subscriptions = testCase.givenSubscriptionMap
			}
			for _, eventType := range testCase.givenSubscription.Status.Types {
				jsSubject := testEnv.jsBackend.getJetStreamSubject(testCase.givenSubscription.Spec.Source, eventType.CleanType, testCase.givenSubscription.Spec.TypeMatching)
				jsSubKey := NewSubscriptionSubjectIdentifier(testCase.givenSubscription, jsSubject)

				// mock the expected calls
				tc.givenMocks(testCase.givenSubscription, testEnv.jsBackend, testEnv.jsCtxMock, jsSubKey)
			}

			// when
			err := testEnv.jsBackend.syncNATSConsumers(testCase.givenSubscription, callback, testEnv.logger.WithContext())

			// then
			// check if error is expected
			if testCase.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, testCase.wantErr)
			}

			// then
			// check the expected assetions
			for _, eventType := range tc.givenSubscription.Status.Types {
				jsSubject := testEnv.jsBackend.getJetStreamSubject(testCase.givenSubscription.Spec.Source, eventType.CleanType, testCase.givenSubscription.Spec.TypeMatching)
				jsSubKey := NewSubscriptionSubjectIdentifier(testCase.givenSubscription, jsSubject)

				tc.givenAssertions(testCase.givenSubscription, testEnv.jsBackend, testEnv.jsCtxMock, jsSubKey)
			}
		})
	}
}

// Test_CheckNATSSubscriptionsCount tests the behaviour of the checkNATSSubscriptionsCount function.
func Test_CheckNATSSubscriptionsCount(t *testing.T) {
	// given
	subWithType := subtesting.NewSubscriptionWithOneType()
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	jsBackend := &JetStream{
		subscriptions: make(map[SubscriptionSubjectIdentifier]Subscriber),
		cleaner:       jsCleaner,
	}

	testCases := []struct {
		name                 string
		givenSubscription    *v1alpha2.Subscription
		givenSubscriptionMap map[SubscriptionSubjectIdentifier]Subscriber
		wantErr              error
	}{
		{
			name:                 "empty subscriptions map with subscription with no types should result in not error",
			givenSubscription:    subtesting.NewSubscriptionWithEmptyTypes(),
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
			wantErr:              &backenderrors.ErrMissingNATSSubscription{},
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

// sets up the TestEnvironment with the mocks instead of running a full-fledged NATS test server.
func getTestEnvironmentWithMock(t *testing.T) *TestEnvironment {
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)
	jsCtx := &mocks.JetStreamContext{}
	natsConfig := defaultNatsConfig("localhost")
	metricsCollector := metrics.NewCollector()

	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	jsBackend := &JetStream{
		Config:           natsConfig,
		Conn:             nil,
		jsCtx:            jsCtx,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
		sinks:            sync.Map{},
		logger:           defaultLogger,
		metricsCollector: metricsCollector,
		cleaner:          jsCleaner,
	}

	return &TestEnvironment{
		jsBackend:  jsBackend,
		logger:     defaultLogger,
		natsConfig: natsConfig,
		cleaner:    jsCleaner,
		jsCtxMock:  jsCtx,
	}
}

// getSubOptsAsVariadicSlice computes the variadic argument for the JetStreamContext.Subscribe call.
func getSubOptsAsVariadicSlice(jsBackend *JetStream, sub *v1alpha2.Subscription, eventType v1alpha2.EventType) []interface{} {
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
	jsSubOptions := jsBackend.getDefaultSubscriptionOptions(jsSubKey, sub.GetMaxInFlightMessages(jsBackend.namedLogger()))
	var args []interface{}
	args = append(args, jsSubject)
	args = append(args, mock.AnythingOfType("nats.MsgHandler"))
	for i := 0; i < len(jsSubOptions); i++ {
		args = append(args, mock.Anything)
	}
	return args
}

// generateConsumerInfra is a helper function to initialize the resources which are necessary to test the consumer logic.
func generateConsumerInfra(jsBackend *JetStream, sub *v1alpha2.Subscription, eventType v1alpha2.EventType) (string, SubscriptionSubjectIdentifier, *nats.ConsumerInfo, *nats.Subscription) {
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
	consumerConfig := jsBackend.getConsumerConfig(sub, jsSubKey, jsSubject)
	newConsumer := &nats.ConsumerInfo{
		Stream: jsBackend.Config.JSStreamName,
		Name:   eventType.OriginalType,
		Config: *consumerConfig,
	}
	natsSub := &nats.Subscription{
		Subject: eventType.CleanType,
		Queue:   "test",
	}
	return jsSubject, jsSubKey, newConsumer, natsSub
}

func errorMessage(val string) *string {
	errString := val
	return &errString
}
