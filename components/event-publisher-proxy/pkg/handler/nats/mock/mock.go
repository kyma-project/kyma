package mock

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/nats-io/nats-server/v2/server"
	natsio "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype/eventtypetest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// NatsHandlerMock represents a mock for the nats.Handler.
type NatsHandlerMock struct {
	ctx                 context.Context
	handler             *nats.Handler
	livenessEndpoint    string
	readinessEndpoint   string
	eventTypePrefix     string
	logger              *logrus.Logger
	natsServer          *server.Server
	natsConfig          *env.NatsConfig
	collector           *metrics.Collector
	legacyTransformer   *legacy.Transformer
	subscribedProcessor *subscribed.Processor
	eventTypeCleaner    eventtype.Cleaner
	connection          *natsio.Conn
}

// NatsHandlerMockOpt represents a NatsHandlerMock option.
type NatsHandlerMockOpt func(*NatsHandlerMock)

// StartOrDie starts a new NatsHandlerMock instance or die if a precondition fails.
// Preconditions: 1) NATS connection and 2) nats.Handler started without errors.
func StartOrDie(ctx context.Context, t *testing.T, opts ...NatsHandlerMockOpt) *NatsHandlerMock {
	port := testingutils.GeneratePortOrDie()
	mock := &NatsHandlerMock{
		ctx:                 ctx,
		livenessEndpoint:    fmt.Sprintf("http://localhost:%d%s", port, health.LivenessURI),
		readinessEndpoint:   fmt.Sprintf("http://localhost:%d%s", port, health.ReadinessURI),
		logger:              logrus.New(),
		natsServer:          testingutils.StartNatsServer(),
		natsConfig:          newNatsConfig(port),
		collector:           metrics.NewCollector(),
		legacyTransformer:   &legacy.Transformer{},
		subscribedProcessor: &subscribed.Processor{},
		eventTypeCleaner:    eventtypetest.CleanerFunc(eventtypetest.DefaultCleaner),
	}

	for _, opt := range opts {
		opt(mock)
	}

	msgReceiver := receiver.NewHTTPMessageReceiver(mock.natsConfig.Port)

	connection, err := pkgnats.Connect(mock.GetNatsURL(),
		pkgnats.WithRetryOnFailedConnect(true),
		pkgnats.WithMaxReconnects(3),
		pkgnats.WithReconnectWait(time.Second),
	)
	require.NoError(t, err)
	mock.connection = connection

	msgSender := sender.NewNatsMessageSender(ctx, mock.connection, mock.logger)

	mock.handler = nats.NewHandler(
		msgReceiver,
		msgSender,
		mock.natsConfig.RequestTimeout,
		mock.legacyTransformer,
		&options.Options{MaxRequestSize: 65536},
		mock.subscribedProcessor,
		mock.logger,
		mock.collector,
		mock.eventTypeCleaner,
	)

	go func() { require.NoError(t, mock.handler.Start(ctx)) }()
	testingutils.WaitForEndpointStatusCodeOrFail(mock.livenessEndpoint, health.StatusCodeHealthy)

	return mock
}

// Stop closes the sender.NatsMessageSender connection and calls the NatsHandlerMock.ShutdownNatsServerAndWait.
func (m *NatsHandlerMock) Stop() {
	m.connection.Close()
	m.ShutdownNatsServerAndWait()
}

// ShutdownNatsServerAndWait shuts down the NATS server used by the NatsHandlerMock and waits for the shutdown.
func (m *NatsHandlerMock) ShutdownNatsServerAndWait() {
	m.natsServer.Shutdown()
	m.natsServer.WaitForShutdown()
}

// GetNatsURL returns the NATS server URL used by the NatsHandlerMock.
func (m *NatsHandlerMock) GetNatsURL() string {
	return m.natsServer.ClientURL()
}

// GetLivenessEndpoint returns the liveness endpoint used by the NatsHandlerMock.
func (m *NatsHandlerMock) GetLivenessEndpoint() string {
	return m.livenessEndpoint
}

// GetReadinessEndpoint returns the readiness endpoint used by the NatsHandlerMock.
func (m *NatsHandlerMock) GetReadinessEndpoint() string {
	return m.readinessEndpoint
}

// GetHandler returns the nats.Handler used by the NatsHandlerMock.
func (m *NatsHandlerMock) GetHandler() *nats.Handler {
	return m.handler
}

// GetMetricsCollector returns the metrics.Collector used by the NatsHandlerMock.
func (m *NatsHandlerMock) GetMetricsCollector() *metrics.Collector {
	return m.collector
}

// GetNatsConfig returns the env.NatsConfig used by the NatsHandlerMock.
func (m *NatsHandlerMock) GetNatsConfig() *env.NatsConfig {
	return m.natsConfig
}

// WithEventTypePrefix returns NatsHandlerMockOpt which sets the eventTypePrefix for the given NatsHandlerMock.
func WithEventTypePrefix(eventTypePrefix string) NatsHandlerMockOpt {
	return func(m *NatsHandlerMock) {
		m.eventTypePrefix = eventTypePrefix
	}
}

// WithSubscription returns NatsHandlerMockOpt which sets the subscribed.Processor for the given NatsHandlerMock.
func WithSubscription(scheme *runtime.Scheme, subscription *eventingv1alpha1.Subscription) NatsHandlerMockOpt {
	return func(m *NatsHandlerMock) {
		m.natsConfig.LegacyEventTypePrefix = m.eventTypePrefix
		dynamicTestClient := dynamicfake.NewSimpleDynamicClient(scheme, subscription)
		dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicTestClient, 10*time.Second, v1.NamespaceAll, nil)
		genericInf := dFilteredSharedInfFactory.ForResource(subscribed.GVR)
		informers.WaitForCacheSyncOrDie(m.ctx, dFilteredSharedInfFactory)
		subLister := genericInf.Lister()
		m.subscribedProcessor = &subscribed.Processor{
			SubscriptionLister: &subLister,
			Config:             m.natsConfig.ToConfig(),
			Logger:             m.logger,
		}
	}
}

// WithApplication returns NatsHandlerMockOpt which sets the legacy.Transformer for the given NatsHandlerMock.
func WithApplication(applicationName string) NatsHandlerMockOpt {
	return func(m *NatsHandlerMock) {
		applicationLister := handlertest.NewApplicationListerOrDie(m.ctx, applicationName)
		m.legacyTransformer = legacy.NewTransformer(
			m.natsConfig.ToConfig().BEBNamespace,
			m.eventTypePrefix,
			applicationLister,
		)
		m.eventTypeCleaner = eventtype.NewCleaner(m.eventTypePrefix, applicationLister, m.logger)
	}
}

func newNatsConfig(port int) *env.NatsConfig {
	return &env.NatsConfig{
		Port:                  port,
		LegacyNamespace:       testingutils.MessagingNamespace,
		LegacyEventTypePrefix: testingutils.MessagingEventTypePrefix,
	}
}
