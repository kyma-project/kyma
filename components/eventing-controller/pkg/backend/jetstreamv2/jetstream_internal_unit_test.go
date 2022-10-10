package jetstreamv2

import (
	"fmt"
	"sync"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/mocks"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test_SyncNATSConsumers tests creation/binding of the consumer and NATS Subscriptions.
func TestVlad_SyncNATSConsumers(t *testing.T) {
	testEnv := getTestEnvironmentWithMock(t)
	callback := func(m *nats.Msg) {}
	subWithType := testingv2.NewSubscription("test", "test",
		testingv2.WithSourceAndType(testingv2.EventSource, testingv2.CloudEventType),
		testingv2.WithTypeMatchingStandard(),
		testingv2.WithMaxInFlight(env.DefaultMaxInFlight),
		testingv2.WithStatusTypes([]v1alpha2.EventType{
			{
				OriginalType: testingv2.CloudEventType,
				CleanType:    testingv2.CloudEventType,
			},
		}),
	)
	_, jsSubKey, defaultConsumerInfo, _ := generateConsumerInfra(testEnv.jsBackend, subWithType, subWithType.Status.Types[0])
	subscriberMock := &mocks.Subscriber{}

	testCases := []struct {
		name                 string
		givenSubscription    *v1alpha2.Subscription
		givenSubscriptionMap map[SubscriptionSubjectIdentifier]Subscriber
		givenMocks           func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier)
		givenAssertions      func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier)
		wantErr              *string
	}{
		{
			name: "No types on subscription should result in no action",
			givenSubscription: testingv2.NewSubscription("test", "test",
				testingv2.WithStatusTypes(nil),
			),
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
			},
			givenAssertions: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				jsCtx.AssertNotCalled(t, "ConsumerInfo", mock.Anything, mock.Anything)
			},
		},
		{
			name:              "Failed to get Consumer info should give expected error",
			givenSubscription: subWithType,
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
			wantErr: errorMessage(nats.ErrStreamNameRequired.Error()),
		},
		{
			name:              "Creating a consumer/subscribing should be successful",
			givenSubscription: subWithType,
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
			givenSubscription: subWithType,
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
			wantErr: errorMessage(nats.ErrJetStreamNotEnabled.Error()),
		},
		{
			name:              "Creating a new NATS Subscription should result in a Stream not found error",
			givenSubscription: subWithType,
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
			wantErr: errorMessage(fmt.Sprintf(FailedToSubscribeMsg, nats.ErrStreamNotFound)),
		},
		{
			name:              "Bind invalid NATS Subscription",
			givenSubscription: subWithType,
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
			givenSubscription: subWithType,
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
			wantErr: errorMessage(fmt.Sprintf(FailedToSubscribeMsg, nats.ErrJetStreamNotEnabled)),
		},
		{
			name:              "Update maxInFlight on the consumer side, when the subscription config changes",
			givenSubscription: subWithType,
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: subscriberMock,
			},
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscriberMock.On("IsValid").Return(true)

					subWithType.Spec.Config[v1alpha2.MaxInFlightMessages] = "20"
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
			givenSubscription: subWithType,
			givenSubscriptionMap: map[SubscriptionSubjectIdentifier]Subscriber{
				jsSubKey: subscriberMock,
			},
			givenMocks: func(sub *v1alpha2.Subscription, jsBackend *JetStream, jsCtx *mocks.JetStreamContext, jsSubKey SubscriptionSubjectIdentifier) {
				for _, eventType := range sub.Status.Types {
					jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, eventType.CleanType, sub.Spec.TypeMatching)
					jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
					subscriberMock.On("IsValid").Return(true)

					defaultConsumerInfo.Config.MaxAckPending = defaultMaxInFlights
					subWithType.Spec.Config[v1alpha2.MaxInFlightMessages] = "20"
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
			wantErr: errorMessage(nats.ErrJetStreamNotEnabled.Error()),
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
				require.Equal(t, err.Error(), *testCase.wantErr)
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

// Test_CheckNATSSubscriptionsCount test the behaviour of the checkNATSSubscriptionsCount function.
func TestVlad_CheckNATSSubscriptionsCount(t *testing.T) {
	// given
	subWithType := testingv2.NewSubscription("test", "test",
		testingv2.WithSourceAndType(testingv2.EventSource, testingv2.CloudEventType),
		testingv2.WithTypeMatchingStandard(),
		testingv2.WithMaxInFlight(env.DefaultMaxInFlight),
		testingv2.WithStatusTypes([]v1alpha2.EventType{
			{
				OriginalType: testingv2.CloudEventType,
				CleanType:    testingv2.CloudEventType,
			},
		}),
	)

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
		givenSubscriptionMap func() map[SubscriptionSubjectIdentifier]Subscriber
		wantErrMsg           *string
	}{
		{
			name: "empty subscriptions map with subscription with no types should result in not error",
			givenSubscription: testingv2.NewSubscription("test", "test",
				testingv2.WithStatusTypes(nil),
			),
			givenSubscriptionMap: func() map[SubscriptionSubjectIdentifier]Subscriber {
				return map[SubscriptionSubjectIdentifier]Subscriber{}
			},
			wantErrMsg: nil,
		},
		//{
		//	name:              "if the subscriptions map contains all the NATS Subscriptions, no error is expected",
		//	givenSubscription: subWithType,
		//	givenSubscriptionMap: func() map[SubscriptionSubjectIdentifier]Subscriber {
		//		_, jsSubKey, _, natsSub := generateConsumerInfra(testEnv.jsBackend, subWithType, subWithType.Status.Types[0])
		//		return map[SubscriptionSubjectIdentifier]Subscriber{
		//			jsSubKey: natsSub,
		//		}
		//	},
		//	wantErrMsg: nil,
		//},
		{
			name:              "unexpected empty subscriptions map should result into an error",
			givenSubscription: subWithType,
			givenSubscriptionMap: func() map[SubscriptionSubjectIdentifier]Subscriber {
				return map[SubscriptionSubjectIdentifier]Subscriber{}
			},
			wantErrMsg: errorMessage(fmt.Sprintf(MissingNATSSubscriptionMsgWithInfo, subWithType.Spec.Types[0])),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {

			// inject the fake subscription map
			jsBackend.subscriptions = testCase.givenSubscriptionMap()

			// when
			err := jsBackend.checkNATSSubscriptionsCount(testCase.givenSubscription)

			// then
			if testCase.wantErrMsg == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, err.Error(), *testCase.wantErrMsg)
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
