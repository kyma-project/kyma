package handlers

import (
	"context"
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	pkgmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/metrics"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

var _ JetStreamBackend = &JetStream{}

const (
	jsHandlerName          = "js-handler"
	idleHeartBeatDuration  = 1 * time.Minute
	jsConsumerMaxRedeliver = 100
	jsConsumerAcKWait      = 30 * time.Second
	jsMaxStreamNameLength  = 32
	separator              = "/"
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

// SubscriptionSubjectIdentifier is used to uniquely identify a Subscription subject.
// It should be used only with JetStream backend.
type SubscriptionSubjectIdentifier struct {
	consumerName, namespacedSubjectName string
}

// NewSubscriptionSubjectIdentifier returns a new SubscriptionSubjectIdentifier instance.
func NewSubscriptionSubjectIdentifier(subscription *eventingv1alpha1.Subscription, subject string) SubscriptionSubjectIdentifier {
	cn := computeConsumerName(subscription, subject)          // compute the consumer name once
	nn := computeNamespacedSubjectName(subscription, subject) // compute the namespaced name with the subject once
	return SubscriptionSubjectIdentifier{consumerName: cn, namespacedSubjectName: nn}
}

// ConsumerName returns the JetStream consumer name.
func (s SubscriptionSubjectIdentifier) ConsumerName() string {
	return s.consumerName
}

// NamespacedName returns the Kubernetes namespaced name.
func (s SubscriptionSubjectIdentifier) NamespacedName() string {
	return s.namespacedSubjectName[:strings.LastIndex(s.namespacedSubjectName, separator)]
}

// computeConsumerName returns JetStream consumer name of the given subscription and subject.
// It uses the crypto/md5 lib to return a string of 32 characters as recommended by the JetStream
// documentation https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming.
func computeConsumerName(subscription *eventingv1alpha1.Subscription, subject string) string {
	cn := subscription.Namespace + separator + subscription.Name + separator + subject
	h := md5.Sum([]byte(cn)) // #nosec
	return hex.EncodeToString(h[:])
}

// computeNamespacedSubjectName returns Kubernetes namespaced name of the given subscription along with the subject.
func computeNamespacedSubjectName(subscription *eventingv1alpha1.Subscription, subject string) string {
	return subscription.Namespace + separator + subscription.Name + separator + subject
}

type JetStream struct {
	config        env.NatsConfig
	conn          *nats.Conn
	jsCtx         nats.JetStreamContext
	client        cev2.Client
	subscriptions map[SubscriptionSubjectIdentifier]*nats.Subscription
	sinks         sync.Map
	// connClosedHandler gets called by the NATS server when conn is closed and retry attempts are exhausted.
	connClosedHandler ConnClosedHandler
	logger            *logger.Logger
	metricsCollector  *pkgmetrics.Collector
}

func NewJetStream(config env.NatsConfig, metricsCollector *pkgmetrics.Collector, logger *logger.Logger) *JetStream {
	return &JetStream{
		config:           config,
		logger:           logger,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]*nats.Subscription),
		metricsCollector: metricsCollector,
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
	if err := js.initNATSConn(connCloseHandler); err != nil {
		return err
	}
	if err := js.initJSContext(); err != nil {
		return err
	}
	if err := js.initCloudEventClient(js.config); err != nil {
		return err
	}
	return js.ensureStreamExists()
}

func (js *JetStream) SyncSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := utils.LoggerWithSubscription(js.namedLogger(), subscription)
	subKeyPrefix := createKeyPrefix(subscription)
	if err := js.checkJetStreamConnection(log); err != nil {
		return err
	}

	if err := js.syncSubscriptionFilters(subscription, log); err != nil {
		return err
	}

	// add/update sink info in map for callbacks
	if sinkURL, ok := js.sinks.Load(subKeyPrefix); !ok || sinkURL != subscription.Spec.Sink {
		js.sinks.Store(subKeyPrefix, subscription.Spec.Sink)
	}

	// async callback for maxInflight messages
	callback := js.getCallback(subKeyPrefix, subscription.Name)
	asyncCallback := func(m *nats.Msg) {
		go callback(m)
	}

	js.bindConsumersForInvalidNATSSubscriptions(subscription, asyncCallback, log)
	if err := js.createConsumer(subscription, asyncCallback, log); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := utils.LoggerWithSubscription(js.namedLogger(), subscription)

	// loop over the global list of subscriptions
	// and delete any related JetStream subscription
	for key, jsSub := range js.subscriptions {
		if !js.isJsSubAssociatedWithKymaSub(key, subscription) {
			continue
		}
		if err := js.deleteSubscriptionFromJetStream(jsSub, key, log); err != nil {
			return err
		}
	}

	// cleanup consumers on nats-server
	// in-case data in js.subscriptions[] was lost due to handler restart
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetstreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if err := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName(), log); err != nil {
			return err
		}
	}

	// delete subscription sink info from storage
	js.sinks.Delete(createKeyPrefix(subscription))

	return nil
}

