package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	period   = time.Minute
	maxTries = 5
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
	protocol, err := cev2.NewHTTP(cev2.WithRoundTripper(transport))
	if err != nil {
		return nil, err
	}
	return cev2.NewClientObserved(protocol, cev2.WithTimeNow(), cev2.WithUUIDs(), cev2.WithTracePropagation)
}

// The returned bool should be ignored now. It's a marker for changed subscription status
func (n *Nats) SyncSubscription(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, params ...interface{}) (bool, error) {
	namespacedName := types.NamespacedName{
		Namespace: sub.Namespace,
		Name:      sub.Name,
	}
	var filters []*eventingv1alpha1.BebFilter
	if sub.Spec.Filter != nil {
		filters = sub.Spec.Filter.Filters
	}
	// Create subscriptions in Nats
	for _, filter := range filters {
		eventType := strings.TrimSpace(filter.EventType.Value)
		if len(eventType) == 0 {
			err := nats.ErrBadSubject
			n.log.Error(err, "failed to create a Nats subscription")
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

		// clean the application name segment in the event-type from none-alphanumeric characters
		eventType, err := cleaner.Clean(eventType)
		if err != nil {
			return false, err
		}

		sub, subscribeErr := n.connection.Subscribe(eventType, callback)
		if subscribeErr != nil {
			n.log.Error(subscribeErr, "failed to create a Nats subscription")
			return false, subscribeErr
		}
		n.subscriptions[fmt.Sprintf("%s.%s", namespacedName.String(), filter.EventType.Value)] = sub
	}
	return false, nil
}

func (n *Nats) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	subNsName := types.NamespacedName{
		Namespace: subscription.Namespace,
		Name:      subscription.Name,
	}
	for k, v := range n.subscriptions {
		if strings.HasPrefix(k, subNsName.String()) {
			n.log.Info("connection status", "status", n.connection.Status())
			// Unsubscribe call to Nats is async hence checking the status of the connection is important
			if n.connection.Status() != nats.CONNECTED {
				initializeErr := n.Initialize(env.Config{})
				if initializeErr != nil {
					return errors.Wrapf(initializeErr, "can't connect to Nats server")
				}
			}
			err := v.Unsubscribe()
			if err != nil {
				return errors.Wrapf(err, "failed to unsubscribe")
			}
			delete(n.subscriptions, k)
			n.log.Info("successfully unsubscribed/deleted subscription")
		}
	}
	return nil
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

		if result := n.client.Send(ctxWithCE, *ce); !cev2.IsACK(result) {
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
