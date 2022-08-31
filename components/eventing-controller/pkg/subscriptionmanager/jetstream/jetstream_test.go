package jetstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/metrics"
	nats2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats"
	jetstream2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats/jetstream"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	subscriptionName      = "test"
	subscriptionNamespace = "test"
)

func TestCleanup(t *testing.T) {
	// given
	testEnv := setUpTestEnvironment(t)
	defer controllertesting.ShutDownNATSServer(testEnv.natsServer)
	defer testEnv.subscriber.Shutdown()
	err := testEnv.testSendingAndReceivingAnEvent()
	require.NoError(t, err)
	testEnv.consumersEquals(t, 1)

	// when
	err = cleanup(testEnv.jsBackend, testEnv.dynamicClient, testEnv.defaultLogger.WithContext())

	// then
	require.NoError(t, err)
	gotSub := testEnv.getK8sSubscription(t)
	wantSubStatus := eventingv1alpha1.SubscriptionStatus{CleanEventTypes: []string{}}
	require.Equal(t, wantSubStatus, gotSub.Status)
	// test JetStream subscriptions/consumers are gone
	testEnv.consumersEquals(t, 0)
	// eventing should fail
	err = testEnv.testSendingAndReceivingAnEvent()
	require.Error(t, err)
}

// utilities and helper functions

type TestEnvironment struct {
	ctx           context.Context
	dynamicClient dynamic.Interface
	jsBackend     *jetstream2.JetStream
	jsCtx         nats.JetStreamContext
	natsServer    *server.Server
	subscriber    *controllertesting.Subscriber
	envConf       env.NatsConfig
	defaultLogger *logger.Logger
}

func (te *TestEnvironment) getK8sSubscription(t *testing.T) *eventingv1alpha1.Subscription {
	unstructuredSub, err := te.dynamicClient.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace(
		subscriptionNamespace).Get(te.ctx, subscriptionName, metav1.GetOptions{})
	require.NoError(t, err)
	subscription, err := controllertesting.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	return subscription
}

func getJetStreamClient(t *testing.T, natsURL string) nats.JetStreamContext {
	conn, err := nats.Connect(natsURL)
	require.NoError(t, err)
	jsClient, err := conn.JetStream()
	require.NoError(t, err)
	return jsClient
}

func getNATSConf(natsURL string) env.NatsConfig {
	return env.NatsConfig{
		URL:                     natsURL,
		MaxReconnects:           10,
		ReconnectWait:           time.Second,
		EventTypePrefix:         controllertesting.EventTypePrefix,
		JSStreamName:            controllertesting.EventTypePrefix,
		JSStreamStorageType:     jetstream2.JetStreamStorageTypeMemory,
		JSStreamRetentionPolicy: jetstream2.JetStreamRetentionPolicyInterest,
	}
}

func setUpNATSServer(t *testing.T) *server.Server {
	// create NATS Server with JetStream enabled
	natsPort, err := controllertesting.GetFreePort()
	require.NoError(t, err)
	natsServer := controllertesting.RunNatsServerOnPort(
		controllertesting.WithPort(natsPort),
		controllertesting.WithJetStreamEnabled())
	return natsServer
}

func createAndSyncSubscription(t *testing.T, sinkURL string, jsBackend *jetstream2.JetStream) *eventingv1alpha1.Subscription {
	subsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	// create test subscription
	testSub := controllertesting.NewSubscription(
		subscriptionName, subscriptionNamespace,
		controllertesting.WithFakeSubscriptionStatus(),
		controllertesting.WithOrderCreatedFilter(),
		controllertesting.WithSinkURL(sinkURL),
		controllertesting.WithStatusConfig(subsConfig),
	)

	cleaner := func(et string) (string, error) {
		return et, nil
	}
	cleanEventTypes, err := nats2.GetCleanSubjects(testSub, eventtype.CleanerFunc(cleaner))
	require.NoError(t, err)
	testSub.Status.CleanEventTypes = cleanEventTypes

	// create NATS subscription
	err = jsBackend.SyncSubscription(testSub)
	require.NoError(t, err)
	return testSub
}

func (te *TestEnvironment) testSendingAndReceivingAnEvent() error {
	data := "sampledata"
	expectedDataInStore := fmt.Sprintf("%q", data)

	// make sure subscriber is reachable via http
	err := te.subscriber.CheckEvent("")
	if err != nil {
		return err
	}

	// send an event
	err = jetstream2.SendEventToJetStream(te.jsBackend, data)

	if err != nil {
		return err
	}

	// check for the event
	err = te.subscriber.CheckEvent(expectedDataInStore)
	if err != nil {
		return err
	}
	return nil
}

func (te *TestEnvironment) consumersEquals(t *testing.T, length int) {
	// verify that the number of consumers is one
	info, err := te.jsCtx.StreamInfo(te.envConf.JSStreamName)
	require.NoError(t, err)
	require.Equal(t, info.State.Consumers, length)
}

func setUpTestEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()
	// create a test subscriber and natsServer
	subscriber := controllertesting.NewSubscriber()
	natsServer := setUpNATSServer(t)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	envConf := getNATSConf(natsServer.ClientURL())
	jsClient := getJetStreamClient(t, natsServer.ClientURL())

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	// Create an instance of the JetStream Backend
	jsBackend := jetstream2.NewJetStream(envConf, metricsCollector, defaultLogger)

	// Initialize JetStream Backend
	err = jsBackend.Initialize(nil)
	require.NoError(t, err)

	testSub := createAndSyncSubscription(t, subscriber.SinkURL, jsBackend)
	// create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(testSub)
	require.NoError(t, err)

	return &TestEnvironment{
		ctx:           ctx,
		dynamicClient: fakeClient,
		jsBackend:     jsBackend,
		jsCtx:         jsClient,
		natsServer:    natsServer,
		subscriber:    subscriber,
		envConf:       envConf,
		defaultLogger: defaultLogger,
	}
}
