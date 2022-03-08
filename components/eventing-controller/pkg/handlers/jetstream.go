package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var _ JetStreamBackend = &JetStream{}

const (
	jsHandlerName         = "js-handler"
	idleHeartBeatDuration = 1 * time.Minute
)

type JetStreamBackend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(connCloseHandler ConnClosedHandler) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of Jetstream.
	SyncSubscription(subscription *eventingv1alpha1.Subscription) error

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error

	// GetJetStreamSubjects returns a list of subjects appended with stream name as prefix if needed
	GetJetStreamSubjects(subjects []string) []string
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

	// check if there is any existing JetStream subscription in global list
	// which is not anymore in this subscription filters (i.e. cleanSubjects).
	// e.g. when filters are modified.
	for key, jsSub := range js.subscriptions {
		// TODO: optimize this call of ConsumerInfo
		// as jsSub.ConsumerInfo() will send an REST call to nats-server for each subject
		info, err := jsSub.ConsumerInfo()
		if err != nil {
			if err == nats.ErrConsumerNotFound {
				continue
			}
			return err
		}

		isRelated, err := js.isJsSubAssociatedWithKymaSub(key, subscription)
		if err != nil {
			return err
		}

		if isRelated && !utils.ContainsString(subscription.Status.CleanEventTypes, info.Config.FilterSubject) {
			if err := js.deleteSubscriptionFromJetStream(jsSub, key, log); err != nil {
				return err
			}
			log.Infow(
				"deleted JetStream subscription because it was deleted from subscription filters",
				"subscriptionKey", key,
				"jetStreamSubject", jsSub.Subject,
			)
		}
	}

	// add/update sink info in map for callbacks
	if sinkURL, ok := js.sinks.Load(subKeyPrefix); !ok || sinkURL != subscription.Spec.Sink {
		js.sinks.Store(subKeyPrefix, subscription.Spec.Sink)
	}

	callback := js.getCallback(subKeyPrefix)
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubKey := js.generateJsSubKey(subject, subscription)

		if js.conn.Status() != nats.CONNECTED {
			if err := js.Initialize(js.connClosedHandler); err != nil {
				log.Errorw("reset JetStream connection failed", "status", js.conn.Stats(), "error", err)
				return err
			}
		}

		// check if the subscription already exists and if it is valid.
		if existingNatsSub, ok := js.subscriptions[jsSubKey]; ok {
			if existingNatsSub.IsValid() {
				log.Debugw("skipping creating subscription on JetStream because it already exists", "subject", subject)
				continue
			}
		}

		// async callback for maxInflight messages
		asyncCallback := func(m *nats.Msg) {
			go callback(m)
		}

		// subscribe to the subject on JetStream
		jsSubscription, err := js.jsCtx.Subscribe(
			subject,
			asyncCallback,
			js.getDefaultSubscriptionOptions(jsSubKey, subscription.Status.Config)...,
		)
		if err != nil {
			log.Errorw("failed to subscribe on JetStream", "subject", subject, "error", err)
			return err
		}
		// save created JetStream subscription in storage
		js.subscriptions[jsSubKey] = jsSubscription
		log.Debugw("created subscription on JetStream", "subject", subject)
	}
	return nil
}

func (js *JetStream) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := utils.LoggerWithSubscription(js.namedLogger(), subscription)

	// loop over the global list of subscriptions
	// and delete any related JetStream subscription
	for key, jsSub := range js.subscriptions {
		if isRelated, err := js.isJsSubAssociatedWithKymaSub(key, subscription); err != nil {
			return err
		} else if isRelated {
			if err = js.deleteSubscriptionFromJetStream(jsSub, key, log); err != nil {
				return err
			}
		}
	}

	// cleanup consumers on nats-server
	// in-case data in js.subscriptions[] was lost due to handler restart
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubKey := js.generateJsSubKey(subject, subscription)
		if err := js.deleteConsumerFromJetStream(jsSubKey, log); err != nil {
			return err
		}
	}

	// delete subscription sink info from storage
	js.sinks.Delete(createKeyPrefix(subscription))

	return nil
}

// GetJsSubjectToSubscribe appends stream name to subject if needed
func (js *JetStream) GetJsSubjectToSubscribe(subject string) string {
	if strings.HasPrefix(subject, js.config.JSStreamName) {
		return subject
	}
	return fmt.Sprintf("%s.%s", js.config.JSStreamName, subject)
}

func (js *JetStream) validateConfig() error {
	if js.config.JSStreamName == "" {
		return errors.New("Stream name cannot be empty")
	}
	if _, err := toJetStreamStorageType(js.config.JSStreamStorageType); err != nil {
		return err
	}
	if _, err := toJetStreamRetentionPolicy(js.config.JSStreamRetentionPolicy); err != nil {
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
		js.namedLogger().Infow("reusing existing Stream", "stream-info", info)
		return nil
	} else if err != nats.ErrStreamNotFound {
		return err
	}
	streamConfig, err := getStreamConfig(js.config)
	if err != nil {
		return err
	}
	js.namedLogger().Infow("Stream not found, creating a new Stream",
		"streamName", js.config.JSStreamName, "streamStorageType", streamConfig.Storage)
	_, err = js.jsCtx.AddStream(streamConfig)
	return err
}