// GetJetStreamSubjects returns a list of subjects appended with prefix if needed
func (js *JetStream) GetJetStreamSubjects(subjects []string) []string {
	var result []string
	for _, subject := range subjects {
		result = append(result, js.GetJetstreamSubject(subject))
	}
	return result
}

// GetJetstreamSubject appends the prefix to subject.
func (js *JetStream) GetJetstreamSubject(subject string) string {
	return fmt.Sprintf("%s.%s", env.JetstreamSubjectPrefix, subject)
}

func (js *JetStream) validateConfig() error {
	if js.config.JSStreamName == "" {
		return errors.New("Stream name cannot be empty")
	}
	if len(js.config.JSStreamName) > jsMaxStreamNameLength {
		return fmt.Errorf("Stream name should be max %d characters long", jsMaxStreamNameLength)
	}
	if _, err := toJetStreamStorageType(js.config.JSStreamStorageType); err != nil {
		return err
	}
	if _, err := toJetStreamRetentionPolicy(js.config.JSStreamRetentionPolicy); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) handleReconnect(_ *nats.Conn) {
	js.namedLogger().Infow("called Reconnect handler for JetStream")
	if err := js.ensureStreamExists(); err != nil {
		js.namedLogger().Errorw("ensure stream exists failed", "error", err)
	}
}

func (js *JetStream) initNATSConn(connCloseHandler ConnClosedHandler) error {
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
		js.connClosedHandler = connCloseHandler
		if js.connClosedHandler != nil {
			js.conn.SetClosedHandler(nats.ConnHandler(js.connClosedHandler))
		}
		js.conn.SetReconnectHandler(js.handleReconnect)
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

func (js *JetStream) ensureStreamExists() error {
	if info, err := js.jsCtx.StreamInfo(js.config.JSStreamName); err == nil {
		// TODO: in case the stream exists, should we make sure all of its configs matches the stream config in the chart?
		js.namedLogger().Infow("reusing existing Stream", "stream-info", info)
		return nil
		// if nats server was restarted, the stream needs to be recreated for memory storage type
		// and hence we get ErrConnectionClosed for the lost stream
	} else if err != nats.ErrStreamNotFound {
		js.namedLogger().Debugw("The connection to NATs server is not established!")
		return err
	}
	streamConfig, err := getStreamConfig(js.config)
	if err != nil {
		return err
	}
	js.namedLogger().Infow("Stream not found, creating a new Stream",
		"streamName", js.config.JSStreamName, "streamStorageType", streamConfig.Storage, "subjects", streamConfig.Subjects)
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
		Replicas:  natsConfig.JSStreamReplicas,
		Retention: retentionPolicy,
		MaxMsgs:   natsConfig.JSStreamMaxMessages,
		MaxBytes:  natsConfig.JSStreamMaxBytes,
		// Since one stream is used to store events of all types, the stream has to match all event types, and therefore
		// we use the wildcard char >. However, to avoid matching internal JetStream and non-Kyma-related subjects, we
		// use a prefix. This prefix is handled only on the JetStream level (i.e. JetStream handler
		// and EPP) and should not be exposed in the Kyma subscription. Any Kyma event type gets appended with the
		// configured stream's subject prefix.
		Subjects: []string{fmt.Sprintf("%s.>", env.JetstreamSubjectPrefix)},
	}
	return streamConfig, nil
}

