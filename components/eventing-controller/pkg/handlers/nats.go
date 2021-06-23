package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"

	"k8s.io/apimachinery/pkg/types"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
)

// compile time check
var _ MessagingBackend = &Nats{}

type Nats struct {
	config        env.NatsConfig
	log           logr.Logger
	client        cev2.Client
	connection    *nats.Conn
	subscriptions map[string]*nats.Subscription
}

func NewNats(config env.NatsConfig, log logr.Logger) *Nats {
	return &Nats{config: config, log: log, subscriptions: make(map[string]*nats.Subscription)}
}

const (
	period         = time.Minute
	maxTries       = 5
	traceParentKey = "traceparent"
)

// Initialize creates a connection to Nats
func (n *Nats) Initialize(cfg env.Config) error {
	n.log.Info("Initialize NATS connection")
	var err error
	if n.connection == nil || n.connection.Status() != nats.CONNECTED {
		n.connection, err = nats.Connect(n.config.Url,
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(n.config.MaxReconnects),
			nats.ReconnectWait(n.config.ReconnectWait))
		if err != nil {
			return errors.Wrapf(err, "failed to connect to Nats")
		}
		if n.connection.Status() != nats.CONNECTED {
			notConnectedErr := fmt.Errorf("not connected: status: %v", n.connection.Status())
			return notConnectedErr
		}
	}
	n.log.Info("Successfully connected to Nats", "status", n.connection.Status())

	if n.client != nil {
		return nil
	}

	n.log.Info("Initialize cloudevents client")
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

// The returned bool should be ignored now. It's a marker for changed subscription status
func (n *Nats) SyncSubscription(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, params ...interface{}) (bool, error) {
	var filters []*eventingv1alpha1.BebFilter
	if sub.Spec.Filter != nil {
		uniqueFilters, err := sub.Spec.Filter.Deduplicate()
		if err != nil {
			return false, errors.Wrap(err, "error deduplicating subscription filters")
		}
		filters = uniqueFilters.Filters
	}
	// Create subscriptions in Nats
	for _, filter := range filters {
		subject, err := createSubject(filter, cleaner)
		if err != nil {
			n.log.Error(err, "failed to create a Nats subject")
			return false, err
		}

		callback := n.getCallback(sub.Spec.Sink)

		if n.connection.Status() != nats.CONNECTED {
			n.log.Info("connection to Nats", "status", fmt.Sprintf("%v", n.connection.Status()))
			initializeErr := n.Initialize(env.Config{})
			if initializeErr != nil {
				n.log.Error(initializeErr, "failed to reset NATS connection")
				return false, initializeErr
			}
		}

		natsSub, subscribeErr := n.connection.Subscribe(subject, callback)
		if subscribeErr != nil {
			n.log.Error(subscribeErr, "failed to create a Nats subscription")
			return false, subscribeErr
		}
		n.subscriptions[createKey(sub, subject)] = natsSub
	}
	return false, nil
}

// DeleteSubscription deletes all NATS subscriptions corresponding to a Kyma subscription
func (n *Nats) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	for key, sub := range n.subscriptions {
		if strings.HasPrefix(key, createKeyPrefix(subscription)) {
			n.log.Info("connection status", "status", n.connection.Status())
			// Unsubscribe call to Nats is async hence checking the status of the connection is important
			if n.connection.Status() != nats.CONNECTED {
				initializeErr := n.Initialize(env.Config{})
				if initializeErr != nil {
					return errors.Wrapf(initializeErr, "can't connect to NATS server")
				}
			}
			if sub.IsValid() {
				if err := sub.Unsubscribe(); err != nil {
					return errors.Wrapf(err, "failed to unsubscribe")
				}
			} else {
				n.log.Info("cannot unsubscribe an invalid NATS subscription: ", "key", key, "subject", sub.Subject)
			}
			delete(n.subscriptions, key)
			n.log.Info("successfully unsubscribed/deleted subscription")
		}
	}
	return nil
}

// GetInvalidSubscriptions returns the NamespacedName of Kyma subscriptions corresponding to NATS subscriptions marked as "invalid" by NATS client
func (n *Nats) GetInvalidSubscriptions() *[]types.NamespacedName {
	var nsn []types.NamespacedName
	for k, v := range n.subscriptions {
		if !v.IsValid() {
			n.log.Info("invalid NATS subscription: ", "key", k, "subject", v.Subject)
			nsn = append(nsn, createKymaSubscriptionNamespacedName(k, v))
		}
	}
	return &nsn
}

func (n *Nats) getCallback(sink string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		ce, err := convertMsgToCE(msg)
		if err != nil {
			n.log.Error(err, "failed to convert Nats message to CE")
			return
		}

		// Creating a context with cancellable
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Creating a context with retries
		ctxWithRetries := cev2.ContextWithRetriesExponentialBackoff(ctxWithCancel, period, maxTries)

		// Creating a context with target
		ctxWithCE := cev2.ContextWithTarget(ctxWithRetries, sink)

		// Add tracing headers to the subsequent request
		traceCtxWithCE := tracing.AddTracingHeadersToContext(ctxWithCE, ce)

		if result := n.client.Send(traceCtxWithCE, *ce); !cev2.IsACK(result) {
			n.log.Error(result, "failed to dispatch event")
			return
		}
		n.log.Info(fmt.Sprintf("Successfully dispatched event id: %s to sink: %s", ce.ID(), sink))
	}
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
	return fmt.Sprintf("%s", namespacedName.String())
}

func createKey(sub *eventingv1alpha1.Subscription, subject string) string {
	return fmt.Sprintf("%s.%s", createKeyPrefix(sub), subject)
}

func createSubject(filter *eventingv1alpha1.BebFilter, cleaner eventtype.Cleaner) (string, error) {
	eventType := strings.TrimSpace(filter.EventType.Value)
	if len(eventType) == 0 {
		return "", nats.ErrBadSubject
	}
	// clean the application name segment in the event-type from none-alphanumeric characters
	// return it as a Nats subject
	return cleaner.Clean(eventType)
}

func createKymaSubscriptionNamespacedName(key string, sub *nats.Subscription) types.NamespacedName {
	nsn := types.NamespacedName{}
	nnvalues := strings.Split(strings.TrimSuffix(strings.TrimSuffix(key, sub.Subject), "."), string(types.Separator))
	nsn.Namespace = nnvalues[0]
	nsn.Name = nnvalues[1]
	return nsn
}
