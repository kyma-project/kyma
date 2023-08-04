package jetstream

import (
	"sync"

	backendutilsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"

	cev2 "github.com/cloudevents/sdk-go/v2"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
)

const (
	separator = "/"
)

//go:generate mockery --name Backend
type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(connCloseHandler backendutilsv2.ConnClosedHandler) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of JetStream.
	SyncSubscription(subscription *eventingv1alpha2.Subscription) error

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha2.Subscription) error

	// DeleteSubscriptionsOnly should delete the JetStream subscriptions only.
	// The corresponding JetStream consumers of the subscriptions must not be deleted.
	DeleteSubscriptionsOnly(subscription *eventingv1alpha2.Subscription) error

	// GetJetStreamSubjects returns a list of subjects appended with stream name and source as prefix if needed
	GetJetStreamSubjects(source string, subjects []string, typeMatching eventingv1alpha2.TypeMatching) []string

	// DeleteInvalidConsumers deletes all JetStream consumers having no subscription types in subscription resources
	DeleteInvalidConsumers(subscriptions []eventingv1alpha2.Subscription) error

	// GetJetStreamContext returns the current JetStreamContext
	GetJetStreamContext() nats.JetStreamContext
}

type JetStream struct {
	Config        env.NATSConfig
	Conn          *nats.Conn
	jsCtx         nats.JetStreamContext
	client        cev2.Client
	subscriptions map[SubscriptionSubjectIdentifier]Subscriber
	sinks         sync.Map
	// connClosedHandler gets called by the NATS server when Conn is closed and retry attempts are exhausted.
	connClosedHandler backendutilsv2.ConnClosedHandler
	logger            *logger.Logger
	metricsCollector  *backendmetrics.Collector
	cleaner           cleaner.Cleaner
	subsConfig        env.DefaultSubscriptionConfig
}

type Subscriber interface {
	SubscriptionSubject() string
	ConsumerInfo() (*nats.ConsumerInfo, error)
	IsValid() bool
	Unsubscribe() error
	SetPendingLimits(int, int) error
	PendingLimits() (int, int, error)
}

type Subscription struct {
	*nats.Subscription
}

// SubscriptionSubjectIdentifier is used to uniquely identify a Subscription subject.
// It should be used only with JetStream backend.
type SubscriptionSubjectIdentifier struct {
	consumerName, namespacedSubjectName string
}

type DefaultSubOpts []nats.SubOpt

//----------------------------------------
// JetStream Backend Test Types
//----------------------------------------

type jetStreamClient struct {
	nats.JetStreamContext
	natsConn *nats.Conn
}

func (js Subscription) SubscriptionSubject() string {
	return js.Subject
}
