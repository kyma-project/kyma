package jetstream

import (
	"sync"

	cev2 "github.com/cloudevents/sdk-go/v2"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnatsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/natsv2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName                      = "jetstream-handler"
	MissingNATSSubscriptionMsg         = "failed to create NATS JetStream subscription"
	MissingNATSSubscriptionMsgWithInfo = MissingNATSSubscriptionMsg + " for subject: %v"
)

type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(connCloseHandler backendnatsv2.ConnClosedHandler) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of JetStream.
	SyncSubscription(subscription *eventingv1alpha2.Subscription) error

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha2.Subscription) error

	// GetJetStreamSubjects returns a list of subjects appended with stream name as prefix if needed
	GetJetStreamSubjects(subjects []string) []string
}

// SubscriptionSubjectIdentifier is used to uniquely identify a Subscription subject.
// It should be used only with JetStream backend.
type SubscriptionSubjectIdentifier struct {
	consumerName, namespacedSubjectName string
}

type JetStream struct {
	Config        env.NatsConfig
	conn          *nats.Conn
	jsCtx         nats.JetStreamContext
	client        cev2.Client
	subscriptions map[SubscriptionSubjectIdentifier]backendnatsv2.Subscriber
	sinks         sync.Map
	// connClosedHandler gets called by the NATS server when conn is closed and retry attempts are exhausted.
	connClosedHandler backendnatsv2.ConnClosedHandler
	logger            *logger.Logger
	metricsCollector  *backendmetrics.Collector
}

func NewJetStream(config env.NatsConfig, metricsCollector *backendmetrics.Collector, logger *logger.Logger) *JetStream {
	return &JetStream{
		Config:           config,
		logger:           logger,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]backendnatsv2.Subscriber),
		metricsCollector: metricsCollector,
	}
}

func (js *JetStream) Initialize(connCloseHandler backendnatsv2.ConnClosedHandler) error {
	return nil
}

func (js *JetStream) SyncSubscription(_ *eventingv1alpha2.Subscription) error {
	return nil
}

func (js *JetStream) DeleteSubscription(_ *eventingv1alpha2.Subscription) error {
	return nil
}

// GetJetStreamSubjects returns a list of subjects appended with prefix if needed.
func (js *JetStream) GetJetStreamSubjects(_ []string) []string {
	var result []string
	return result
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
