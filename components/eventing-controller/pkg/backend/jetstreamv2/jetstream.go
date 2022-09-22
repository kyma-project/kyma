package jetstreamv2

import (
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

var _ Backend = &JetStream{}

type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(connCloseHandler backendnats.ConnClosedHandler) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of JetStream.
	SyncSubscription(subscription *eventingv1alpha2.Subscription) error

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha2.Subscription) error

	// GetJetStreamSubjects returns a list of subjects appended with stream name as prefix if needed
	GetJetStreamSubjects(subjects []string) []string
}

type JetStream struct {
	Config           env.NatsConfig
	subscriptions    map[string]backendnats.Subscriber
	logger           *logger.Logger
	metricsCollector *backendmetrics.Collector
}

func NewJetStream(config env.NatsConfig, metricsCollector *backendmetrics.Collector, logger *logger.Logger) *JetStream {
	return &JetStream{
		Config:           config,
		logger:           logger,
		subscriptions:    make(map[string]backendnats.Subscriber),
		metricsCollector: metricsCollector,
	}
}

func (js *JetStream) Initialize(connCloseHandler backendnats.ConnClosedHandler) error {
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
