package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"k8s.io/apimachinery/pkg/types"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2context "github.com/cloudevents/sdk-go/v2/context"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
)

// compile time check
var _ NatsBackend = &Nats{}

const (
	backoffStrategy = cev2context.BackoffStrategyConstant
	natsHandlerName = "nats-handler"
)

type ConnClosedHandler func(conn *nats.Conn)

type NatsBackend interface {
	// Initialize connects and initializes the NATS backend.
	// connCloseHandler can be used to register a handler that gets called when connection
	// to the NATS server is closed and retry attempts are exceeded.
	Initialize(connCloseHandler ConnClosedHandler) error

	// SyncSubscription synchronizes the Kyma Subscription on the NATS backend.
	SyncSubscription(subscription *eventingv1alpha1.Subscription) error

	// DeleteSubscription deletes the corresponding subscription on the NATS backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

type Nats struct {
	config            env.NatsConfig
	defaultSubsConfig env.DefaultSubscriptionConfig
	logger            *logger.Logger
	client            cev2.Client
	connection        *nats.Conn
	subscriptions     map[string]*nats.Subscription
	sinks             sync.Map
	connClosedHandler ConnClosedHandler
}

func NewNats(config env.NatsConfig, subsConfig env.DefaultSubscriptionConfig, logger *logger.Logger) *Nats {
	return &Nats{
		config:            config,
		defaultSubsConfig: subsConfig,
		logger:            logger,
		subscriptions:     make(map[string]*nats.Subscription),
	}
}

// Initialize creates a connection to NATS.
func (n *Nats) Initialize(connCloseHandler ConnClosedHandler) (err error) {
	if n.connection == nil || n.connection.Status() != nats.CONNECTED {
		natsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(n.config.MaxReconnects),
			nats.ReconnectWait(n.config.ReconnectWait),
		}
		n.connection, err = nats.Connect(n.config.URL, natsOptions...)
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

	if n.client, err = newCloudeventClient(n.config); err != nil {
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

// GetCleanSubjects returns a list of clean eventTypes from the unique filters in the subscription.
func GetCleanSubjects(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner) ([]string, error) {
	var filters []*eventingv1alpha1.BEBFilter
	if sub.Spec.Filter != nil {
		uniqueFilters, err := sub.Spec.Filter.Deduplicate()
		if err != nil {
			return []string{}, errors.Wrap(err, "deduplicate subscription filters failed")
		}
		filters = uniqueFilters.Filters
	}

	var cleanSubjects []string
	for _, filter := range filters {
		subject, err := getCleanSubject(filter, cleaner)
		if err != nil {
			return []string{}, err
		}
		cleanSubjects = append(cleanSubjects, subject)
	}
	return cleanSubjects, nil
}

// SyncSubscription synchronizes the given Kyma subscription to NATS subscription.
func (n *Nats) SyncSubscription(sub *eventingv1alpha1.Subscription) error {
	// Format logger
	log := utils.LoggerWithSubscription(n.namedLogger(), sub)
	subKeyPrefix := createKeyPrefix(sub)

	// check if there is any existing NATS subscription in global list
	// which is not anymore in this subscription filters (i.e. cleanSubjects).
	// e.g. when filters are modified.
	for key, s := range n.subscriptions {
		if isNatsSubAssociatedWithKymaSub(key, s, sub) && !utils.ContainsString(sub.Status.CleanEventTypes, s.Subject) {
			if err := n.deleteSubscriptionFromNATS(s, key, log); err != nil {
				return err
			}
			log.Infow(
				"deleted NATS subscription because it was deleted from subscription filters",
				"subscriptionKey", key,
				"natsSubject", s.Subject,
			)
		}
	}

	// add/update sink info in map for callbacks
	if sinkURL, ok := n.sinks.Load(subKeyPrefix); !ok || sinkURL != sub.Spec.Sink {
		n.sinks.Store(subKeyPrefix, sub.Spec.Sink)
	}

	for _, subject := range sub.Status.CleanEventTypes {
		callback := n.getCallback(subKeyPrefix)

		if n.connection.Status() != nats.CONNECTED {
			if err := n.Initialize(n.connClosedHandler); err != nil {
				log.Errorw("reset NATS connection failed", "status", n.connection.Stats(), "error", err)
				return err
			}
		}

		for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
			// queueGroupName must be unique for each subscription and subject
			queueGroupName := createKeyPrefix(sub) + string(types.Separator) + subject
			natsSubKey := createKey(sub, subject, i)

			// check if the subscription already exists and if it is valid.
			if existingNatsSub, ok := n.subscriptions[natsSubKey]; ok {
				if existingNatsSub.Subject != subject {
					if err := n.deleteSubscriptionFromNATS(existingNatsSub, natsSubKey, log); err != nil {
						return err
					}
				} else if existingNatsSub.IsValid() {
					log.Debugw("skipping creating subscription on NATS because it already exists", "subject", subject)
					continue
				}
			}

			// otherwise, create subscription on NATS
			natsSub, err := n.connection.QueueSubscribe(subject, queueGroupName, callback)
			if err != nil {
				log.Errorw("create NATS subscription failed", "error", err)
				return err
			}

			// save created NATS subscription in storage
			n.subscriptions[natsSubKey] = natsSub
		}
	}

	return nil
}

// DeleteSubscription deletes all NATS subscriptions corresponding to a Kyma subscription
func (n *Nats) DeleteSubscription(sub *eventingv1alpha1.Subscription) error {
	subKeyPrefix := createKeyPrefix(sub)
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

		if isNatsSubAssociatedWithKymaSub(key, s, sub) {
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
			n.namedLogger().Debugw("invalid NATS subscription", "key", k, "subject", v.Subject)
			nsn = append(nsn, createKymaSubscriptionNamespacedName(k, v))
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
			log.Errorw("connect to NATS failed", "status", n.connection.Status(), "error", err)
			return errors.Wrapf(err, "connect to NATS failed")
		}
	}
	if natsSub.IsValid() {
		if err := natsSub.Unsubscribe(); err != nil {
			log.Errorw("unsubscribe failed", "error", err)
			return errors.Wrapf(err, "unsubscribe failed")
		}
	}
	delete(n.subscriptions, subKey)
	log.Debugw("unsubscribe from NATS succeeded", "subscriptionKey", subKey)

	return nil
}

func (n *Nats) getCallback(subKeyPrefix string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		// fetch sink info from storage
		sinkValue, ok := n.sinks.Load(subKeyPrefix)
		if !ok {
			n.namedLogger().Errorw("cannot find sink url in storage", "keyPrefix", subKeyPrefix)
			return
		}
		// convert interface type to string
		sink, ok := sinkValue.(string)
		if !ok {
			n.namedLogger().Errorw("failed to convert sink value to string", "sinkValue", sinkValue)
			return
		}

		ce, err := convertMsgToCE(msg)
		if err != nil {
			n.namedLogger().Errorw("convert NATS message to CE failed", "error", err)
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
		if result := n.doWithRetry(traceCtxWithCE, retryParams, ce); !cev2.IsACK(result) {
			n.namedLogger().Errorw("event dispatch failed after retries", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink, "error", result)
			return
		}
		n.namedLogger().Infow("event dispatched", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
	}
}

func (n *Nats) doWithRetry(ctx context.Context, params cev2context.RetryParams, ce *cev2event.Event) cev2protocol.Result {
	retry := 0
	for {
		result := n.client.Send(ctx, *ce)
		if cev2protocol.IsACK(result) {
			return result
		}
		n.namedLogger().Errorw("event dispatch failed", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "error", result, "retry", retry)
		// Try again?
		if err := params.Backoff(ctx, retry+1); err != nil {
			// do not try again.
			n.namedLogger().Errorw("backoff error, will not try again", "error", err)
			return result
		}
		retry++
	}
}

func (n *Nats) namedLogger() *zap.SugaredLogger {
	return n.logger.WithContext().Named(natsHandlerName)
}

func convertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	if err := event.Validate(); err != nil {
		return nil, err
	}
	return &event, nil
}

func createKeyPrefix(sub *eventingv1alpha1.Subscription) string {
	namespacedName := types.NamespacedName{
		Namespace: sub.Namespace,
		Name:      sub.Name,
	}
	return namespacedName.String()
}

func createKeySuffix(subject string, queueGoupInstanceNo int) string {
	return subject + string(types.Separator) + strconv.Itoa(queueGoupInstanceNo)
}

func createKey(sub *eventingv1alpha1.Subscription, subject string, queueGoupInstanceNo int) string {
	return fmt.Sprintf("%s.%s", createKeyPrefix(sub), createKeySuffix(subject, queueGoupInstanceNo))
}

func getCleanSubject(filter *eventingv1alpha1.BEBFilter, cleaner eventtype.Cleaner) (string, error) {
	eventType := strings.TrimSpace(filter.EventType.Value)
	if len(eventType) == 0 {
		return "", nats.ErrBadSubject
	}
	// clean the application name segment in the event-type from none-alphanumeric characters
	// return it as a NATS subject
	return cleaner.Clean(eventType)
}

func createKymaSubscriptionNamespacedName(key string, sub *nats.Subscription) types.NamespacedName {
	nsn := types.NamespacedName{}
	nnvalues := strings.Split(key, string(types.Separator))
	nsn.Namespace = nnvalues[0]
	nsn.Name = strings.TrimSuffix(strings.TrimSuffix(nnvalues[1], sub.Subject), ".")
	return nsn
}

// isNatsSubAssociatedWithKymaSub checks if the NATS subscription is associated / related to Kyma subscription or not.
func isNatsSubAssociatedWithKymaSub(natsSubKey string, natsSub *nats.Subscription, sub *eventingv1alpha1.Subscription) bool {
	return createKeyPrefix(sub) == createKymaSubscriptionNamespacedName(natsSubKey, natsSub).String()
}
