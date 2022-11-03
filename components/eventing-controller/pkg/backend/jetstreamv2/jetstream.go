package jetstreamv2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendutilsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils/v2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName          = "jetstream-handler"
	jsMaxStreamNameLength  = 32
	idleHeartBeatDuration  = 1 * time.Minute
	jsConsumerMaxRedeliver = 100
	jsConsumerAcKWait      = 30 * time.Second
)

func NewJetStream(config env.NatsConfig, metricsCollector *backendmetrics.Collector,
	cleaner cleaner.Cleaner, subsConfig env.DefaultSubscriptionConfig, logger *logger.Logger) *JetStream {
	return &JetStream{
		Config:           config,
		logger:           logger,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
		metricsCollector: metricsCollector,
		cleaner:          cleaner,
		subsConfig:       subsConfig,
	}
}

func (js *JetStream) Initialize(connCloseHandler backendutilsv2.ConnClosedHandler) error {
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
	subKeyPrefix := createKeyPrefix(subscription)
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if err := js.syncSubscriptionTypes(subscription); err != nil {
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

	if err := js.syncConsumerAndSubscription(subscription, asyncCallback); err != nil {
		return err
	}

	return nil
}

func (js *JetStream) DeleteSubscription(subscription *eventingv1alpha2.Subscription) error {
	// checking the status of the connection is important
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}
	// loop over the global list of subscriptions
	// and delete any related JetStream subscription
	for key, jsSub := range js.subscriptions {
		if !isJsSubAssociatedWithKymaSub(key, subscription) {
			continue
		}
		if err := js.deleteSubscriptionFromJetStream(jsSub, key); err != nil {
			return err
		}
	}

	// cleanup consumers on nats-server
	// in-case data in js.subscriptions[] was lost due to handler restart
	for _, subject := range subscription.Status.Types {
		jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
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
func (js *JetStream) GetJetStreamSubjects(source string, subjects []string,
	typeMatching eventingv1alpha2.TypeMatching) []string {
	var result []string
	for _, subject := range subjects {
		result = append(result, js.getJetStreamSubject(source, subject, typeMatching))
	}
	return result
}

// getJetStreamSubject appends the prefix and the cleaned source to subject.
func (js *JetStream) getJetStreamSubject(source, subject string, typeMatching eventingv1alpha2.TypeMatching) string {
	if typeMatching == eventingv1alpha2.TypeMatchingExact {
		return fmt.Sprintf("%s.%s", env.JetStreamSubjectPrefix, subject)
	}
	cleanSource, _ := js.cleaner.CleanSource(source)
	return fmt.Sprintf("%s.%s.%s", env.JetStreamSubjectPrefix, cleanSource, subject)
}

func (js *JetStream) validateConfig() error {
	if js.Config.JSStreamName == "" {
		return errors.New("Stream name cannot be empty")
	}
	if len(js.Config.JSStreamName) > jsMaxStreamNameLength {
		return fmt.Errorf("stream name should be max %d characters long", jsMaxStreamNameLength)
	}
	if _, err := toJetStreamStorageType(js.Config.JSStreamStorageType); err != nil {
		return err
	}
	if _, err := toJetStreamRetentionPolicy(js.Config.JSStreamRetentionPolicy); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) initNATSConn(connCloseHandler backendutilsv2.ConnClosedHandler) error {
	if js.Conn == nil || js.Conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.Config.MaxReconnects),
			nats.ReconnectWait(js.Config.ReconnectWait),
		}
		conn, err := nats.Connect(js.Config.URL, jsOptions...)
		if err != nil || !conn.IsConnected() {
			return fmt.Errorf("failed to connect to NATS JetStream: %w", err)
		}
		js.Conn = conn
		js.connClosedHandler = connCloseHandler
		if js.connClosedHandler != nil {
			js.Conn.SetClosedHandler(nats.ConnHandler(js.connClosedHandler))
		}
		js.Conn.SetReconnectHandler(js.handleReconnect)
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
	} else if !errors.Is(err, nats.ErrStreamNotFound) {
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
	jsCtx, err := js.Conn.JetStream()
	if err != nil {
		return fmt.Errorf("failed to create the JetStream context: %w", err)
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
	if js.Conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			return fmt.Errorf("failed to connect to JetStream with status %d: %w", js.Conn.Status(), err)
		}
	}
	return nil
}

