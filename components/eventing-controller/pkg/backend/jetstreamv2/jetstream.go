package jetstreamv2

import (
	"context"
	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendutilsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utilsv2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"net/http"
	"time"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName                      = "jetstream-handler"
	jsMaxStreamNameLength              = 32
	MissingNATSSubscriptionMsg         = "failed to create NATS JetStream subscription"
	MissingNATSSubscriptionMsgWithInfo = MissingNATSSubscriptionMsg + " for subject: %v"
	idleHeartBeatDuration              = 1 * time.Minute
	jsConsumerMaxRedeliver             = 100
	jsConsumerAcKWait                  = 30 * time.Second
)

func NewJetStream(config env.NatsConfig, metricsCollector *backendmetrics.Collector, logger *logger.Logger) *JetStream {
	return &JetStream{
		Config:           config,
		logger:           logger,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
		metricsCollector: metricsCollector,
	}
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
	if err := js.initCloudEventClient(js.Config); err != nil {
		return err
	}
	return js.ensureStreamExists()
}

func (js *JetStream) SyncSubscription(subscription *eventingv1alpha2.Subscription) error {
	log := backendutilsv2.LoggerWithSubscription(js.namedLogger(), subscription)
	subKeyPrefix := createKeyPrefix(subscription)
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if err := js.syncSubscriptionTypes(subscription, log); err != nil {
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

	if err := js.checkNATSSubscriptionsCount(subscription); err != nil {
		return err
	}

	return nil
}

func (js *JetStream) DeleteSubscription(subscription *eventingv1alpha2.Subscription) error {
	log := backendutilsv2.LoggerWithSubscription(js.namedLogger(), subscription)

	// loop over the global list of subscriptions
	// and delete any related JetStream subscription
	for key, jsSub := range js.subscriptions {
		if !isJsSubAssociatedWithKymaSub(key, subscription) {
			continue
		}
		if err := js.deleteSubscriptionFromJetStream(jsSub, key, log); err != nil {
			return err
		}
	}

	// cleanup consumers on nats-server
	// in-case data in js.subscriptions[] was lost due to handler restart
	for _, subject := range subscription.Status.Types {
		jsSubject := getJetStreamSubject(subject.CleanType)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if err := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName()); err != nil {
			return err
		}
	}

	// delete subscription sink info from storage
	js.sinks.Delete(createKeyPrefix(subscription))

	return nil
}

// GetJetStreamSubjects returns a list of subjects appended with prefix if needed.
func (js *JetStream) GetJetStreamSubjects(subjects []string) []string {
	var result []string
	for _, subject := range subjects {
		result = append(result, getJetStreamSubject(subject))
	}
	return result
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
	return nil
}

func (js *JetStream) initNATSConn(connCloseHandler ConnClosedHandler) error {
	if js.conn == nil || js.conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.Config.MaxReconnects),
			nats.ReconnectWait(js.Config.ReconnectWait),
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

func (js *JetStream) handleReconnect(_ *nats.Conn) {
	js.namedLogger().Infow("Called reconnect handler for JetStream")
	if err := js.ensureStreamExists(); err != nil {
		js.namedLogger().Errorw("Failed to ensure the stream exists", "error", err)
	}
}

func (js *JetStream) ensureStreamExists() error {
	if info, err := js.jsCtx.StreamInfo(js.Config.JSStreamName); err == nil {
		// TODO: in case the stream exists, should we make sure all of its configs matches the stream config in the chart?
		js.namedLogger().Infow("Reusing existing Stream", "stream-info", info)
		return nil
		// if nats server was restarted, the stream needs to be recreated for memory storage type
		// and hence we get ErrConnectionClosed for the lost stream
	} else if err != nats.ErrStreamNotFound {
		js.namedLogger().Debugw("The connection to NATS server is not established!")
		return err
	}
	streamConfig, err := getStreamConfig(js.Config)
	if err != nil {
		return err
	}
	js.namedLogger().Infow("Stream not found, creating a new Stream",
		"streamName", js.Config.JSStreamName, "streamStorageType", streamConfig.Storage, "subjects", streamConfig.Subjects)
	_, err = js.jsCtx.AddStream(streamConfig)
	return err
}

func (js *JetStream) initJSContext() error {
	jsCtx, err := js.conn.JetStream()
	if err != nil {
		return xerrors.Errorf("failed to create the JetStream context: %v", err)
	}
	js.jsCtx = jsCtx
	return nil
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

// checkJetStreamConnection reconnects to the server if the server is not connected.
func (js *JetStream) checkJetStreamConnection() error {
	if js.conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			return xerrors.Errorf("failed to connect to JetStream with status %d: %v", js.conn.Status(), err)
		}
	}
	return nil
}

// syncSubscriptionTypes syncs the Kyma subscription types with NATS subscriptions.
func (js *JetStream) syncSubscriptionTypes(subscription *eventingv1alpha2.Subscription, log *zap.SugaredLogger) error {
	for key, jsSub := range js.subscriptions {
		err := js.syncSubscriptionType(key, subscription, jsSub, log)
		if err != nil {
			return err
		}
	}
	return nil
}

