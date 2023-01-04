package jetstream

import (
	"context"
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"k8s.io/apimachinery/pkg/api/resource"

	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName                      = "jetstream-handler"
	idleHeartBeatDuration              = 1 * time.Minute
	jsConsumerMaxRedeliver             = 100
	jsConsumerAcKWait                  = 30 * time.Second
	jsMaxStreamNameLength              = 32
	separator                          = "/"
	MissingNATSSubscriptionMsg         = "failed to create NATS JetStream subscription"
	MissingNATSSubscriptionMsgWithInfo = MissingNATSSubscriptionMsg + " for subject: %v"
	RequeueDuration                    = 10 * time.Second
)

type Backend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(connCloseHandler backendnats.ConnClosedHandler) error

	// SyncSubscription should synchronize the Kyma eventing subscription with the subscriber infrastructure of JetStream.
	SyncSubscription(subscription *eventingv1alpha1.Subscription) error

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error

	// GetJetStreamSubjects returns a list of subjects appended with stream name as prefix if needed
	GetJetStreamSubjects(subjects []string) []string

	// DeleteInvalidConsumers deletes all JetStream consumers having no subscription types in subscription resources
	DeleteInvalidConsumers(subscriptions []eventingv1alpha1.Subscription) error

	// GetJetStreamContext returns the current JetStreamContext
	GetJetStreamContext() nats.JetStreamContext
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
	Config        backendnats.Config
	conn          *nats.Conn
	jsCtx         nats.JetStreamContext
	client        cev2.Client
	subscriptions map[SubscriptionSubjectIdentifier]backendnats.Subscriber
	sinks         sync.Map
	// connClosedHandler gets called by the NATS server when conn is closed and retry attempts are exhausted.
	connClosedHandler backendnats.ConnClosedHandler
	logger            *logger.Logger
	metricsCollector  *backendmetrics.Collector
	cleaner           eventtype.Cleaner
}

func NewJetStream(config backendnats.Config, metricsCollector *backendmetrics.Collector, jsCleaner eventtype.Cleaner,
	logger *logger.Logger) *JetStream {
	return &JetStream{
		Config:           config,
		logger:           logger,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]backendnats.Subscriber),
		metricsCollector: metricsCollector,
		cleaner:          jsCleaner,
	}
}

