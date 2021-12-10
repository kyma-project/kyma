package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
var _ MessagingBackend = &Nats{}

type Nats struct {
	config            env.NatsConfig
	defaultSubsConfig env.DefaultSubscriptionConfig
	logger            *logger.Logger
	client            cev2.Client
	connection        *nats.Conn
	subscriptions     map[string]*nats.Subscription
}

func NewNats(config env.NatsConfig, subsConfig env.DefaultSubscriptionConfig, logger *logger.Logger) *Nats {
	return &Nats{
		config:            config,
		logger:            logger,
		subscriptions:     make(map[string]*nats.Subscription),
		defaultSubsConfig: subsConfig,
	}
}

const (
	backoffStrategy = cev2context.BackoffStrategyConstant
	natsHandlerName = "nats-handler"
)

// Initialize creates a connection to NATS.
func (n *Nats) Initialize(env.Config) (err error) {
	if n.connection == nil || n.connection.Status() != nats.CONNECTED {
		n.connection, err = nats.Connect(n.config.URL, nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(n.config.MaxReconnects), nats.ReconnectWait(n.config.ReconnectWait))
		if err != nil {
			return errors.Wrapf(err, "connect to NATS failed")
		}
		if n.connection.Status() != nats.CONNECTED {
			return errors.Errorf("connect to NATS failed status: %v", n.connection.Status())
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

// SyncSubscription synchronizes the given Kyma subscription to NATS subscription.
// note: the returned bool should be ignored now. It should act as a marker for changed subscription status.
func (n *Nats) SyncSubscription(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, _ ...interface{}) (bool, error) {
	var filters []*eventingv1alpha1.BEBFilter
	if sub.Spec.Filter != nil {
		uniqueFilters, err := sub.Spec.Filter.Deduplicate()
		if err != nil {
			return false, errors.Wrap(err, "deduplicate subscription filters failed")
		}
		filters = uniqueFilters.Filters
	}

	// Format logger
	log := utils.LoggerWithSubscription(n.namedLogger(), sub)

	subsConfig := eventingv1alpha1.MergeSubsConfigs(sub.Spec.Config, &n.defaultSubsConfig)
	// Create subscriptions in NATS
	for _, filter := range filters {
		subject, err := createSubject(filter, cleaner)
		if err != nil {
			log.Errorw("create NATS subject failed", "error", err)
			return false, err
		}

		callback := n.getCallback(sub.Spec.Sink)

		if n.connection.Status() != nats.CONNECTED {
			if err := n.Initialize(env.Config{}); err != nil {
				log.Errorw("reset NATS connection failed", "status", n.connection.Stats(), "error", err)
				return false, err
			}
		}

		for i := 0; i < subsConfig.MaxInFlightMessages; i++ {
			// queueGroupName must be unique for each subscription and subject
			queueGroupName := createKeyPrefix(sub) + string(types.Separator) + subject
			natsSub, subscribeErr := n.connection.QueueSubscribe(subject, queueGroupName, callback)
			if subscribeErr != nil {
				log.Errorw("create NATS subscription failed", "error", err)
				return false, subscribeErr
			}
			n.subscriptions[createKey(sub, subject, i)] = natsSub
		}
	}
	sub.Status.Config = subsConfig

	return false, nil
}

// DeleteSubscription deletes all NATS subscriptions corresponding to a Kyma subscription
func (n *Nats) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	for key, sub := range n.subscriptions {
		// Format logger
		log := n.namedLogger().With(
			"kind", subscription.GetObjectKind().GroupVersionKind().Kind,
			"name", subscription.GetName(),
			"namespace", subscription.GetNamespace(),
			"version", subscription.GetGeneration(),
			"key", key,
			"subject", sub.Subject,
		)

		if strings.HasPrefix(key, createKeyPrefix(subscription)) {
			// Unsubscribe call to NATS is async hence checking the status of the connection is important
			if n.connection.Status() != nats.CONNECTED {
				if err := n.Initialize(env.Config{}); err != nil {
					log.Errorw("connect to NATS failed", "status", n.connection.Status(), "error", err)
					return errors.Wrapf(err, "connect to NATS failed")
				}
			}
			if sub.IsValid() {
				if err := sub.Unsubscribe(); err != nil {
					log.Errorw("unsubscribe failed", "error", err)
					return errors.Wrapf(err, "unsubscribe failed")
				}
			}
			delete(n.subscriptions, key)
			log.Infow("unsubscribe succeeded")
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

func (n *Nats) getCallback(sink string) nats.MsgHandler {
	return func(msg *nats.Msg) {
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

func createSubject(filter *eventingv1alpha1.BEBFilter, cleaner eventtype.Cleaner) (string, error) {
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
