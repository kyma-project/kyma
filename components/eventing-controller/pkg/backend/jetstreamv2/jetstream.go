package jetstreamv2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	backenderrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2/errors"

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
	"golang.org/x/xerrors"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName          = "jetstream-handler"
	jsMaxStreamNameLength  = 32
	idleHeartBeatDuration  = 1 * time.Minute
	jsConsumerMaxRedeliver = 100
	jsConsumerAcKWait      = 30 * time.Second
)

var (
	// todo refine and check all the errors again
	ErrConsumerInfo   = errors.New("failed to fetch consumer info")
	ErrConsumerAdd    = errors.New("failed to add consumer")
	ErrConsumerUpdate = errors.New("failed to update consumer")
)

func NewJetStream(config env.NatsConfig, metricsCollector *backendmetrics.Collector, cleaner cleaner.Cleaner, logger *logger.Logger) *JetStream {
	return &JetStream{
		Config:           config,
		logger:           logger,
		subscriptions:    make(map[SubscriptionSubjectIdentifier]Subscriber),
		metricsCollector: metricsCollector,
		cleaner:          cleaner,
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

	if err := js.syncNATSConsumers(subscription, asyncCallback); err != nil {
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
func (js *JetStream) GetJetStreamSubjects(source string, subjects []string, typeMatching eventingv1alpha2.TypeMatching) []string {
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
	if js.Conn == nil || js.Conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.Config.MaxReconnects),
			nats.ReconnectWait(js.Config.ReconnectWait),
		}
		conn, err := nats.Connect(js.Config.URL, jsOptions...)
		if err != nil || !conn.IsConnected() {
			return xerrors.Errorf("failed to connect to NATS JetStream: %v", err)
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
	jsCtx, err := js.Conn.JetStream()
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
	if js.Conn.Status() != nats.CONNECTED {
		if err := js.Initialize(js.connClosedHandler); err != nil {
			return xerrors.Errorf("failed to connect to JetStream with status %d: %v", js.Conn.Status(), err)
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

func (js *JetStream) syncSubscriptionType(key SubscriptionSubjectIdentifier, subscription *eventingv1alpha2.Subscription, subscriber Subscriber) error {
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

func (js *JetStream) cleanupUnnecessaryJetStreamSubscribers(jsSub Subscriber, subscription *eventingv1alpha2.Subscription, log *zap.SugaredLogger, info *nats.ConsumerInfo, key SubscriptionSubjectIdentifier) error {
	if utils.ContainsString(js.GetJetStreamSubjects(subscription.Spec.Source, getCleanEventTypesFromStatus(subscription.Status), subscription.Spec.TypeMatching), info.Config.FilterSubject) {
		return nil
	}
	log.Infow(
		"Deleting JetStream subscription because it was deleted from subscription types",
		"jetStreamSubject", info.Config.FilterSubject,
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

// syncNATSConsumers makes sure there are consumers and subscriptions created on the NATS Backend.
func (js *JetStream) syncNATSConsumers(subscription *eventingv1alpha2.Subscription, asyncCallback func(m *nats.Msg)) error {
	for _, subject := range subscription.Status.Types {
		jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		log := backendutilsv2.LoggerWithSubscriptionAndEventType(js.namedLogger(), subscription, subject)

		consumerInfo, err := js.jsCtx.ConsumerInfo(js.Config.JSStreamName, jsSubKey.ConsumerName())
		if err != nil {
			// create the consumer in case it doesn't exist
			if errors.Is(err, nats.ErrConsumerNotFound) {
				consumerInfo, err = js.jsCtx.AddConsumer(
					js.Config.JSStreamName,
					js.getConsumerConfig(subscription, jsSubKey, jsSubject),
				)
				if err != nil {
					return fmt.Errorf("%w: %v", ErrConsumerAdd, err)
				}
				log.Debug("Created consumer on JetStream")
			} else {
				return fmt.Errorf("%w: %v", ErrConsumerInfo, err)
			}
		}

		natsSubscription, subExists := js.subscriptions[jsSubKey]

		// try to create a NATS Subscription if it doesn't exist
		if !subExists && !consumerInfo.PushBound {
			if err := js.createNATSSubscription(subscription, subject, jsSubject, jsSubKey, asyncCallback, log); err != nil {
				return err
			}
		}

		// try to bind invalid NATS Subscriptions
		if subExists && !natsSubscription.IsValid() {
			if err := js.bindConsumersForInvalidNATSSubscriptions(jsSubject, jsSubKey, asyncCallback, log); err != nil {
				return err
			}
		}

		if consumerInfo != nil {
			if err := js.checkSubscriptionConfig(subscription, *consumerInfo, log); err != nil {
				return err
			}
		}
	}
	if err := js.checkNATSSubscriptionsCount(subscription); err != nil {
		return err
	}

	return nil
}

// createNATSSubscription creates a NATS Subscription and binds it to the already existing consumer.
func (js *JetStream) createNATSSubscription(subscription *eventingv1alpha2.Subscription, subject eventingv1alpha2.EventType, jsSubject string, jsSubKey SubscriptionSubjectIdentifier, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) error {
	maxInFlight := subscription.GetMaxInFlightMessages(js.namedLogger())
	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		js.getDefaultSubscriptionOptions(jsSubKey, maxInFlight)...,
	)
	if err != nil {
		return backenderrors.NewFailedToSubscribeOnNATSError(err)
	}
	// save created JetStream subscription in storage
	js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
	js.metricsCollector.RecordEventTypes(subscription.Name, subscription.Namespace, subject.CleanType, jsSubKey.ConsumerName())
	log.Debugw("Created subscription on JetStream")

	return nil
}

// bindConsumersForInvalidNATSSubscriptions tries to bind the invalid NATS Subscription to the existing consumer.
func (js *JetStream) bindConsumersForInvalidNATSSubscriptions(jsSubject string, jsSubKey SubscriptionSubjectIdentifier, asyncCallback func(m *nats.Msg), log *zap.SugaredLogger) error {
	log.Debugw("Recreating subscription on JetStream because it was invalid")
	// bind the existing consumer to a new subscription on JetStream
	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		nats.Bind(js.Config.JSStreamName, jsSubKey.ConsumerName()),
	)
	if err != nil {
		return backenderrors.NewFailedToSubscribeOnNATSError(err)
	}
	// save recreated JetStream subscription in storage
	js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
	log.Debugw("Recreated subscription on JetStream")
	return nil
}

// checkSubscriptionConfig checks that the latest Subscription Config changes are propagated to the consumer.
// In our case config contains only the "maxInFlightMessages" property, which is the maxAckPending on the consumer side.
func (js *JetStream) checkSubscriptionConfig(subscription *eventingv1alpha2.Subscription, consumerInfo nats.ConsumerInfo, log *zap.SugaredLogger) error {
	// skip the up-to-date consumers
	subMaxInFlight := subscription.GetMaxInFlightMessages(js.namedLogger())
	if consumerInfo.Config.MaxAckPending == subMaxInFlight {
		return nil
	}

	// set the new maxInFlight value
	consumerConfig := consumerInfo.Config
	consumerConfig.MaxAckPending = subMaxInFlight

	// update the consumer
	if _, err := js.jsCtx.UpdateConsumer(js.Config.JSStreamName, &consumerConfig); err != nil {
		return fmt.Errorf("%w: %v", ErrConsumerUpdate, err)
	}
	return nil
}

// checkNATSSubscriptionsCount checks whether NATS Subscription(s) were created for all the Kyma Subscription types
func (js *JetStream) checkNATSSubscriptionsCount(subscription *eventingv1alpha2.Subscription) error {
	for _, subject := range subscription.Status.Types {
		jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if _, ok := js.subscriptions[jsSubKey]; !ok {
			return backenderrors.NewMissingNatsSubscriptionError(subject)
		}
	}
	return nil
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
