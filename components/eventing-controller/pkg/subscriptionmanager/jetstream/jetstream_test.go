package jetstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	data                  = "sampledata"
	subscriptionName      = "test"
	subscriptionNamespace = "test"
)

var expectedDataInStore = fmt.Sprintf("%q", data)

func TestCleanup(t *testing.T) {
	// given
	testEnv := setUpTestEnvironment(t)
	defer controllertesting.ShutDownNATSServer(testEnv.natsServer)
	defer testEnv.subscriber.Shutdown()
	err := testEventing(testEnv)
	require.NoError(t, err)
	consumersEquals(t, 1, testEnv)

	// when
	err = cleanup(testEnv.jsBackend, testEnv.dynamicClient, testEnv.defaultLogger.WithContext())

	// then
	require.NoError(t, err)
	gotSub := getK8sSubscription(t, testEnv)
	wantSubStatus := eventingv1alpha1.SubscriptionStatus{}
	require.Equal(t, wantSubStatus, gotSub.Status)
	// test JetStream subscriptions/consumers are gone
	consumersEquals(t, 0, testEnv)
	// eventing should fail
	err = testEventing(testEnv)
	require.Error(t, err)
}

// utilities and helper functions

type TestEnvironment struct {
	ctx           context.Context
	dynamicClient dynamic.Interface
	jsBackend     *handlers.JetStream
	jsCtx         nats.JetStreamContext
	natsServer    *server.Server
	subscriber    *controllertesting.Subscriber
	envConf       env.NatsConfig
	defaultLogger *logger.Logger
}

func getK8sSubscription(t *testing.T, testEnv *TestEnvironment) *eventingv1alpha1.Subscription {
	unstructuredSub, err := testEnv.dynamicClient.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace(
		subscriptionNamespace).Get(testEnv.ctx, subscriptionName, metav1.GetOptions{})
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
		JSStreamStorageType:     handlers.JetStreamStorageTypeMemory,
		JSStreamRetentionPolicy: handlers.JetStreamRetentionPolicyInterest,
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

func createAndSyncSubscription(t *testing.T, sinkURL string, jsBackend *handlers.JetStream) *eventingv1alpha1.Subscription {
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
	cleanedSubjects, _ := handlers.GetCleanSubjects(testSub, eventtype.CleanerFunc(cleaner))
	testSub.Status.CleanEventTypes = jsBackend.GetJetStreamSubjects(cleanedSubjects)

	// create NATS subscription
	err := jsBackend.SyncSubscription(testSub)
	require.NoError(t, err)
	return testSub
}

func testEventing(testEnv *TestEnvironment) error {
	// make sure subscriber works
	err := testEnv.subscriber.CheckEvent("")
	if err != nil {
		return err
	}

	// send an event
	err = handlers.SendEventToJetStream(testEnv.jsBackend, data)
	if err != nil {
		return err
	}

	// check for the event
	err = testEnv.subscriber.CheckEvent(expectedDataInStore)
	if err != nil {
		return err
	}
	return nil
}

func consumersEquals(t *testing.T, length int, testEnv *TestEnvironment) {
	// verify that the number of consumers is one
	info, err := testEnv.jsCtx.StreamInfo(testEnv.envConf.JSStreamName)
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

	// Create an instance of the JetStream Backend
	jsBackend := handlers.NewJetStream(envConf, defaultLogger)

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