// syncSubscriptionTypes syncs the Kyma subscription types with NATS subscriptions.
func (js *JetStream) syncSubscriptionTypes(subscription *eventingv1alpha2.Subscription) error {
	for key, jsSub := range js.subscriptions {
		err := js.syncSubscriptionType(key, subscription, jsSub)
		if err != nil {
			return err
		}
	}
	return nil
}

func (js *JetStream) syncSubscriptionType(key SubscriptionSubjectIdentifier,
	subscription *eventingv1alpha2.Subscription, subscriber Subscriber) error {
	if !isJsSubAssociatedWithKymaSub(key, subscription) || !subscriber.IsValid() {
		return nil
	}

	// TODO: optimize this call of ConsumerInfo
	// as jsSub.ConsumerInfo() will send an REST call to nats-server for each subject
	info, err := subscriber.ConsumerInfo()
	log := backendutilsv2.LoggerWithSubscription(js.namedLogger(), subscription)
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

func (js *JetStream) cleanupUnnecessaryJetStreamSubscribers(jsSub Subscriber,
	subscription *eventingv1alpha2.Subscription,
	log *zap.SugaredLogger, info *nats.ConsumerInfo, key SubscriptionSubjectIdentifier) error {
	if utils.ContainsString(
		js.GetJetStreamSubjects(
			subscription.Spec.Source,
			GetCleanEventTypesFromEventTypes(subscription.Status.Types),
			subscription.Spec.TypeMatching),
		info.Config.FilterSubject,
	) {
		return nil
	}
	log.Infow(
		"Deleting JetStream subscription because it was deleted from subscription types",
		"jetStreamSubject", info.Config.FilterSubject,
	)
	return js.deleteSubscriptionFromJetStream(jsSub, key)
}

// deleteSubscriptionFromJetStream deletes subscription from NATS server and from in-memory db.
func (js *JetStream) deleteSubscriptionFromJetStream(jsSub Subscriber, jsSubKey SubscriptionSubjectIdentifier) error {
	// unsubscribe call to JetStream is async hence checking the status of the connection is important
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if err := js.unsubscribeOnNats(jsSub, jsSubKey); err != nil {
		return err
	}

	delete(js.subscriptions, jsSubKey)
	return nil
}

// unsubscribeOnNats removes the subscription and consumer on NATS server.
func (js *JetStream) unsubscribeOnNats(jsSub Subscriber, jsSubKey SubscriptionSubjectIdentifier) error {
	if jsSub.IsValid() {
		// unsubscribe will also delete the consumer on JS server
		if err := jsSub.Unsubscribe(); err != nil {
			return utils.MakeSubscriptionError(ErrFailedUnsubscribe, err, jsSub)
		}
	}

	// delete the consumer manually, since it was created by hand, too
	if consDelErr := js.deleteConsumerFromJetStream(jsSubKey.ConsumerName()); consDelErr != nil {
		return consDelErr
	}

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
			ceLogger.Errorw("Failed to dispatch the CloudEvent")
			// Do not NAK the msg so that the server waits for AckWait and then redeliver the msg.
			return
		}

		// event was successfully dispatched, check if acknowledged by the NATS server
		// if not, the message is redelivered.
		if ackErr := msg.Ack(); ackErr != nil {
			ceLogger.Errorw("Failed to ACK an event on JetStream")
		}

		js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusOK)
		ceLogger.Infow("CloudEvent was dispatched")
	}
}

// deleteConsumerFromJS deletes consumer on NATS Server.
func (js *JetStream) deleteConsumerFromJetStream(name string) error {
	if err := js.jsCtx.DeleteConsumer(js.Config.JSStreamName, name); err != nil &&
		!errors.Is(err, nats.ErrConsumerNotFound) {
		// if it is not a Not Found error, then return error
		return fmt.Errorf("failed to delete consumer %s from JetStream: %w", name, err)
	}

	return nil
}