// syncSubscriptionFilters syncs the Kyma subscription filters with NATS subscriptions.
func (js *JetStream) syncSubscriptionFilters(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	for key, jsSub := range js.subscriptions {
		if !js.isJsSubAssociatedWithKymaSub(key, subscription) || !jsSub.IsValid() {
			continue
		}

		// TODO: optimize this call of ConsumerInfo
		// as jsSub.ConsumerInfo() will send an REST call to nats-server for each subject
		info, err := jsSub.ConsumerInfo()
		if err != nil {
			if err == nats.ErrConsumerNotFound {
				log.Infow("Deleting invalid Consumer!")
				if err := js.deleteConsumerFromJetStream(key.ConsumerName(), log); err != nil {
					return err
				}
				delete(js.subscriptions, key)
				continue
			}
			return err
		}

		if !utils.ContainsString(js.GetJetStreamSubjects(subscription.Status.CleanEventTypes), info.Config.FilterSubject) {
			if err := js.deleteSubscriptionFromJetStream(jsSub, key, log); err != nil {
				return err
			}
			log.Infow(
				"deleted JetStream subscription because it was deleted from subscription filters",
				"subscriptionSubject", key,
				"jetStreamSubject", jsSub.Subject,
			)
		}
	}
	return nil
}

// bindConsumersForInvalidNATSSubscriptions attempts to bind an existing consumer to a new NATS subscription,
// when the previous subscription that the consumer was associated with becomes invalid. If binding fails,
// we will delete the subscription from our internal subscriptions map.
func (js *JetStream) bindConsumersForInvalidNATSSubscriptions(subscription *eventingv1alpha1.Subscription, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetstreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)

		if existingNatsSub, ok := js.subscriptions[jsSubKey]; !ok {
			continue
		} else if existingNatsSub.IsValid() {
			log.Debugw("skipping creating subscription on JetStream because it already exists", "subject", subject)
			continue
		}
		log.Debugw("recreating subscription on JetStream because it was invalid", "subject", subject)
		// bind the existing consumer to a new subscription on JetStream
		jsSubscription, err := js.jsCtx.Subscribe(
			jsSubject,
			asyncCallback,
			nats.Bind(js.config.JSStreamName, jsSubKey.ConsumerName()),
		)
		if err != nil {
			if err != nats.ErrConsumerNotFound {
				log.Errorw("failed to bind subscription to an existing consumer", "subject", subject, "error", err)
			}
			delete(js.subscriptions, jsSubKey)
		} else {
			// save recreated JetStream subscription in storage
			js.subscriptions[jsSubKey] = jsSubscription
			log.Debugw("recreated subscription on JetStream", "subject", subject)
		}
	}
}

// createConsumer creates a new consumer on NATS for each CleanEventType,
// when there is no NATS subscription associated with the CleanEventType.
func (js *JetStream) createConsumer(subscription *eventingv1alpha1.Subscription, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) error {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetstreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)

		if _, ok := js.subscriptions[jsSubKey]; ok {
			continue
		}

		var jsSubscription *nats.Subscription
		var err error

		if jsSubject == "custom-subject" {
			log.Infow("Creating subscriber with DeliverPolicy: All on JetStream", "subject", subject)
			jsSubscription, err = js.jsCtx.Subscribe(
				jsSubject,
				asyncCallback,
				js.getCustomSubscriptionOptions(jsSubKey, subscription.Status.Config)...,
			)
		} else {
			log.Infow("Creating subscriber with DeliverPolicy: New on JetStream", "subject", subject)
			jsSubscription, err = js.jsCtx.Subscribe(
				jsSubject,
				asyncCallback,
				js.getDefaultSubscriptionOptions(jsSubKey, subscription.Status.Config)...,
			)
		}

		if err != nil {
			log.Errorw("failed to subscribe on JetStream", "subject", subject, "error", err)
			return err
		}
		// save created JetStream subscription in storage
		js.subscriptions[jsSubKey] = jsSubscription
		js.metricsCollector.RecordEventTypes(subscription.Name, subscription.Namespace, subject, jsSubKey.ConsumerName())
		log.Debugw("created subscription on JetStream", "subject", subject)
	}
	return nil
}

type DefaultSubOpts []nats.SubOpt

func (js *JetStream) getDefaultSubscriptionOptions(consumer SubscriptionSubjectIdentifier, subConfig *eventingv1alpha1.SubscriptionConfig) DefaultSubOpts {
	defaultOpts := DefaultSubOpts{
		nats.Durable(consumer.consumerName),
		nats.Description(consumer.namespacedSubjectName),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.IdleHeartbeat(idleHeartBeatDuration),
		nats.EnableFlowControl(),
		toJetStreamConsumerDeliverPolicyOptOrDefault(js.config.JSConsumerDeliverPolicy),
		nats.MaxAckPending(subConfig.MaxInFlightMessages),
		nats.MaxDeliver(jsConsumerMaxRedeliver),
		nats.AckWait(jsConsumerAcKWait),
	}
	return defaultOpts
}