func (js *JetStream) syncSubscriptionType(key SubscriptionSubjectIdentifier, subscription *eventingv1alpha2.Subscription, subscriber Subscriber, log *zap.SugaredLogger) error {
	if !isJsSubAssociatedWithKymaSub(key, subscription) || !subscriber.IsValid() {
		return nil
	}

	// TODO: optimize this call of ConsumerInfo
	// as jsSub.ConsumerInfo() will send an REST call to nats-server for each subject
	info, err := subscriber.ConsumerInfo()
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

	err = js.cleanupUnnecessaryJetStreamSubscribers(subscriber, subscription, log, info, key)
	if err != nil {
		return err
	}
	return nil
}

func (js *JetStream) cleanupUnnecessaryJetStreamSubscribers(jsSub Subscriber, subscription *eventingv1alpha2.Subscription, log *zap.SugaredLogger, info *nats.ConsumerInfo, key SubscriptionSubjectIdentifier) error {
	if utils.ContainsString(js.GetJetStreamSubjects(getCleanEventTypesFromStatus(subscription.Status)), info.Config.FilterSubject) {
		return nil
	}
	log.Infow(
		"Deleting JetStream subscription because it was deleted from subscription types",
		"subscriptionSubject", key,
		"jetStreamSubject", jsSub.SubscriptionSubject(),
	)
	return js.deleteSubscriptionFromJetStream(jsSub, key, log)
}

// deleteSubscriptionFromJS deletes subscription from JetStream and from in-memory db.
func (js *JetStream) deleteSubscriptionFromJetStream(jsSub Subscriber, jsSubKey SubscriptionSubjectIdentifier, log *zap.SugaredLogger) error {
	// unsubscribe call to JetStream is async hence checking the status of the connection is important
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if jsSub.IsValid() {
		// unsubscribe will also delete the consumer on JS server
		if err := jsSub.Unsubscribe(); err != nil {
			return xerrors.Errorf("failed to unsubscribe subscription %v from JetStream: %v", jsSub, err)
		}
	} else {
		// if JS sub is not valid, then we need to delete the consumer on JetStream
		if err := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName()); err != nil {
			return err
		}
	}

	delete(js.subscriptions, jsSubKey)
	log.Debugw("Unsubscribed from JetStream", "subscriptionSubject", jsSubKey)

	return nil
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
		ce, err := convertMsgToCE(msg)
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
			ceLogger.Errorw("Failed to dispatch the CloudEvent")
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

// bindConsumersForInvalidNATSSubscriptions attempts to bind an existing consumer to a new NATS subscription,
// when the previous subscription that the consumer was associated with becomes invalid. If binding fails,
// we will delete the subscription from our internal subscriptions map.
func (js *JetStream) bindConsumersForInvalidNATSSubscriptions(subscription *eventingv1alpha2.Subscription, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) {
	for _, subject := range subscription.Status.Types {
		jsSubject := getJetStreamSubject(subject.CleanType)
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
			js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
			log.Debugw("Recreated subscription on JetStream")
		}
	}
}

// createConsumer creates a new consumer on NATS for each CleanEventType,
// when there is no NATS subscription associated with the CleanEventType.
func (js *JetStream) createConsumer(subscription *eventingv1alpha2.Subscription, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) error {
	for _, subject := range subscription.Status.Types {
		jsSubject := getJetStreamSubject(subject.CleanType)
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
			js.getDefaultSubscriptionOptions(jsSubKey, subscription.Status.Backend.MaxInFlightMessages)...,
		)
		if err != nil {
			return xerrors.Errorf("failed to subscribe on JetStream: %v", err)
		}
		// save created JetStream subscription in storage
		js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
		js.metricsCollector.RecordEventTypes(subscription.Name, subscription.Namespace, subject.CleanType, jsSubKey.ConsumerName())
		log.Debugw("Created subscription on JetStream")
	}
	return nil
}

// checkNATSSubscriptionsCount checks whether NATS Subscription(s) were created for all the Kyma Subscription types
func (js *JetStream) checkNATSSubscriptionsCount(subscription *eventingv1alpha2.Subscription) error {
	for _, subject := range subscription.Status.Types {
		jsSubject := getJetStreamSubject(subject.CleanType)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if _, ok := js.subscriptions[jsSubKey]; !ok {
			return errors.Errorf(MissingNATSSubscriptionMsgWithInfo, subject.CleanType)
		}
	}
	return nil
}

func (js *JetStream) getDefaultSubscriptionOptions(consumer SubscriptionSubjectIdentifier, maxInFlightMessages int) DefaultSubOpts {
	defaultOpts := DefaultSubOpts{
		nats.Durable(consumer.consumerName),
		nats.Description(consumer.namespacedSubjectName),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.IdleHeartbeat(idleHeartBeatDuration),
		nats.EnableFlowControl(),
		toJetStreamConsumerDeliverPolicyOptOrDefault(js.Config.JSConsumerDeliverPolicy),
		nats.MaxAckPending(maxInFlightMessages),
		nats.MaxDeliver(jsConsumerMaxRedeliver),
		nats.AckWait(jsConsumerAcKWait),
	}
	return defaultOpts
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