// syncConsumerAndSubscription makes sure there is a consumer and subscription created on the NATS Backend.
// these also must be bound to each other to ensure that NATS JetStream eventing logic works as expected.
func (js *JetStream) syncConsumerAndSubscription(subscription *eventingv1alpha2.Subscription,
	asyncCallback func(m *nats.Msg)) error {
	for _, eventType := range subscription.Status.Types {
		jsSubject := js.getJetStreamSubject(subscription.Spec.Source, eventType.CleanType, subscription.Spec.TypeMatching)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)

		consumerInfo, err := js.getOrCreateConsumer(subscription, eventType)
		if err != nil {
			return err
		}

		natsSubscription, subExists := js.subscriptions[jsSubKey]

		// try to create a NATS Subscription if it doesn't exist
		if !subExists && !consumerInfo.PushBound {
			if createErr := js.createNATSSubscription(subscription, eventType, asyncCallback); createErr != nil {
				return createErr
			}
		}

		if _, ok := js.subscriptions[jsSubKey]; !ok {
			return utils.MakeError(ErrMissingSubscription, err)
		}

		// try to bind invalid NATS Subscriptions
		if subExists && !natsSubscription.IsValid() {
			if bindErr := js.bindInvalidSubscriptions(subscription, eventType, asyncCallback); bindErr != nil {
				return bindErr
			}
		}

		// checks and updates the NATS consumer configs in case they are not up-to-date with the Subscription CR.
		if syncMaxInFlightErr := js.syncConsumerMaxInFlight(subscription, *consumerInfo); syncMaxInFlightErr != nil {
			return syncMaxInFlightErr
		}
	}
	return nil
}

// getOrCreateConsumer fetches the ConsumerInfo from NATS Server or creates it in case it doesn't exist.
func (js *JetStream) getOrCreateConsumer(subscription *eventingv1alpha2.Subscription,
	subject eventingv1alpha2.EventType) (*nats.ConsumerInfo, error) {
	jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)

	consumerInfo, err := js.jsCtx.ConsumerInfo(js.Config.JSStreamName, jsSubKey.ConsumerName())
	if err != nil {
		if errors.Is(err, nats.ErrConsumerNotFound) {
			consumerInfo, err = js.jsCtx.AddConsumer(
				js.Config.JSStreamName,
				js.getConsumerConfig(jsSubKey, jsSubject, subscription.GetMaxInFlightMessages(&js.subsConfig)),
			)
			if err != nil {
				return nil, utils.MakeError(ErrAddConsumer, err)
			}
		} else {
			return nil, utils.MakeError(ErrGetConsumer, err)
		}
	}
	return consumerInfo, nil
}

// createNATSSubscription creates a NATS Subscription and binds it to the already existing consumer.
func (js *JetStream) createNATSSubscription(subscription *eventingv1alpha2.Subscription,
	subject eventingv1alpha2.EventType, asyncCallback func(m *nats.Msg)) error {
	jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)

	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		js.getDefaultSubscriptionOptions(jsSubKey, subscription.GetMaxInFlightMessages(&js.subsConfig))...,
	)
	if err != nil {
		return utils.MakeError(ErrFailedSubscribe, err)
	}
	// save created JetStream subscription in storage
	js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
	js.metricsCollector.RecordEventTypes(
		subscription.Name,
		subscription.Namespace,
		subject.CleanType,
		jsSubKey.ConsumerName(),
	)

	return nil
}

// bindInvalidSubscriptions tries to bind the invalid NATS Subscription to the existing consumer.
func (js *JetStream) bindInvalidSubscriptions(subscription *eventingv1alpha2.Subscription,
	subject eventingv1alpha2.EventType, asyncCallback func(m *nats.Msg)) error {
	jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
	// bind the existing consumer to a new subscription on JetStream
	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		nats.Bind(js.Config.JSStreamName, jsSubKey.ConsumerName()),
	)
	if err != nil {
		return utils.MakeError(ErrFailedSubscribe, err)
	}
	// save recreated JetStream subscription in storage
	js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
	return nil
}

// syncConsumerMaxInFlight checks that the latest Subscription's maxInFlight value
// is propagated to the NATS consumer as MaxAckPending.
func (js *JetStream) syncConsumerMaxInFlight(subscription *eventingv1alpha2.Subscription,
	consumerInfo nats.ConsumerInfo) error {
	maxInFlight := subscription.GetMaxInFlightMessages(&js.subsConfig)

	if consumerInfo.Config.MaxAckPending == maxInFlight {
		return nil
	}

	// set the new maxInFlight value
	consumerConfig := consumerInfo.Config
	consumerConfig.MaxAckPending = maxInFlight

	// update the consumer
	if _, updateErr := js.jsCtx.UpdateConsumer(js.Config.JSStreamName, &consumerConfig); updateErr != nil {
		return utils.MakeError(ErrUpdateConsumer, updateErr)
	}
	return nil
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
