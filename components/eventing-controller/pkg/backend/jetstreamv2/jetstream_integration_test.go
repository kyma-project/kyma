package jetstreamv2

import (
	"fmt"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

// TestJetStreamSubAfterSync_SinkChange tests the SyncSubscription method
// when only the sink is changed in subscription, then it should not re-create
// NATS subjects on nats-server.
func TestJetStreamSubAfterSync_SinkChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscribers
	subscriber1 := evtesting.NewSubscriber()
	defer subscriber1.Shutdown()
	require.True(t, subscriber1.IsRunning())
	subscriber2 := evtesting.NewSubscriber()
	defer subscriber2.Shutdown()
	require.True(t, subscriber2.IsRunning())

	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber1.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := testEnvironment.cleaner.CleanEventType(sub.Spec.Types[0])
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	require.NoError(t,
		SendCloudEventToJetStream(jsBackend,
			jsBackend.GetJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching),
			evtestingv2.CloudEventData,
			types.ContentModeBinary),
	)
	require.NoError(t, subscriber1.CheckEvent(evtestingv2.CloudEventData))

	// set metadata on NATS subscriptions
	msgLimit, bytesLimit := 2048, 2048
	require.Len(t, jsBackend.subscriptions, 1)
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2.SinkURL

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject := jsBackend.GetJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())

	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.Equal(t, subMsgLimit, msgLimit)
	require.Equal(t, subBytesLimit, bytesLimit)

	// Test if the subscription is working for new sink only
	require.NoError(t,
		SendCloudEventToJetStream(jsBackend,
			jsBackend.GetJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching),
			evtestingv2.CloudEventData,
			types.ContentModeBinary),
	)

	// Old sink should not have received the event, the new sink should have
	require.Error(t, subscriber1.CheckEvent(evtestingv2.CloudEventData))
	require.NoError(t, subscriber2.CheckEvent(evtestingv2.CloudEventData))
}

// TestMultipleJSSubscriptionsToSameEvent tests the behaviour of JS
// when multiple subscriptions need to receive the same event.
func TestMultipleJSSubscriptionsToSameEvent(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	// Create 3 subscriptions having the same sink and the same event type
	var subs [3]*eventingv1alpha2.Subscription
	for i := 0; i < len(subs); i++ {
		subs[i] = evtestingv2.NewSubscription(fmt.Sprintf("sub-%d", i), "foo",
			evtestingv2.WithSourceAndType(evtestingv2.EventSource, evtestingv2.OrderCreatedEventType),
			evtestingv2.WithSinkURL(subscriber.SinkURL),
			evtestingv2.WithTypeMatchingStandard(),
			evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
		)
		AddJSCleanEventTypesToStatus(subs[i], testEnvironment.cleaner)
		// when
		err := jsBackend.SyncSubscription(subs[i])
		// then
		require.NoError(t, err)
	}

	// Send only one event. It should be multiplexed to 3 by NATS, cause 3 subscriptions exist
	require.NoError(t,
		SendCloudEventToJetStream(jsBackend,
			jsBackend.GetJetStreamSubject(evtestingv2.EventSource,
				evtestingv2.OrderCreatedEventType,
				eventingv1alpha2.TypeMatchingStandard),
			evtestingv2.CloudEventData,
			types.ContentModeBinary),
	)
	// Check for the 3 events that should be received by the subscriber
	for i := 0; i < len(subs); i++ {
		require.NoError(t, subscriber.CheckEvent(evtestingv2.CloudEventData))
	}
	// Delete all 3 subscription
	for i := 0; i < len(subs); i++ {
		require.NoError(t, jsBackend.DeleteSubscription(subs[i]))
	}
	// Check if all subscriptions are deleted in NATS
	// Send an event again which should not be delivered to subscriber
	require.NoError(t,
		SendCloudEventToJetStream(jsBackend,
			jsBackend.GetJetStreamSubject(evtestingv2.EventSource,
				evtestingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingStandard),
			evtestingv2.CloudEventData2,
			types.ContentModeBinary),
	)
	// Check for the event that did not reach the subscriber
	// Store should never return evtestingv2.CloudEventData2
	// hence CheckEvent should fail to match evtestingv2.CloudEventData2
	require.Error(t, subscriber.CheckEvent(evtestingv2.CloudEventData2))
}

// TestJSSubscriptionRedeliverWithFailedDispatch tests the redelivering
// of event when the dispatch fails.
func TestJSSubscriptionRedeliverWithFailedDispatch(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	subscriber.Shutdown() // shutdown the subscriber intentionally
	require.False(t, subscriber.IsRunning())

	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithSourceAndType(evtestingv2.EventSource, evtestingv2.OrderCreatedCleanEvent),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingExact(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// when
	// send an event

	require.NoError(t,
		SendCloudEventToJetStream(jsBackend,
			jsBackend.GetJetStreamSubject(evtestingv2.EventSource,
				evtestingv2.OrderCreatedCleanEvent,
				eventingv1alpha2.TypeMatchingExact),
			evtestingv2.CloudEventData,
			types.ContentModeBinary),
	)

	// then
	// it should have failed to dispatch
	require.Error(t, subscriber.CheckEvent(evtestingv2.CloudEventData))

	// when
	// start a new subscriber
	subscriber = evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())
	// and update sink in the subscription
	sub.Spec.Sink = subscriber.SinkURL
	require.NoError(t, jsBackend.SyncSubscription(sub))

	// then
	// the same event should be redelivered
	require.Eventually(t, func() bool {
		return subscriber.CheckEvent(evtestingv2.CloudEventData) == nil
	}, 60*time.Second, 5*time.Second)
}

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                     url,
		MaxReconnects:           DefaultMaxReconnects,
		ReconnectWait:           3 * time.Second,
		JSStreamName:            DefaultStreamName,
		JSStreamStorageType:     StorageTypeMemory,
		JSStreamRetentionPolicy: RetentionPolicyInterest,
	}
}

// getJetStreamClient creates a client with JetStream context, or fails the caller test.
func getJetStreamClient(t *testing.T, serverURL string) *jetStreamClient {
	conn, err := nats.Connect(serverURL)
	if err != nil {
		t.Error(err.Error())
	}
	jsCtx, err := conn.JetStream()
	if err != nil {
		conn.Close()
		t.Error(err.Error())
	}
	return &jetStreamClient{
		JetStreamContext: jsCtx,
		natsConn:         conn,
	}
}

// setupTestEnvironment is a TestEnvironment constructor.
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer, natsPort, err := natstesting.StartNATSServer(evtesting.WithJetStreamEnabled())
	require.NoError(t, err)
	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	jsClient := getJetStreamClient(t, natsConfig.URL)
	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	jsBackend := NewJetStream(natsConfig, metricsCollector, jsCleaner, defaultSubsConfig, defaultLogger)

	return &TestEnvironment{
		jsBackend:  jsBackend,
		logger:     defaultLogger,
		natsServer: natsServer,
		jsClient:   jsClient,
		natsConfig: natsConfig,
		cleaner:    jsCleaner,
		natsPort:   natsPort,
	}
}
