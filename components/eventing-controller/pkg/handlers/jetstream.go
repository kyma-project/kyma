package handlers

import (
	"context"
	"fmt"
	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"net/http"
	"sync"
	"time"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var _ JetStreamBackend = &JetStream{}

const (
	jsHandlerName = "js-handler"
)

type JetStreamBackend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(connCloseHandler ConnClosedHandler) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of Jetstream.
	SyncSubscription(subscription *eventingv1alpha1.Subscription) error

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

type JetStream struct {
	config        env.NatsConfig
	conn          *nats.Conn
	jsCtx         nats.JetStreamContext
	client        cev2.Client
	subscriptions map[string]*nats.Subscription
	sinks         sync.Map
	// connClosedHandler gets called by the NATS server when conn is closed and retry attempts are exhausted.
	connClosedHandler ConnClosedHandler
	logger            *logger.Logger
}

func NewJetStream(config env.NatsConfig, logger *logger.Logger) *JetStream {
	return &JetStream{
		config:        config,
		logger:        logger,
		subscriptions: make(map[string]*nats.Subscription),
	}
}

func (js *JetStream) initCloudEventClient(config env.NatsConfig) error {
	if js.client != nil {
		return nil
	}
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	client, err := cev2.NewClientHTTP(cev2.WithRoundTripper(transport))
	if err != nil {
		return err
	}
	js.client = client
	return nil
}

func (js *JetStream) Initialize(connCloseHandler ConnClosedHandler) error {
	if err := js.validateConfig(); err != nil {
		return err
	}
	if err := js.initNATSConn(); err != nil {
		return err
	}
	if err := js.initJSContext(); err != nil {
		return err
	}
	if err := js.initCloudEventClient(js.config); err != nil {
		return err
	}
	js.initJSConnCloseHandler(connCloseHandler)
	return js.ensureStreamExists()
}

func (js *JetStream) SyncSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := utils.LoggerWithSubscription(js.namedLogger(), subscription)
	subKeyPrefix := createKeyPrefix(subscription)

	// add/update sink info in map for callbacks
	if sinkURL, ok := js.sinks.Load(subKeyPrefix); !ok || sinkURL != subscription.Spec.Sink {
		js.sinks.Store(subKeyPrefix, subscription.Spec.Sink)
	}

	callback := js.getCallback(subKeyPrefix)

	for _, subject := range subscription.Status.CleanEventTypes {
		consumerHash := js.generateConsumerHash(subject, subscription)
		log.Infow("Unique Consumer and Subject", "consumerHash", consumerHash, "subject", subject)

		if js.conn.Status() != nats.CONNECTED {
			if err := js.Initialize(js.connClosedHandler); err != nil {
				log.Errorw("reset NATS connection failed", "status", js.conn.Stats(), "error", err)
				return err
			}
		}

		// check if the subscription already exists and if it is valid.
		if existingNatsSub, ok := js.subscriptions[consumerHash]; ok {
			// TODO: Compare if the subjects are the same
			if existingNatsSub.IsValid() {
				log.Debugw("skipping creating subscription on JetStream because it already exists", "subject", subject)
				continue
			}
		}

		jsSubscription, err := js.jsCtx.Subscribe(
			fmt.Sprintf("%s", subject),
			callback,
			js.getDefaultSubscriptionOptions(consumerHash)...,
		)
		if err != nil {
			log.Errorw("Subscription error", "err", err)
			return err
		}
		// save created JetStream subscription in storage
		js.subscriptions[consumerHash] = jsSubscription
	}
	return nil
}

type DefaultSubOpts []nats.SubOpt

func (js *JetStream) getDefaultSubscriptionOptions(consumer string) DefaultSubOpts {
	defaultOpts := DefaultSubOpts{
		nats.Durable(consumer),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.IdleHeartbeat(30 * time.Second),
		nats.EnableFlowControl(),
		nats.MaxAckPending(250),
		nats.MaxDeliver(3),
	}
	return defaultOpts
}

func (js *JetStream) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := utils.LoggerWithSubscription(js.namedLogger(), subscription)

	// loop over the global list of subscriptions
	// and delete any related JetStream subscription
	for key, jsSub := range js.subscriptions {
		if js.isJsSubAssociatedWithKymaSub(key, subscription) {
			if err := js.deleteSubscriptionFromJS(jsSub, key, log); err != nil {
				return err
			}
		}
	}

	// cleanup consumers on nats-server
	// in-case data in js.subscriptions[] was lost due to handler restart
	for _, subject := range subscription.Status.CleanEventTypes {
		consumerHash := js.generateConsumerHash(subject, subscription)
		if err := js.deleteConsumerFromJS(consumerHash, log); err != nil {
			return err
		}
	}

	// delete subscription sink info from storage
	js.sinks.Delete(createKeyPrefix(subscription))

	return nil
}

func (js *JetStream) validateConfig() error {
	if js.config.JSStreamName == "" {
		return errors.New("Stream name cannot be empty")
	}
	if _, err := toJetStreamStorageType(js.config.JSStreamStorageType); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) initNATSConn() error {
	if js.conn == nil || js.conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.config.MaxReconnects),
			nats.ReconnectWait(js.config.ReconnectWait),
		}
		conn, err := nats.Connect(js.config.URL, jsOptions...)
		if err != nil || !conn.IsConnected() {
			return errors.Wrapf(err, "failed to connect to NATS JetStream")
		}
		js.conn = conn
		if js.connClosedHandler != nil {
			js.conn.SetClosedHandler(nats.ConnHandler(js.connClosedHandler))
		}
	}
	return nil
}

