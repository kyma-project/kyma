package mock

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/xerrors"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype/eventtypetest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/beb"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/handlertest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

const (
	defaultTokenEndpoint         = "/token"
	defaultEventsEndpoint        = "/events"
	defaultEventsHTTP400Endpoint = "/events400"
)

// BEBHandlerMock represents a mock for the beb.Handler.
type BEBHandlerMock struct {
	ctx                 context.Context
	cfg                 *env.BEBConfig
	logger              *logger.Logger
	collector           *metrics.Collector
	livenessEndpoint    string
	readinessEndpoint   string
	legacyTransformer   *legacy.Transformer
	subscribedProcessor *subscribed.Processor
	mockServer          *testingutils.MockServer
	eventTypeCleaner    eventtype.Cleaner
}

// BEBHandlerMockOpt represents a BEBHandlerMock option.
type BEBHandlerMockOpt func(*BEBHandlerMock)

// GetPort returns the port used by the BEBHandlerMock.
func (m *BEBHandlerMock) GetPort() int {
	return m.cfg.Port
}

// GetMetricsCollector returns the metrics.Collector used by the BEBHandlerMock.
func (m *BEBHandlerMock) GetMetricsCollector() *metrics.Collector {
	return m.collector
}

// Close closes the testing.MockServer used by the BEBHandlerMock.
func (m *BEBHandlerMock) Close() {
	m.mockServer.Close()
}

// GetLivenessEndpoint returns the liveness endpoint used by the BEBHandlerMock.
func (m *BEBHandlerMock) GetLivenessEndpoint() string {
	return m.livenessEndpoint
}

// GetReadinessEndpoint returns the readiness endpoint used by the BEBHandlerMock.
func (m *BEBHandlerMock) GetReadinessEndpoint() string {
	return m.readinessEndpoint
}

// StartOrDie starts a new BEBHandlerMock instance or die if a precondition fails.
// Preconditions: 1) beb.Handler started without errors.
func StartOrDie(ctx context.Context, t *testing.T, requestSize int, eventTypePrefix, eventsEndpoint string,
	requestTimeout, serverResponseTime time.Duration, opts ...BEBHandlerMockOpt) *BEBHandlerMock {
	mockServer := testingutils.NewMockServer(testingutils.WithResponseTime(serverResponseTime))
	mockServer.Start(t, defaultTokenEndpoint, defaultEventsEndpoint, defaultEventsHTTP400Endpoint)

	cfg := testingutils.NewEnvConfig(
		fmt.Sprintf("%s%s", mockServer.URL(), eventsEndpoint),
		fmt.Sprintf("%s%s", mockServer.URL(), defaultTokenEndpoint),
		testingutils.WithPort(testingutils.GeneratePortOrDie()),
		testingutils.WithBEBNamespace(testingutils.MessagingNamespace),
		testingutils.WithRequestTimeout(requestTimeout),
		testingutils.WithEventTypePrefix(eventTypePrefix),
	)

	mockedLogger, err := logger.New("json", "info")
	require.NoError(t, err)

	mock := &BEBHandlerMock{
		ctx:                 ctx,
		cfg:                 cfg,
		logger:              mockedLogger,
		collector:           metrics.NewCollector(),
		livenessEndpoint:    fmt.Sprintf("http://localhost:%d%s", cfg.Port, health.LivenessURI),
		readinessEndpoint:   fmt.Sprintf("http://localhost:%d%s", cfg.Port, health.ReadinessURI),
		legacyTransformer:   &legacy.Transformer{},
		subscribedProcessor: &subscribed.Processor{},
		mockServer:          mockServer,
		eventTypeCleaner:    eventtypetest.CleanerFunc(eventtypetest.DefaultCleaner),
	}

	for _, opt := range opts {
		opt(mock)
	}

	client := oauth.NewClient(ctx, mock.cfg)
	defer client.CloseIdleConnections()

	msgReceiver := receiver.NewHTTPMessageReceiver(mock.cfg.Port)
	msgSender := sender.NewBEBMessageSender(mock.cfg.EmsPublishURL, client)
	msgHandlerOpts := &options.Options{MaxRequestSize: int64(requestSize)}
	msgHandler := beb.NewHandler(
		msgReceiver,
		msgSender,
		mock.cfg.RequestTimeout,
		mock.legacyTransformer,
		msgHandlerOpts,
		mock.subscribedProcessor,
		mock.logger,
		mock.collector,
		mock.eventTypeCleaner,
	)

	go func() { require.NoError(t, msgHandler.Start(ctx)) }()
	testingutils.WaitForEndpointStatusCodeOrFail(mock.livenessEndpoint, health.StatusCodeHealthy)

	return mock
}

