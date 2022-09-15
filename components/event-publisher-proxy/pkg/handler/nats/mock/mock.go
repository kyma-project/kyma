package mock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/nats-io/nats-server/v2/server"
	natsio "github.com/nats-io/nats.go"
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
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// NATSHandlerMock represents a mock for the nats.Handler.
type NATSHandlerMock struct {
	ctx                 context.Context
	handler             *nats.Handler
	livenessEndpoint    string
	readinessEndpoint   string
	eventTypePrefix     string
	logger              *logger.Logger
	natsServer          *server.Server
	jetstreamMode       bool
	natsConfig          *env.NATSConfig
	collector           *metrics.Collector
	legacyTransformer   *legacy.Transformer
	subscribedProcessor *subscribed.Processor
	eventTypeCleaner    eventtype.Cleaner
	connection          *natsio.Conn
}

// NATSHandlerMockOpt represents a NATSHandlerMock option.
type NATSHandlerMockOpt func(*NATSHandlerMock)

// StartOrDie starts a new NATSHandlerMock instance or die if a precondition fails.
// Preconditions: 1) NATS connection and 2) nats.Handler started without errors.
func StartOrDie(ctx context.Context, t *testing.T, opts ...NATSHandlerMockOpt) *NATSHandlerMock {
	port := testingutils.GeneratePortOrDie()

	mockedLogger, err := logger.New("json", "info")
	require.NoError(t, err)

	mock := &NATSHandlerMock{
		ctx:                 ctx,
		livenessEndpoint:    fmt.Sprintf("http://localhost:%d%s", port, health.LivenessURI),
		readinessEndpoint:   fmt.Sprintf("http://localhost:%d%s", port, health.ReadinessURI),
		logger:              mockedLogger,
		natsConfig:          newNATSConfig(port),
		collector:           metrics.NewCollector(),
		legacyTransformer:   &legacy.Transformer{},
		subscribedProcessor: &subscribed.Processor{},
		eventTypeCleaner:    eventtypetest.CleanerFunc(eventtypetest.DefaultCleaner),
	}

	for _, opt := range opts {
		opt(mock)
	}
	mock.natsServer = testingutils.StartNATSServer(mock.jetstreamMode)

	msgReceiver := receiver.NewHTTPMessageReceiver(mock.natsConfig.Port)

	connection, err := testingutils.ConnectToNATSServer(mock.GetNATSURL())
	require.NoError(t, err)
	mock.connection = connection

	//nolint:gosimple
	var msgSender sender.GenericSender
	msgSender = sender.NewNATSMessageSender(ctx, mock.connection, mock.logger)

	mock.handler = nats.NewHandler(
		msgReceiver,
		&msgSender,
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

// Stop closes the sender.NATSMessageSender connection and calls the NATSHandlerMock.ShutdownNATSServerAndWait.
func (m *NATSHandlerMock) Stop() {
	m.connection.Close()
	m.ShutdownNATSServerAndWait()
}

// ShutdownNATSServerAndWait shuts down the NATS server used by the NATSHandlerMock and waits for the shutdown.
func (m *NATSHandlerMock) ShutdownNATSServerAndWait() {
	m.natsServer.Shutdown()
	m.natsServer.WaitForShutdown()
}

// GetNATSURL returns the NATS server URL used by the NATSHandlerMock.
func (m *NATSHandlerMock) GetNATSURL() string {
	return m.natsServer.ClientURL()
}

// GetLivenessEndpoint returns the liveness endpoint used by the NATSHandlerMock.
func (m *NATSHandlerMock) GetLivenessEndpoint() string {
	return m.livenessEndpoint
}

// GetReadinessEndpoint returns the readiness endpoint used by the NATSHandlerMock.
func (m *NATSHandlerMock) GetReadinessEndpoint() string {
	return m.readinessEndpoint
}

// GetHandler returns the nats.Handler used by the NATSHandlerMock.
func (m *NATSHandlerMock) GetHandler() *nats.Handler {
	return m.handler
}

// GetMetricsCollector returns the metrics.Collector used by the NATSHandlerMock.
func (m *NATSHandlerMock) GetMetricsCollector() *metrics.Collector {
	return m.collector
}

// GetNATSConfig returns the env.NATSConfig used by the NATSHandlerMock.
func (m *NATSHandlerMock) GetNATSConfig() *env.NATSConfig {
	return m.natsConfig
}

// WithEventTypePrefix returns NATSHandlerMockOpt which sets the eventTypePrefix for the given NATSHandlerMock.
func WithEventTypePrefix(eventTypePrefix string) NATSHandlerMockOpt {
	return func(m *NATSHandlerMock) {
		m.eventTypePrefix = eventTypePrefix
	}
}

// WithSubscription returns NATSHandlerMockOpt which sets the subscribed.Processor for the given NATSHandlerMock.
func WithSubscription(scheme *runtime.Scheme, subscription *eventingv1alpha1.Subscription) NATSHandlerMockOpt {
	return func(m *NATSHandlerMock) {
		m.natsConfig.EventTypePrefix = m.eventTypePrefix
		dynamicTestClient := dynamicfake.NewSimpleDynamicClient(scheme, subscription)
		dFilteredSharedInfFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicTestClient, 10*time.Second, v1.NamespaceAll, nil)
		genericInf := dFilteredSharedInfFactory.ForResource(subscribed.GVR)
		mockedLogger, _ := logger.New("json", "info")
		informers.WaitForCacheSyncOrDie(m.ctx, dFilteredSharedInfFactory, mockedLogger)
		subLister := genericInf.Lister()
		m.subscribedProcessor = &subscribed.Processor{
			SubscriptionLister: &subLister,
			Config:             m.natsConfig.ToConfig(),
			Logger:             m.logger,
		}
	}
}

// WithApplication returns NATSHandlerMockOpt which sets the legacy.Transformer for the given NATSHandlerMock.
func WithApplication(applicationName string) NATSHandlerMockOpt {
	return func(m *NATSHandlerMock) {
		applicationLister := handlertest.NewApplicationListerOrDie(m.ctx, applicationName)
		m.legacyTransformer = legacy.NewTransformer(
			m.natsConfig.ToConfig().BEBNamespace,
			m.eventTypePrefix,
			applicationLister,
		)
		m.eventTypeCleaner = eventtype.NewCleaner(m.eventTypePrefix, applicationLister, m.logger)
	}
}

// WithJetstream returns NATSHandlerMockOpt which starts the NATS server in the jetstream mode for the given NATSHandlerMock.
func WithJetstream(jsEnabled bool) NATSHandlerMockOpt {
	return func(m *NATSHandlerMock) {
		m.jetstreamMode = jsEnabled
	}
}

func newNATSConfig(port int) *env.NATSConfig {
	return &env.NATSConfig{
		Port:            port,
		LegacyNamespace: testingutils.MessagingNamespace,
		EventTypePrefix: testingutils.MessagingEventTypePrefix,
		JSStreamName:    testingutils.StreamName,
	}
}