func (js *JetStream) initJSContext() error {
	jsCtx, err := js.conn.JetStream()
	if err != nil {
		return errors.Wrapf(err, "failed to create the JetStream context")
	}
	js.jsCtx = jsCtx
	return nil
}

func (js *JetStream) initJSConnCloseHandler(connCloseHandler ConnClosedHandler) {
	js.connClosedHandler = connCloseHandler
	if js.connClosedHandler != nil {
		js.conn.SetClosedHandler(nats.ConnHandler(js.connClosedHandler))
	}
}

func (js *JetStream) ensureStreamExists() error {
	if info, err := js.jsCtx.StreamInfo(js.config.JSStreamName); err == nil {
		// TODO: in case the stream exists, should we make sure all of its configs matches the stream config in the chart?
		// TODO: Handle if info is nil
		js.namedLogger().Infow("reusing existing Stream", "stream-info", info)
		return nil
	} else if err != nats.ErrStreamNotFound {
		return err
	}
	storage, err := toJetStreamStorageType(js.config.JSStreamStorageType)
	if err != nil {
		return err
	}
	js.namedLogger().Infow("Stream not found, creating a new Stream",
		"streamName", js.config.JSStreamName, "streamStorageType", storage.String())
	_, err = js.jsCtx.AddStream(&nats.StreamConfig{
		Name:    js.config.JSStreamName,
		Storage: storage,
		// Since one stream is used to store events of all types, the stream has to match all event types, and therefore
		// we use the wildcard char >. However, to avoid matching internal JetStream and non-Kyma-related subjects, we
		// use the stream name as a prefix. This prefix is handled only on the JetStream level (i.e. JetStream handler
		// and EPP) and should not be exposed in the Kyma subscription. Any Kyma event type gets appended with the
		// configured stream name.
		// TODO: Discuss on the prefix
		Subjects: []string{fmt.Sprintf("%s.>", "sap")},
	})
	return err
}

func (js *JetStream) getCallback(subKeyPrefix string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		// fetch sink info from storage
		sinkValue, ok := js.sinks.Load(subKeyPrefix)
		if !ok {
			js.namedLogger().Errorw("cannot find sink url in storage", "keyPrefix", subKeyPrefix)
			return
		}
		// convert interface type to string
		sink, ok := sinkValue.(string)
		if !ok {
			js.namedLogger().Errorw("failed to convert sink value to string", "sinkValue", sinkValue)
			return
		}

		ce, err := convertMsgToCE(msg)
		if err != nil {
			js.namedLogger().Errorw("convert Jetstream message to CE failed", "error", err)
			return
		}

		ctxWithCancel, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctxWithCE := cev2.ContextWithTarget(ctxWithCancel, sink)
		traceCtxWithCE := tracing.AddTracingHeadersToContext(ctxWithCE, ce)

		js.namedLogger().Infow("Sending the cloud event", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
		result := js.client.Send(traceCtxWithCE, *ce)
		if !cev2protocol.IsACK(result) {
			if err := msg.Nak(); err != nil {
				js.namedLogger().Errorw("Event dispatch failed and also failed to NAK")
			}
			js.namedLogger().Errorw("Event dispatch failed")
			return
		}
		if err := msg.Ack(); err != nil {
			js.namedLogger().Errorw("Message Acknowledgement failed")
		}
		js.namedLogger().Infow("event dispatched", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
	}
}

// isNatsSubAssociatedWithKymaSub checks if the JetStream subscription is associated / related to Kyma subscription or not.
func (js *JetStream) isJsSubAssociatedWithKymaSub(jsSubKey string, subscription *eventingv1alpha1.Subscription) bool {
	for _, subject := range subscription.Status.CleanEventTypes {
		if js.generateConsumerHash(subject, subscription) == jsSubKey {
			return true
		}
	}
	return false
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}

// deleteSubscriptionFromJS deletes subscription from JetStream and from in-memory db
func (js *JetStream) deleteSubscriptionFromJS(jsSub *nats.Subscription, subKey string, log *zap.SugaredLogger) error {
	// Unsubscribe call to NATS is async hence checking the status of the connection is important
	if js.conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			log.Errorw("connect to JetStream failed", "status", js.conn.Status(), "error", err)
			return errors.Wrapf(err, "connect to JetStream failed")
		}
	}

	if jsSub.IsValid() {
		if err := jsSub.Unsubscribe(); err != nil {
			log.Errorw("unsubscribe from JetStream failed", "error", err, "jsSub", jsSub)
			return errors.Wrapf(err, "unsubscribe failed")
		}
	}

	delete(js.subscriptions, subKey)
	log.Debugw("unsubscribe from JetStream succeeded", "subscriptionKey", subKey)

	return nil
}

// deleteConsumerFromJS deletes consumer from JetStream
func (js *JetStream) deleteConsumerFromJS(name string, log *zap.SugaredLogger) error {
	// checking the status of the connection is important
	if js.conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			log.Errorw("connect to JetStream failed", "status", js.conn.Status(), "error", err)
			return errors.Wrapf(err, "connect to JetStream failed")
		}
	}

	if err := js.jsCtx.DeleteConsumer(js.config.JSStreamName, name); err != nil && err != nats.ErrConsumerNotFound {
		// if it is not a Not Found error, then return error
		log.Errorw("failed to delete consumer from JetStream", "error", err, "consumer", name)
		return err
	}

	return nil
}

// generateConsumerHash generates a hash for consumer
func (js *JetStream) generateConsumerHash(subject string, subscription *eventingv1alpha1.Subscription) string {
	return generateHashForString(fmt.Sprintf("%s/%s", createKeyPrefix(subscription), subject))
}