func getStreamConfig(natsConfig env.NatsConfig) (*nats.StreamConfig, error) {
	storage, err := toJetStreamStorageType(natsConfig.JSStreamStorageType)
	if err != nil {
		return nil, err
	}
	retentionPolicy, err := toJetStreamRetentionPolicy(natsConfig.JSStreamRetentionPolicy)
	if err != nil {
		return nil, err
	}
	streamConfig := &nats.StreamConfig{
		Name:      natsConfig.JSStreamName,
		Storage:   storage,
		Retention: retentionPolicy,
		MaxMsgs:   natsConfig.JSStreamMaxMessages,
		MaxBytes:  natsConfig.JSStreamMaxBytes,
		// Since one stream is used to store events of all types, the stream has to match all event types, and therefore
		// we use the wildcard char >. However, to avoid matching internal JetStream and non-Kyma-related subjects, we
		// use the stream name as a prefix. This prefix is handled only on the JetStream level (i.e. JetStream handler
		// and EPP) and should not be exposed in the Kyma subscription. Any Kyma event type gets appended with the
		// configured stream name.
		Subjects: []string{fmt.Sprintf("%s.>", natsConfig.JSStreamName)},
	}
	return streamConfig, nil
}

type DefaultSubOpts []nats.SubOpt

func (js *JetStream) getDefaultSubscriptionOptions(consumerName string, subConfig *eventingv1alpha1.SubscriptionConfig) DefaultSubOpts {
	defaultOpts := DefaultSubOpts{
		nats.Durable(consumerName),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.IdleHeartbeat(idleHeartBeatDuration),
		nats.EnableFlowControl(),
		nats.DeliverNew(),
		nats.MaxAckPending(subConfig.MaxInFlightMessages),
	}
	return defaultOpts
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

		// setup context for dispatching
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctxWithCE := cev2.ContextWithTarget(ctxWithCancel, sink)
		traceCtxWithCE := tracing.AddTracingHeadersToContext(ctxWithCE, ce)

		js.namedLogger().Debugw("sending the cloud event", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)

		// dispatch the event to sink
		result := js.client.Send(traceCtxWithCE, *ce)
		if !cev2protocol.IsACK(result) {
			if err := msg.Nak(); err != nil {
				js.namedLogger().Errorw("failed to NAK an event on JetStream")
			}
			js.namedLogger().Errorw("failed to dispatch an event", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
			return
		}
		if err := msg.Ack(); err != nil {
			js.namedLogger().Errorw("failed to ACK an event on JetStream", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
		}
		js.namedLogger().Infow("event dispatched", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
	}
}

// isJsSubAssociatedWithKymaSub checks if the JetStream subscription is associated / related to Kyma subscription or not.
func (js *JetStream) isJsSubAssociatedWithKymaSub(jsSubKey string, subscription *eventingv1alpha1.Subscription) (bool, error) {
	// extract out namespacedName of subscription from key
	namespacedName, err := createJSSubscriptionNamespacedName(jsSubKey)
	if err != nil {
		return false, err
	}
	// check if the namespacedName matches the target subscription
	return createKeyPrefix(subscription) == namespacedName.String(), nil
}

// deleteSubscriptionFromJS deletes subscription from JetStream and from in-memory db
func (js *JetStream) deleteSubscriptionFromJetStream(jsSub *nats.Subscription, jsSubKey string, log *zap.SugaredLogger) error {
	// unsubscribe call to JetStream is async hence checking the status of the connection is important
	if js.conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			log.Errorw("connect to JetStream failed", "status", js.conn.Status(), "error", err)
			return errors.Wrapf(err, "connect to JetStream failed")
		}
	}

	if jsSub.IsValid() {
		// unsubscribe will also delete the consumer on JS server
		if err := jsSub.Unsubscribe(); err != nil {
			log.Errorw("unsubscribe from JetStream failed", "error", err, "jsSub", jsSub)
			return errors.Wrapf(err, "unsubscribe failed")
		}
	} else {
		// if JS sub is not valid, then we need to delete the consumer on JetStream
		if err := js.deleteConsumerFromJetStream(jsSubKey, log); err != nil {
			return err
		}
	}

	delete(js.subscriptions, jsSubKey)
	log.Debugw("unsubscribe from JetStream succeeded", "subscriptionKey", jsSubKey)

	return nil
}

// deleteConsumerFromJS deletes consumer from JetStream
func (js *JetStream) deleteConsumerFromJetStream(name string, log *zap.SugaredLogger) error {
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

// generateJsSubKey generates an encoded unique key for JetStream subscription
func (js *JetStream) generateJsSubKey(subject string, subscription *eventingv1alpha1.Subscription) string {
	return encodeString(fmt.Sprintf("%s/%s", createKeyPrefix(subscription), subject))
}

// GetJetStreamSubjects returns a list of subjects appended with prefix if needed
func (js *JetStream) GetJetStreamSubjects(subjects []string) []string {
	var result []string
	for _, subject := range subjects {
		result = append(result, js.GetJsSubjectToSubscribe(subject))
	}
	return result
}

func createJSSubscriptionNamespacedName(jsSubKey string) (types.NamespacedName, error) {
	consumer, err := decodeString(jsSubKey)
	if err != nil {
		return types.NamespacedName{}, err
	}
	nsn := types.NamespacedName{}
	nnValues := strings.Split(consumer, string(types.Separator))
	nsn.Namespace = nnValues[0]
	nsn.Name = nnValues[1]
	return nsn, nil
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