// validateEventTypeContainsApplicationName extracts the cloud event type from the http.Request and validates that
// it contains the given application name.
func validateEventTypeContainsApplicationName(name string) testingutils.Validator {
	return func(r *http.Request) error {
		eventType, err := extractEventTypeFromRequest(r)
		if err != nil {
			return err
		}
		if !strings.Contains(eventType, name) {
			return xerrors.Errorf("event-type:%s does not contain application name:%s", eventType, name)
		}
		return nil
	}
}

// extractEventTypeFromRequest returns the cloud event type from the given http.Request.
func extractEventTypeFromRequest(r *http.Request) (string, error) {
	// structured
	if r.Header.Get(cev2http.ContentType) == cev2event.ApplicationCloudEventsJSON {
		message := cev2http.NewMessageFromHttpRequest(r)
		defer func() { _ = message.Finish(nil) }()
		event, err := binding.ToEvent(context.Background(), message)
		if err != nil {
			return "", err
		}
		return event.Type(), nil
	}

	// binary
	eventType := r.Header.Get(testingutils.CeTypeHeader)
	if strings.TrimSpace(eventType) == "" {
		return "", errors.New("event-type header is not found or empty")
	}
	return eventType, nil
}

// WithEventTypePrefix returns BEBHandlerMockOpt which sets the eventTypePrefix for the given BEBHandlerMock.
func WithEventTypePrefix(eventTypePrefix string) BEBHandlerMockOpt {
	return func(m *BEBHandlerMock) {
		m.cfg.EventTypePrefix = eventTypePrefix
	}
}

// WithApplication returns BEBHandlerMockOpt which sets the subscribed.Processor for the given BEBHandlerMock.
func WithApplication(applicationNameToCreate, applicationNameToValidate string) BEBHandlerMockOpt {
	return func(m *BEBHandlerMock) {
		applicationLister := handlertest.NewApplicationListerOrDie(m.ctx, applicationNameToCreate)
		m.legacyTransformer = legacy.NewTransformer(m.cfg.BEBNamespace, m.cfg.EventTypePrefix, applicationLister)
		validator := validateEventTypeContainsApplicationName(applicationNameToValidate)
		testingutils.WithValidator(validator)(m.mockServer)
		m.eventTypeCleaner = eventtype.NewCleaner(m.cfg.EventTypePrefix, applicationLister, m.logger)
	}
}

// WithSubscription returns BEBHandlerMockOpt which sets the subscribed.Processor for the given BEBHandlerMock.
func WithSubscription(scheme *runtime.Scheme, subscription *eventingv1alpha1.Subscription) BEBHandlerMockOpt {
	return func(m *BEBHandlerMock) {
		dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, subscription)
		informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, time.Second, v1.NamespaceAll, nil)
		genericInformer := informerFactory.ForResource(subscribed.GVR)
		mockedLogger, _ := logger.New("json", "info")
		informers.WaitForCacheSyncOrDie(m.ctx, informerFactory, mockedLogger)
		subscriptionLister := genericInformer.Lister()
		m.subscribedProcessor = &subscribed.Processor{SubscriptionLister: &subscriptionLister, Config: m.cfg, Logger: m.logger}
	}
}
