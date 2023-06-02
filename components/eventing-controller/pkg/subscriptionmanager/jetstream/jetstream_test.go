package jetstream

import (
	"context"
	"fmt"
	"testing"
	"time"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
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
	// a consumer should exist on JetStream
	testEnv.consumersEquals(t, 1)

	// when
	err := cleanupv2(testEnv.jsBackend, testEnv.dynamicClient, testEnv.defaultLogger.WithContext())

	// then
	require.NoError(t, err)
	// the consumer on JetStream should have being deleted
	testEnv.consumersEquals(t, 0)
}

// utilities and helper functions

type TestEnvironment struct {
	ctx           context.Context
	dynamicClient dynamic.Interface
	jsBackend     *backendnats.JetStream
	jsCtx         nats.JetStreamContext
	natsServer    *server.Server
	subscriber    *controllertesting.Subscriber
	envConf       env.NATSConfig
	defaultLogger *logger.Logger
}

func getJetStreamClient(t *testing.T, natsURL string) nats.JetStreamContext {
	conn, err := nats.Connect(natsURL)
	require.NoError(t, err)
	jsClient, err := conn.JetStream()
	require.NoError(t, err)
	return jsClient
}

func getNATSConf(natsURL string, natsPort int) env.NATSConfig {
	streamName := fmt.Sprintf("%s%d", controllertesting.JSStreamName, natsPort)
	return env.NATSConfig{
		URL:                     natsURL,
		MaxReconnects:           10,
		ReconnectWait:           time.Second,
		EventTypePrefix:         controllertesting.EventTypePrefix,
		JSStreamName:            streamName,
		JSSubjectPrefix:         streamName,
		JSStreamStorageType:     backendnats.StorageTypeMemory,
		JSStreamRetentionPolicy: backendnats.RetentionPolicyInterest,
		JSStreamDiscardPolicy:   backendnats.DiscardPolicyNew,
	}
}

func createAndSyncSubscription(t *testing.T, sinkURL string,
	jsBackend *backendnats.JetStream) *eventingv1alpha2.Subscription {
	// create test subscription
	testSub := controllertesting.NewSubscription(
		subscriptionName, subscriptionNamespace,
		controllertesting.WithSource(controllertesting.EventSourceClean),
		controllertesting.WithSinkURL(sinkURL),
		controllertesting.WithOrderCreatedV1Event(),
		controllertesting.WithStatusTypes([]eventingv1alpha2.EventType{
			{
				OriginalType: controllertesting.OrderCreatedV1Event,
				CleanType:    controllertesting.OrderCreatedV1Event,
			},
		}),
	)

	// create NATS subscription
	err := jsBackend.SyncSubscription(testSub)
	require.NoError(t, err)
	return testSub
}

func (te *TestEnvironment) consumersEquals(t *testing.T, length int) {
	// verify that the number of consumers is one
	info, err := te.jsCtx.StreamInfo(te.envConf.JSStreamName)
	require.NoError(t, err)
	require.Equal(t, length, info.State.Consumers)
}

func setUpTestEnvironment(t *testing.T) *TestEnvironment {
	ctx := context.Background()
	// create a test subscriber and natsServer
	subscriber := controllertesting.NewSubscriber()
	// create NATS Server with JetStream enabled
	natsPort, err := controllertesting.GetFreePort()
	require.NoError(t, err)
	natsServer := controllertesting.StartDefaultJetStreamServer(natsPort)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	envConf := getNATSConf(natsServer.ClientURL(), natsPort)
	jsClient := getJetStreamClient(t, natsServer.ClientURL())

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}

	// Create an instance of the JetStream Backend
	jsBackend := backendnats.NewJetStream(envConf, metricsCollector, jsCleaner, defaultSubsConfig, defaultLogger)

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
