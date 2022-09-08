package core

import (
	"context"
	"net/http"
	"sync"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2context "github.com/cloudevents/sdk-go/v2/context"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// compile time check
var _ Backend = &Nats{}

const (
	backoffStrategy = cev2context.BackoffStrategyConstant
	natsHandlerName = "nats-handler"
)

type Backend interface {
	// Initialize connects and initializes the NATS backend.
	// connCloseHandler can be used to register a handler that gets called when connection
	// to the NATS server is closed and retry attempts are exceeded.
	Initialize(connCloseHandler backendnats.ConnClosedHandler) error

	// SyncSubscription synchronizes the Kyma Subscription on the NATS backend.
	SyncSubscription(subscription *eventingv1alpha1.Subscription) error

	// DeleteSubscription deletes the corresponding subscription on the NATS backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

type Nats struct {
	Config            env.NatsConfig
	defaultSubsConfig env.DefaultSubscriptionConfig
	logger            *logger.Logger
	client            cev2.Client
	connection        *nats.Conn
	subscriptions     map[string]*nats.Subscription
	sinks             sync.Map
	connClosedHandler backendnats.ConnClosedHandler
	metricsCollector  *backendmetrics.Collector
}

func NewNats(config env.NatsConfig, subsConfig env.DefaultSubscriptionConfig, metricsCollector *backendmetrics.Collector, logger *logger.Logger) *Nats {
	return &Nats{
		Config:            config,
		defaultSubsConfig: subsConfig,
		logger:            logger,
		subscriptions:     make(map[string]*nats.Subscription),
		metricsCollector:  metricsCollector,
	}
}

// Initialize creates a connection to NATS.
func (n *Nats) Initialize(connCloseHandler backendnats.ConnClosedHandler) (err error) {
	if n.connection == nil || n.connection.Status() != nats.CONNECTED {
		natsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(n.Config.MaxReconnects),
			nats.ReconnectWait(n.Config.ReconnectWait),
		}
		n.connection, err = nats.Connect(n.Config.URL, natsOptions...)
		if err != nil {
			return errors.Wrapf(err, "connect to NATS failed")
		}
		if n.connection.Status() != nats.CONNECTED {
			return errors.Errorf("connect to NATS failed status: %v", n.connection.Status())
		}
		n.connClosedHandler = connCloseHandler
		if n.connClosedHandler != nil {
			n.connection.SetClosedHandler(nats.ConnHandler(n.connClosedHandler))
		}
	}

	if n.client != nil {
		return nil
	}

	if n.client, err = newCloudeventClient(n.Config); err != nil {
		return err
	}

	return nil
}

func newCloudeventClient(config env.NatsConfig) (cev2.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	return cev2.NewClientHTTP(cev2.WithRoundTripper(transport))
}

// SyncSubscription synchronizes the given Kyma subscription to NATS subscription.
func (n *Nats) SyncSubscription(sub *eventingv1alpha1.Subscription) error {
	// Format logger
	log := utils.LoggerWithSubscription(n.namedLogger(), sub)
	subKeyPrefix := backendnats.CreateKeyPrefix(sub)

	err := n.cleanupNATSSubscriptions(sub, log)
	if err != nil {
		return err
	}

	// add/update sink info in map for callbacks
	if sinkURL, ok := n.sinks.Load(subKeyPrefix); !ok || sinkURL != sub.Spec.Sink {
		n.sinks.Store(subKeyPrefix, sub.Spec.Sink)
	}

	for _, subject := range sub.Status.CleanEventTypes {
		callback := n.getCallback(subKeyPrefix, sub.Name)

		if n.connection.Status() != nats.CONNECTED {
			if err := n.Initialize(n.connClosedHandler); err != nil {
				log.Errorw("Failed to reset NATS connection", "status", n.connection.Stats(), "error", err)
				return err
			}
		}

		err = n.createSubscriberPerMaxInFlight(sub, subject, log, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Nats) createSubscriberPerMaxInFlight(sub *eventingv1alpha1.Subscription, subject string, log *zap.SugaredLogger, callback nats.MsgHandler) error {
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		// queueGroupName must be unique for each subscription and subject
		queueGroupName := backendnats.CreateKeyPrefix(sub) + string(types.Separator) + subject
		natsSubKey := backendnats.CreateKey(sub, subject, i)

		// check if the subscription already exists and if it is valid.
		if existingNatsSub, ok := n.subscriptions[natsSubKey]; ok {
			if existingNatsSub.Subject != subject {
				if err := n.deleteSubscriptionFromNATS(existingNatsSub, natsSubKey, log); err != nil {
					return err
				}
			} else if existingNatsSub.IsValid() {
				log.Debugw("Skipping creating subscription on NATS because it already exists", "subject", subject)
				continue
			}
		}

		// otherwise, create subscription on NATS
		natsSub, err := n.connection.QueueSubscribe(subject, queueGroupName, callback)
		if err != nil {
			log.Errorw("Failed to create NATS subscription", "error", err)
			return err
		}

		// save created NATS subscription in storage
		n.subscriptions[natsSubKey] = natsSub
	}
	return nil
}

func (n *Nats) cleanupNATSSubscriptions(sub *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	// check if there is any existing NATS subscription in global list
	// which is not anymore in this subscription filters (i.e. cleanSubjects).
	// e.g. when filters are modified.
	for key, s := range n.subscriptions {
		if backendnats.IsNatsSubAssociatedWithKymaSub(key, s, sub) && !utils.ContainsString(sub.Status.CleanEventTypes, s.Subject) {
			if err := n.deleteSubscriptionFromNATS(s, key, log); err != nil {
				return err
			}
			log.Infow(
				"Deleted NATS subscription because it was deleted from subscription filters",
				"subscriptionKey", key,
				"natsSubject", s.Subject,
			)
		}
	}
	return nil
}

// DeleteSubscription deletes all NATS subscriptions corresponding to a Kyma subscription
func (n *Nats) DeleteSubscription(sub *eventingv1alpha1.Subscription) error {
	subKeyPrefix := backendnats.CreateKeyPrefix(sub)
	for key, s := range n.subscriptions {
		// format logger
		log := n.namedLogger().With(
			"kind", sub.GetObjectKind().GroupVersionKind().Kind,
			"name", sub.GetName(),
			"namespace", sub.GetNamespace(),
			"version", sub.GetGeneration(),
			"key", key,
			"subject", s.Subject,
		)

		if backendnats.IsNatsSubAssociatedWithKymaSub(key, s, sub) {
			if err := n.deleteSubscriptionFromNATS(s, key, log); err != nil {
				return err
			}
			// delete subscription sink info from storage
			n.sinks.Delete(subKeyPrefix)
		}
	}
	return nil
}

// GetInvalidSubscriptions returns the NamespacedName of Kyma subscriptions corresponding to NATS subscriptions marked as "invalid" by NATS client.
func (n *Nats) GetInvalidSubscriptions() *[]types.NamespacedName {
	var nsn []types.NamespacedName
	for k, v := range n.subscriptions {
		if !v.IsValid() {
			n.namedLogger().Debugw("Invalid NATS subscription", "key", k, "subject", v.Subject)
			nsn = append(nsn, backendnats.CreateKymaSubscriptionNamespacedName(k, v))
		}
	}
	return &nsn
}

// GetAllSubscriptions returns the map which contains all details of subscription
// Use this only for testing purposes
func (n *Nats) GetAllSubscriptions() map[string]*nats.Subscription {
	return n.subscriptions
}

// deleteSubFromNats deletes subscription from NATS and from in-memory db
func (n *Nats) deleteSubscriptionFromNATS(natsSub *nats.Subscription, subKey string, log *zap.SugaredLogger) error {
	// Unsubscribe call to NATS is async hence checking the status of the connection is important
	if n.connection.Status() != nats.CONNECTED {
		if err := n.Initialize(n.connClosedHandler); err != nil {
			log.Errorw("Failed to connect to NATS", "status", n.connection.Status(), "error", err)
			return errors.Wrapf(err, "connect to NATS failed")
		}
	}
	if natsSub.IsValid() {
		if err := natsSub.Unsubscribe(); err != nil {
			log.Errorw("Failed to unsubscribe", "error", err)
			return errors.Wrapf(err, "unsubscribe failed")
		}
	}
	delete(n.subscriptions, subKey)
	log.Debugw("Unsubscribed from NATS", "subscriptionKey", subKey)

	return nil
}

func (n *Nats) getCallback(subKeyPrefix, subscriptionName string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		// fetch sink info from storage
		sinkValue, ok := n.sinks.Load(subKeyPrefix)
		if !ok {
			n.namedLogger().Errorw("Failed to find sink URL in storage", "keyPrefix", subKeyPrefix)
			return
		}
		// convert interface type to string
		sink, ok := sinkValue.(string)
		if !ok {
			n.namedLogger().Errorw("Failed to convert sink value to string", "sinkValue", sinkValue)
			return
		}

		ce, err := backendnats.ConvertMsgToCE(msg)
		if err != nil {
			n.namedLogger().Errorw("Failed to convert NATS message to CloudEvent", "error", err)
			return
		}

		// Creating a context with cancellable
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Creating a context with target
		ctxWithCE := cev2.ContextWithTarget(ctxWithCancel, sink)

		// Add tracing headers to the subsequent request
		traceCtxWithCE := tracing.AddTracingHeadersToContext(ctxWithCE, ce)
		// set retries parameters
		retryParams := cev2context.RetryParams{
			Strategy: backoffStrategy,
			MaxTries: n.defaultSubsConfig.DispatcherMaxRetries,
			Period:   n.defaultSubsConfig.DispatcherRetryPeriod,
		}

		ceLogger := n.namedLogger().With("id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
		if result := n.doWithRetry(traceCtxWithCE, retryParams, ce); !cev2.IsACK(result) {
			n.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusInternalServerError)
			ceLogger.Errorw("Faied to dispatch CloudEvent failed after retries", "error", result)
			return
		}
		n.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusOK)
		ceLogger.Infow("CloudEvent was dispatched")
	}
}

func (n *Nats) doWithRetry(ctx context.Context, params cev2context.RetryParams, ce *cev2event.Event) cev2protocol.Result {
	retry := 0
	for {
		result := n.client.Send(ctx, *ce)
		if cev2protocol.IsACK(result) {
			return result
		}
		n.namedLogger().Errorw("Failed to dispatch CloudEvent", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "error", result, "retry", retry)
		// Try again?
		if err := params.Backoff(ctx, retry+1); err != nil {
			// do not try again.
			n.namedLogger().Errorw("Backoff error, will not try again", "error", err)
			return result
		}
		retry++
	}
}

func (n *Nats) namedLogger() *zap.SugaredLogger {
	return n.logger.WithContext().Named(natsHandlerName)
}