func (js *JetStream) getCustomSubscriptionOptions(consumer SubscriptionSubjectIdentifier, subConfig *eventingv1alpha1.SubscriptionConfig) DefaultSubOpts {
	defaultOpts := DefaultSubOpts{
		nats.Durable(consumer.consumerName),
		nats.Description(consumer.namespacedSubjectName),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.IdleHeartbeat(idleHeartBeatDuration),
		nats.EnableFlowControl(),
		nats.DeliverAll(),
		nats.MaxAckPending(subConfig.MaxInFlightMessages),
		nats.MaxDeliver(jsConsumerMaxRedeliver),
		nats.AckWait(jsConsumerAcKWait),
	}
	return defaultOpts
}

func (js *JetStream) getCallback(subKeyPrefix, subscriptionName string) nats.MsgHandler {
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
			js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusInternalServerError)
			js.namedLogger().Errorw("failed to dispatch an event", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
			// Do not NAK the msg so that the server waits for AckWait and then redeliver the msg.
			return
		}

		// event was successfully dispatched, check if acknowledged by the NATS server
		// if not, the message is redelivered.
		if err := msg.Ack(); err != nil {
			js.namedLogger().Errorw("failed to ACK an event on JetStream", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
		}

		js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusOK)
		js.namedLogger().Infow("event dispatched", "id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)
	}
}

// isJsSubAssociatedWithKymaSub returns true if the given SubscriptionSubjectIdentifier and Kyma subscription
// have the same namespaced name, otherwise returns false.
func (js *JetStream) isJsSubAssociatedWithKymaSub(jsSubKey SubscriptionSubjectIdentifier, subscription *eventingv1alpha1.Subscription) bool {
	return createKeyPrefix(subscription) == jsSubKey.NamespacedName()
}

// deleteSubscriptionFromJS deletes subscription from JetStream and from in-memory db.
func (js *JetStream) deleteSubscriptionFromJetStream(jsSub *nats.Subscription, jsSubKey SubscriptionSubjectIdentifier, log *zap.SugaredLogger) error {
	// unsubscribe call to JetStream is async hence checking the status of the connection is important
	if err := js.checkJetStreamConnection(log); err != nil {
		return err
	}

	if jsSub.IsValid() {
		// unsubscribe will also delete the consumer on JS server
		if err := jsSub.Unsubscribe(); err != nil {
			log.Errorw("unsubscribe from JetStream failed", "error", err, "jsSub", jsSub)
			return errors.Wrapf(err, "unsubscribe failed")
		}
	} else {
		// if JS sub is not valid, then we need to delete the consumer on JetStream
		if err := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName(), log); err != nil {
			return err
		}
	}

	delete(js.subscriptions, jsSubKey)
	log.Debugw("unsubscribe from JetStream succeeded", "subscriptionSubject", jsSubKey)

	return nil
}

// deleteConsumerFromJS deletes consumer from JetStream.
func (js *JetStream) deleteConsumerFromJetStream(name string, log *zap.SugaredLogger) error {
	// checking the status of the connection is important
	if err := js.checkJetStreamConnection(log); err != nil {
		return err
	}

	if err := js.jsCtx.DeleteConsumer(js.config.JSStreamName, name); err != nil && err != nats.ErrConsumerNotFound {
		// if it is not a Not Found error, then return error
		log.Errorw("failed to delete consumer from JetStream", "error", err, "consumer", name)
		return err
	}

	return nil
}

// checkJetStreamConnection reconnects to the server if the server is not connected.
func (js *JetStream) checkJetStreamConnection(log *zap.SugaredLogger) error {
	if js.conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			log.Errorw("connect to JetStream failed", "status", js.conn.Status(), "error", err)
			return errors.Wrapf(err, "connect to JetStream failed")
		}
	}
	return nil
}

// GetAllSubscriptions returns the map which contains details of all subscriptions and consumers.
// Use this only for testing purposes.
func (js *JetStream) GetAllSubscriptions() map[SubscriptionSubjectIdentifier]*nats.Subscription {
	return js.subscriptions
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