func (js *JetStream) initCloudEventClient(config backendnats.Config) error {
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

func (js *JetStream) Initialize(connCloseHandler backendnats.ConnClosedHandler) error {
	if err := js.validateConfig(); err != nil {
		return err
	}
	if err := js.initNATSConn(connCloseHandler); err != nil {
		return err
	}
	if err := js.initJSContext(); err != nil {
		return err
	}
	if err := js.initCloudEventClient(js.Config); err != nil {
		return err
	}
	return js.ensureStreamExistsAndIsConfiguredCorrectly()
}

func (js *JetStream) SyncSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := backendutils.LoggerWithSubscription(js.namedLogger(), subscription)
	subKeyPrefix := backendnats.CreateKeyPrefix(subscription)
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if err := js.syncSubscriptionEventFilters(subscription, log); err != nil {
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

	if err := js.checkSubscriptionConfig(subscription, log); err != nil {
		return err
	}

	if err := js.checkNATSSubscriptionsCount(subscription); err != nil {
		return err
	}

	return nil
}

func (js *JetStream) DeleteInvalidConsumers(subscriptions []eventingv1alpha1.Subscription) error {
	consumers := js.jsCtx.Consumers(js.Config.JSStreamName)
	for con := range consumers {
		// consumer should have no interest and subscription types to delete it
		if !con.PushBound && !js.isConsumerUsedByKymaSub(con.Name, subscriptions) {
			if err := js.deleteConsumerFromJetStream(con.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (js *JetStream) isConsumerUsedByKymaSub(consumerName string, subscriptions []eventingv1alpha1.Subscription) bool {
	if len(subscriptions) == 0 {
		return false
	}
	for ix := range subscriptions {
		// subjects := subscriptions[ix].Status.CleanEventTypes
		cleanedSubjects, err := backendnats.GetCleanSubjects(&subscriptions[ix], js.cleaner)
		if err != nil {
			js.namedLogger().Errorw("failed to clean subscription filter subjects", "error", err,
				"subscription namespace", &subscriptions[ix].Namespace, "subscription name", &subscriptions[ix].Name)
			return true
		}
		for _, subject := range cleanedSubjects {
			jsSubject := js.GetJetStreamSubject(subject)
			computedConsumerNameFromSubject := computeConsumerName(&subscriptions[ix], jsSubject)
			if consumerName == computedConsumerNameFromSubject {
				return true
			}
		}
	}
	return false
}

// checkSubscriptionConfig checks that the latest Subscription Config changes are propagated to the consumer.
// In our case config contains only the "maxInFlightMessages" property, which is the maxAckPending on the consumer side.
func (js *JetStream) checkSubscriptionConfig(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetStreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		log := log.With("subject", subject)

		if _, ok := js.subscriptions[jsSubKey]; !ok {
			continue
		}

		consumerInfo, err := js.jsCtx.ConsumerInfo(js.Config.JSStreamName, jsSubKey.ConsumerName())
		if err != nil && err != nats.ErrConsumerNotFound {
			log.Errorw("Failed to get the consumer info", "error", err)
			return err
		}

		subscriptionMaxInFlight := subscription.Status.Config.MaxInFlightMessages
		if subscription.Spec.Config != nil {
			subscriptionMaxInFlight = subscription.Spec.Config.MaxInFlightMessages
		}

		// skip the up-to-date consumers
		if consumerInfo.Config.MaxAckPending == subscriptionMaxInFlight {
			continue
		}

		// set the new maxInFlight value
		consumerConfig := consumerInfo.Config
		consumerConfig.MaxAckPending = subscriptionMaxInFlight

		// update the consumer
		if _, err := js.jsCtx.UpdateConsumer(js.Config.JSStreamName, &consumerConfig); err != nil {
			log.Errorw("Failed to update the consumer", "error", err)
			return err
		}
	}
	return nil
}

func (js *JetStream) DeleteSubscription(subscription *eventingv1alpha1.Subscription) error {
	log := backendutils.LoggerWithSubscription(js.namedLogger(), subscription)

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
		jsSubject := js.GetJetStreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if err := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName()); err != nil {
			return err
		}
	}

	// delete subscription sink info from storage
	js.sinks.Delete(backendnats.CreateKeyPrefix(subscription))

	return nil
}

// GetJetStreamSubjects returns a list of subjects appended with prefix if needed.
func (js *JetStream) GetJetStreamSubjects(subjects []string) []string {
	var result []string
	for _, subject := range subjects {
		result = append(result, js.GetJetStreamSubject(subject))
	}
	return result
}

// GetJetStreamSubject appends the prefix to subject.
func (js *JetStream) GetJetStreamSubject(subject string) string {
	return fmt.Sprintf("%s.%s", js.Config.JSSubjectPrefix, subject)
}

// GetJetStreamContext returns the current JetStreamContext.
func (js *JetStream) GetJetStreamContext() nats.JetStreamContext {
	return js.jsCtx
}

func (js *JetStream) validateConfig() error {
	if js.Config.JSStreamName == "" {
		return errors.New("Stream name cannot be empty")
	}
	if len(js.Config.JSStreamName) > jsMaxStreamNameLength {
		return xerrors.Errorf("Stream name should be max %d characters long", jsMaxStreamNameLength)
	}
	if _, err := toJetStreamStorageType(js.Config.JSStreamStorageType); err != nil {
		return err
	}
	if _, err := toJetStreamRetentionPolicy(js.Config.JSStreamRetentionPolicy); err != nil {
		return err
	}
	if _, err := toJetStreamDiscardPolicy(js.Config.JSStreamDiscardPolicy); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) handleReconnect(_ *nats.Conn) {
	js.namedLogger().Infow("Called reconnect handler for JetStream")
	if err := js.ensureStreamExistsAndIsConfiguredCorrectly(); err != nil {
		js.namedLogger().Errorw("Failed to ensure the stream exists", "error", err)
	}
}

func (js *JetStream) initNATSConn(connCloseHandler backendnats.ConnClosedHandler) error {
	if js.conn == nil || js.conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.Config.MaxReconnects),
			nats.ReconnectWait(js.Config.ReconnectWait),
			nats.Name("Kyma Controller"),
		}
		conn, err := nats.Connect(js.Config.URL, jsOptions...)
		if err != nil || !conn.IsConnected() {
			return xerrors.Errorf("failed to connect to NATS JetStream: %v", err)
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
		return xerrors.Errorf("failed to create the JetStream context: %v", err)
	}
	js.jsCtx = jsCtx
	return nil
}

