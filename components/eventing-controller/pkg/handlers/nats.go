package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ceclient "github.com/cloudevents/sdk-go/v2/client"

	cev2event "github.com/cloudevents/sdk-go/v2/event"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

// compile time check
var _ NatsInterface = &Nats{}

type NatsInterface interface {
	Initialize() error
	SyncSubscription(subscription *eventingv1alpha1.Subscription) error
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

type Nats struct {
	Connection    *nats.Conn
	Subscriptions map[string]*nats.Subscription
	Log           logr.Logger
	Config        env.NatsConfig
}

const (
	period   = time.Minute
	maxTries = 5
)

// Initialize creates a connection to Nats
func (n *Nats) Initialize() error {
	n.Log.Info("Initialize NATS connection")
	var err error
	if n.Connection == nil || n.Connection.Status() != nats.CONNECTED {
		n.Connection, err = nats.Connect(n.Config.Url,
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(n.Config.MaxReconnects),
			nats.ReconnectWait(n.Config.ReconnectWait))
		if err != nil {
			return errors.Wrapf(err, "failed to connect to Nats")
		}
		if n.Connection.Status() != nats.CONNECTED {
			notConnectedErr := fmt.Errorf("not connected: status: %v", n.Connection.Status())
			return notConnectedErr
		}
	}
	n.Log.Info(fmt.Sprintf("Successfully connected to Nats: %v", n.Connection.Status()))
	return nil
}

func (n *Nats) SyncSubscription(sub *eventingv1alpha1.Subscription) error {
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
		eventType := filter.EventType.Value
		callback := n.getCallback(sub)

		if n.Connection.Status() != nats.CONNECTED {
			n.Log.Info("connection to Nats", "status",
				fmt.Sprintf("%v", n.Connection.Status()))
			initializeErr := n.Initialize()
			if initializeErr != nil {
				n.Log.Error(initializeErr, "failed to reset NATS connection")
				return initializeErr
			}
		}

		sub, subscribeErr := n.Connection.Subscribe(eventType, callback)
		if subscribeErr != nil {
			n.Log.Error(subscribeErr, "failed to create a Nats subscription")
			return subscribeErr
		}
		n.Subscriptions[fmt.Sprintf("%s.%s", namespacedName.String(), filter.EventType.Value)] = sub
	}
	return nil
}

func (n *Nats) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	subNsName := types.NamespacedName{
		Namespace: subscription.Namespace,
		Name:      subscription.Name,
	}
	for k, v := range n.Subscriptions {
		if strings.HasPrefix(k, subNsName.String()) {
			n.Log.Info("connection status", "status", n.Connection.Status())
			// Unsubscribe call to Nats is async hence checking the status of the connection is important
			if n.Connection.Status() != nats.CONNECTED {
				initializeErr := n.Initialize()
				if initializeErr != nil {
					return errors.Wrapf(initializeErr, "can't connect to Nats server")
				}
			}
			err := v.Unsubscribe()
			if err != nil {
				return errors.Wrapf(err, "failed to unsubscribe")
			}
			delete(n.Subscriptions, k)
			n.Log.Info("successfully unsubscribed/deleted subscription")
		}
	}
	return nil
}

func (n Nats) getCallback(sub *eventingv1alpha1.Subscription) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		ce, err := convertMsgToCE(msg)
		if err != nil {
			n.Log.Error(err, "failed to convert Nats message to CE")
			n.NAcknowledge(msg)
			return
		}
		ctx := context.Background()
		// Creating a context with cancellable
		ctxWithCancel, cancel := context.WithCancel(ctx)
		// Creating a context with retries
		ctxWithRetries := cloudevents.ContextWithRetriesExponentialBackoff(ctxWithCancel, period, maxTries)
		// Creating a context with target
		ctxWithCE := cloudevents.ContextWithTarget(ctxWithRetries, sub.Spec.Sink)
		// Cancel later to clean up
		defer cancel()
		cev2Client, err := ceclient.NewDefault()
		if err != nil {
			n.Log.Error(err, "failed to create CE client")
			n.NAcknowledge(msg)
			return
		}
		err = cev2Client.Send(ctxWithCE, *ce)
		if err != nil {
			if !strings.Contains(err.Error(), "200") {
				n.Log.Error(err, "failed to dispatch event")
				n.NAcknowledge(msg)
				return
			}
		}
		// Msgs are Acked automatically on return from the callback. Ref: https://docs.nats.io/developing-with-nats-streaming/acks
		n.Log.Info(fmt.Sprintf("Successfully dispatched event id: %s to sink: %s", ce.ID(),
			sub.Spec.Sink))
	}
}

func (n Nats) NAcknowledge(msg *nats.Msg) {
	err := msg.Nak()
	if err != nil {
		n.Log.Error(err, "failed to NAK event to Nats")
	}
}

func convertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	//Trimming leading and trailing quote from data added by json.Unmarshal
	event.DataEncoded = event.Data()[1 : len(event.Data())-1]
	if err := event.Validate(); err != nil {
		return nil, err
	}
	return &event, nil
}
