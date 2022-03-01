package mock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

// NatsHandlerMock represents a mock for the nats.Handler.
type NatsHandlerMock struct {
	ctx                 context.Context
	handler             *nats.Handler
	livenessEndpoint    string
	readinessEndpoint   string
	logger              *logrus.Logger
	natsServer          *server.Server
	natsConfig          *env.NatsConfig
	collector           *metrics.Collector
	legacyTransformer   *legacy.Transformer
	subscribedProcessor *subscribed.Processor
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
		natsConfig:          &env.NatsConfig{Port: port},
		collector:           metrics.NewCollector(),
		legacyTransformer:   &legacy.Transformer{},
		subscribedProcessor: &subscribed.Processor{},
	}

	for _, opt := range opts {
		opt(mock)
	}

	msgReceiver := receiver.NewHTTPMessageReceiver(mock.natsConfig.Port)

	backendConnection := pkgnats.NewBackendConnection(mock.GetNatsURL(), true, 3, time.Second)
	err := backendConnection.Connect()
	require.NoError(t, err)

	msgSender := sender.NewNatsMessageSender(ctx, backendConnection, mock.logger)

	mock.handler = nats.NewHandler(
		msgReceiver,
		msgSender,
		mock.natsConfig.RequestTimeout,
		mock.legacyTransformer,
		&options.Options{MaxRequestSize: 65536},
		mock.subscribedProcessor,
		mock.logger,
		mock.collector,
	)

	go func() { require.NoError(t, mock.handler.Start(ctx)) }()
	testingutils.WaitForEndpointStatusCodeOrFail(mock.livenessEndpoint, health.StatusCodeHealthy)

	return mock
}

// ShutdownNatsServer shuts down the NATS server used by the NatsHandlerMock.
func (m *NatsHandlerMock) ShutdownNatsServer() {
	m.natsServer.Shutdown()
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