func (js *JetStream) ensureStreamExistsAndIsConfiguredCorrectly() error {
	streamConfig, err := getStreamConfig(js.Config)
	if err != nil {
		return err
	}
	info, err := js.jsCtx.StreamInfo(js.Config.JSStreamName)
	if errors.Is(err, nats.ErrStreamNotFound) {
		info, err = js.jsCtx.AddStream(streamConfig)
		if err != nil {
			return err
		}
		js.namedLogger().Infow("Stream not found, created a new Stream",
			"stream-info", info)
		return nil
	}
	if err != nil {
		return err
	}

	if !streamIsConfiguredCorrectly(info.Config, *streamConfig) {
		newInfo, err := js.jsCtx.UpdateStream(streamConfig)
		if err != nil {
			return err
		}
		js.namedLogger().Infow("Updated existing Stream:", "stream-info", newInfo)
		return nil
	}

	js.namedLogger().Infow("Reusing existing Stream", "stream-info", info)
	return nil
}

func streamIsConfiguredCorrectly(got nats.StreamConfig, want nats.StreamConfig) bool {
	return reflect.DeepEqual(got, want)
}

func getStreamConfig(natsConfig backendnats.Config) (*nats.StreamConfig, error) {
	storage, err := toJetStreamStorageType(natsConfig.JSStreamStorageType)
	if err != nil {
		return nil, err
	}
	retentionPolicy, err := toJetStreamRetentionPolicy(natsConfig.JSStreamRetentionPolicy)
	if err != nil {
		return nil, err
	}

	// Quantities must not be empty. So we default here to "-1"
	if natsConfig.JSStreamMaxBytes == "" {
		natsConfig.JSStreamMaxBytes = "-1"
	}
	maxBytes, err := resource.ParseQuantity(natsConfig.JSStreamMaxBytes)
	if err != nil {
		return nil, err
	}
	discardPolicy, err := toJetStreamDiscardPolicy(natsConfig.JSStreamDiscardPolicy)
	if err != nil {
		return nil, err
	}
	streamConfig := &nats.StreamConfig{
		Name:      natsConfig.JSStreamName,
		Storage:   storage,
		Replicas:  natsConfig.JSStreamReplicas,
		Retention: retentionPolicy,
		MaxMsgs:   natsConfig.JSStreamMaxMessages,
		MaxBytes:  maxBytes.Value(),
		Discard:   discardPolicy,
		// Since one stream is used to store events of all types, the stream has to match all event types, and therefore
		// we use the wildcard char >. However, to avoid matching internal JetStream and non-Kyma-related subjects, we
		// use a prefix. This prefix is handled only on the JetStream level (i.e. JetStream handler
		// and EPP) and should not be exposed in the Kyma subscription. Any Kyma event type gets appended with the
		// configured stream's subject prefix.
		Subjects: []string{fmt.Sprintf("%s.>", natsConfig.JSSubjectPrefix)},
	}
	return streamConfig, nil
}

// syncSubscriptionEventFilters syncs the Kyma subscription filters with NATS subscriptions.
func (js *JetStream) syncSubscriptionEventFilters(subscription *eventingv1alpha1.Subscription,
	log *zap.SugaredLogger) error {
	for key, jsSub := range js.subscriptions {
		if js.isJsSubAssociatedWithKymaSub(key, subscription) {
			err := js.syncSubscriptionEventFilter(key, subscription, jsSub, log)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// syncSubscriptionEventFilter syncs controller runtime subscriptions to subscription CR event filters and to JetStream
// subscriptions/consumers.
func (js *JetStream) syncSubscriptionEventFilter(key SubscriptionSubjectIdentifier,
	subscription *eventingv1alpha1.Subscription, subscriber backendnats.Subscriber, log *zap.SugaredLogger) error {
	// don't try to delete invalid subscriber and its consumer if subscriber has type in subscription CR it belongs to.
	// This means that it will be bound to the existing JetStream consumer in later steps.
	if !subscriber.IsValid() && js.runtimeSubscriptionExistsInKymaSub(key, subscription) {
		return nil
	}

	return js.cleanupUnnecessaryJetStreamSubscribers(subscriber, subscription, log, key)
}

// runtimeSubscriptionExistsInKymaSub returns true if runtime subscriber subject exists in subscription CR.
func (js *JetStream) runtimeSubscriptionExistsInKymaSub(cachedSubscriptionKey SubscriptionSubjectIdentifier,
	subscription *eventingv1alpha1.Subscription) bool {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetStreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if cachedSubscriptionKey.consumerName == jsSubKey.consumerName {
			return true
		}
	}
	return false
}

func (js *JetStream) cleanupUnnecessaryJetStreamSubscribers(
	jsSub backendnats.Subscriber,
	subscription *eventingv1alpha1.Subscription,
	log *zap.SugaredLogger,
	key SubscriptionSubjectIdentifier) error {
	consumer, err := js.jsCtx.ConsumerInfo(js.Config.JSStreamName, key.ConsumerName())
	if err != nil {
		if errors.Is(err, nats.ErrConsumerNotFound) {
			log.Infow("Deleting invalid Consumer!")
			if err = js.deleteConsumerFromJetStream(key.ConsumerName()); err != nil {
				return err
			}
			delete(js.subscriptions, key)
			return nil
		}
		return err
	}

	// delete NATS consumer if it doesn't exist in subscription CR
	if !js.consumerSubjectExistsInKymaSub(consumer, subscription) {
		log.Infow(
			"Deleting JetStream subscription because it was deleted from subscription filters",
			"subscriptionSubject", key,
			"jetStreamSubject", jsSub.SubscriptionSubject(),
		)
		return js.deleteSubscriptionFromJetStream(jsSub, key, log)
	}
	return nil
}

// consumerSubjectExistsInKymaSub checks if the specified consumer is used by the subscription.
func (js *JetStream) consumerSubjectExistsInKymaSub(consumer *nats.ConsumerInfo,
	subscription *eventingv1alpha1.Subscription) bool {
	return utils.ContainsString(
		js.GetJetStreamSubjects(subscription.Status.CleanEventTypes),
		consumer.Config.FilterSubject)
}

// bindConsumersForInvalidNATSSubscriptions attempts to bind an existing consumer to a new NATS subscription,
// when the previous subscription that the consumer was associated with becomes invalid. If binding fails,
// we will delete the subscription from our internal subscriptions map.
func (js *JetStream) bindConsumersForInvalidNATSSubscriptions(subscription *eventingv1alpha1.Subscription, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetStreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		log := log.With("subject", subject)

		if existingNatsSub, ok := js.subscriptions[jsSubKey]; !ok {
			continue
		} else if existingNatsSub.IsValid() {
			log.Debugw("Skipping creation subscription on JetStream because it already exists")
			continue
		}
		log.Debugw("Recreating subscription on JetStream because it was invalid")
		// bind the existing consumer to a new subscription on JetStream
		jsSubscription, err := js.jsCtx.Subscribe(
			jsSubject,
			asyncCallback,
			nats.Bind(js.Config.JSStreamName, jsSubKey.ConsumerName()),
		)
		if err != nil {
			if err != nats.ErrConsumerNotFound {
				log.Errorw("Failed to bind subscription to an existing consumer", "error", err)
			}
			delete(js.subscriptions, jsSubKey)
		} else {
			// save recreated JetStream subscription in storage
			js.subscriptions[jsSubKey] = &backendnats.Subscription{Subscription: jsSubscription}
			log.Debugw("Recreated subscription on JetStream")
		}
	}
}

// createConsumer creates a new consumer on NATS for each CleanEventType,
// when there is no NATS subscription associated with the CleanEventType.
func (js *JetStream) createConsumer(subscription *eventingv1alpha1.Subscription, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) error {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetStreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		log := log.With("subject", subject)

		if _, ok := js.subscriptions[jsSubKey]; ok {
			continue
		}

		// TODO: optimize this call of ConsumerInfo
		consumerInfo, err := js.jsCtx.ConsumerInfo(js.Config.JSStreamName, jsSubKey.ConsumerName())
		if err != nil && err != nats.ErrConsumerNotFound {
			log.Errorw("Failed to get consumer info", "error", err)
			return err
		}

		// skip the subject, in case there is a bound consumer on NATS
		if consumerInfo != nil && consumerInfo.PushBound {
			continue
		}

		jsSubscription, err := js.jsCtx.Subscribe(
			jsSubject,
			asyncCallback,
			js.getDefaultSubscriptionOptions(jsSubKey, subscription.Status.Config)...,
		)
		if err != nil {
			return xerrors.Errorf("failed to subscribe on JetStream: %v", err)
		}
		// save created JetStream subscription in storage
		js.subscriptions[jsSubKey] = &backendnats.Subscription{Subscription: jsSubscription}
		js.metricsCollector.RecordEventTypes(subscription.Name, subscription.Namespace, subject, jsSubKey.ConsumerName())
		log.Debugw("Created subscription on JetStream")
	}
	return nil
}

// checkNATSSubscriptionsCount checks whether NATS Subscription(s) were created for all the Kyma Subscription filters
func (js *JetStream) checkNATSSubscriptionsCount(subscription *eventingv1alpha1.Subscription) error {
	for _, subject := range subscription.Status.CleanEventTypes {
		jsSubject := js.GetJetStreamSubject(subject)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if _, ok := js.subscriptions[jsSubKey]; !ok {
			return errors.Errorf(MissingNATSSubscriptionMsgWithInfo, subject)
		}
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
		toJetStreamConsumerDeliverPolicyOptOrDefault(js.Config.JSConsumerDeliverPolicy),
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
			js.namedLogger().Errorw("Failed to find sink URL in storage", "keyPrefix", subKeyPrefix)
			return
		}
		// convert interface type to string
		sink, ok := sinkValue.(string)
		if !ok {
			js.namedLogger().Errorw("Failed to convert sink value to string", "sinkValue", sinkValue)
			return
		}
		ce, err := backendutils.ConvertMsgToCE(msg)
		if err != nil {
			js.namedLogger().Errorw("Failed to convert JetStream message to CloudEvent", "error", err)
			return
		}

		// setup context for dispatching
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctxWithCE := cev2.ContextWithTarget(ctxWithCancel, sink)
		traceCtxWithCE := tracing.AddTracingHeadersToContext(ctxWithCE, ce)

		// decorate the logger with CloudEvent context
		ceLogger := js.namedLogger().With("id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)

		ceLogger.Debugw("Sending the CloudEvent")

		// dispatch the event to sink
		result := js.client.Send(traceCtxWithCE, *ce)
		if !cev2protocol.IsACK(result) {
			js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusInternalServerError)
			ceLogger.Errorw("Failed to dispatch the CloudEvent", "error", result.Error())
			// Do not NAK the msg so that the server waits for AckWait and then redeliver the msg.
			return
		}

		// event was successfully dispatched, check if acknowledged by the NATS server
		// if not, the message is redelivered.
		if err := msg.Ack(); err != nil {
			ceLogger.Errorw("Failed to ACK an event on JetStream")
		}

		js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusOK)
		ceLogger.Infow("CloudEvent was dispatched")
	}
}

// isJsSubAssociatedWithKymaSub returns true if the given SubscriptionSubjectIdentifier and Kyma subscription
// have the same namespaced name, otherwise returns false.
func (js *JetStream) isJsSubAssociatedWithKymaSub(jsSubKey SubscriptionSubjectIdentifier, subscription *eventingv1alpha1.Subscription) bool {
	return backendnats.CreateKeyPrefix(subscription) == jsSubKey.NamespacedName()
}

// deleteSubscriptionFromJS deletes subscription from JetStream and from in-memory db.
func (js *JetStream) deleteSubscriptionFromJetStream(jsSub backendnats.Subscriber, jsSubKey SubscriptionSubjectIdentifier, log *zap.SugaredLogger) error {
	// unsubscribe call to JetStream is async hence checking the status of the connection is important
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if jsSub.IsValid() {
		// unsubscribe will also delete the consumer on JS server
		if err := jsSub.Unsubscribe(); err != nil {
			return xerrors.Errorf("failed to unsubscribe subscription %v from JetStream: %v", jsSub, err)
		}
	}

	if err := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName()); err != nil &&
		!errors.Is(err, nats.ErrConsumerNotFound) {
		return fmt.Errorf("failed to delete consumer %s from JetStream: %w", jsSubKey.ConsumerName(), err)
	}

	delete(js.subscriptions, jsSubKey)
	log.Debugw("Unsubscribed from JetStream", "subscriptionSubject", jsSubKey)

	return nil
}

// deleteConsumerFromJS deletes consumer from JetStream.
func (js *JetStream) deleteConsumerFromJetStream(name string) error {
	// checking the status of the connection is important
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if err := js.jsCtx.DeleteConsumer(js.Config.JSStreamName, name); err != nil && err != nats.ErrConsumerNotFound {
		// if it is not a Not Found error, then return error
		return xerrors.Errorf("failed to delete consumer %s from JetStream: %v", name, err)
	}

	return nil
}

// checkJetStreamConnection reconnects to the server if the server is not connected.
func (js *JetStream) checkJetStreamConnection() error {
	if js.conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			return xerrors.Errorf("failed to connect to JetStream with status %d: %v", js.conn.Status(), err)
		}
	}
	return nil
}

// GetAllSubscriptions returns the map which contains details of all subscriptions and consumers.
// Use this only for testing purposes.
func (js *JetStream) GetAllSubscriptions() map[SubscriptionSubjectIdentifier]backendnats.Subscriber {
	return js.subscriptions
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
