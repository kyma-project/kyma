package jetstreamv2

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/cloudevent"
	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendutilsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils/v2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName          = "jetstream-handler"
	jsMaxStreamNameLength  = 32
	idleHeartBeatDuration  = 1 * time.Minute
	jsConsumerMaxRedeliver = 100
	jsConsumerAcKWait      = 30 * time.Second
	originalTypeHeaderName = "originaltype"
)

func NewJetStream(config backendnats.Config, metricsCollector *backendmetrics.Collector, cleaner cleaner.Cleaner,
	subsConfig env.DefaultSubscriptionConfig, logger logger.KLogger) *JetStream {
	return &JetStream{
		logger:            logger,
		subscriptions:     make(map[SubscriptionSubjectIdentifier]Subscriber),
		metricsCollector:  metricsCollector,
		cleaner:           cleaner,
		connectionBuilder: NewConnectionBuilder(config),
		config:            config,
		ceClientFactory:   cloudevent.ClientFactory{},
		subsConfig:        subsConfig,
	}
}

// TODO(nils): In theory I don't see a reason why the integration tests need to know about the connection status of
// the backend. Talk to @raypinto how we can resolve this problem.
// To keep the tests green, for now I leave it as it is.
func (js *JetStream) IsConnected() bool {
	return js.connection.IsConnected()
}

// TODO(nils): I don't see a reason why the integration tests need to get the config from the backend. Since the tests
// start the backend, they should already be aware of the config.
// To keep the tests green, for now I leave it as it is.
// Another solution might be to use a singleton for the config and don't access it from the JetStream backend.
func (js *JetStream) GetConfig() *backendnats.Config {
	return &js.config
}

// Initialize the NATS JetStream backend by:
// - Establishing a connection to the backend.
// - Creating a stream.
// - Updating the stream config (if it changed).
func (js *JetStream) Initialize(handleConnectionClosedEvent backendutilsv2.ConnClosedHandler) error {
	config := js.config

	if err := js.initNATSConn(); err != nil {
		return err
	}
	js.setConnHandlers(handleConnectionClosedEvent)
	if err := js.initJSContext(); err != nil {
		return err
	}
	if err := js.initCloudEventClient(config); err != nil {
		return err
	}
	streamInfo, streamConfig, err := js.ensureStreamExists()
	if err != nil {
		return err
	}
	// Try to update the stream configuration (if necessary).
	if err = js.ensureCorrectStreamConfiguration(streamInfo, streamConfig); err != nil {
		return err
	}
	return nil
}

// initNATSConn initializes a connection to the NATS JetStream server and saves the connection as JetStream.connection.
func (js *JetStream) initNATSConn() error {
	if js.connection != nil {
		return nil
	}

	c, err := js.connectionBuilder.Build()
	if err != nil {
		return err
	}

	js.connection = c
	return nil
}

// setConnHandlers sets the handlers for the following events:
// - connection closed: this can be used by external services to be notified and act upon the event.
// - connection is in reconnection mode: in this case we ensure the stream is created and correctly configured.
func (js *JetStream) setConnHandlers(connCloseHandler backendutilsv2.ConnClosedHandler) {
	if js.handleConnectionClosedEvent == nil {
		js.handleConnectionClosedEvent = connCloseHandler
		js.connection.SetClosedHandler(nats.ConnHandler(js.handleConnectionClosedEvent))
	}
	js.connection.SetReconnectHandler(js.handleReconnectEvent)
}

// initJSContext creates a context which can be used to communicate with JetStream.
func (js *JetStream) initJSContext() error {
	ctx, err := js.connection.JetStream()
	if err != nil {
		return pkgerrors.MakeError(ErrContext, err)
	}
	js.jsCtx = ctx
	return nil
}

// initCloudEventClient creates a client at the CloudEvent protocol level and saves it as JetStream.client.
func (js *JetStream) initCloudEventClient(config backendnats.Config) error {
	if js.ceClient != nil {
		return nil
	}
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	client, err := js.ceClientFactory.NewHTTP(cev2.WithRoundTripper(transport))
	if err != nil {
		return pkgerrors.MakeError(ErrCEClient, err)
	}
	js.ceClient = client
	return nil
}

// ensureStreamExists creates the JetStream stream if it does not exist yet.
func (js *JetStream) ensureStreamExists() (*nats.StreamInfo, *nats.StreamConfig, error) {
	config := js.config
	streamConfig, err := convertNatsConfigToStreamConfig(config)
	if err != nil {
		return nil, streamConfig, pkgerrors.MakeError(ErrConfig, err)
	}

	// Create the stream if it does exist yet and get the stream config.
	info, err := js.jsCtx.StreamInfo(config.JSStreamName)
	if errors.Is(err, nats.ErrStreamNotFound) {
		info, err = js.jsCtx.AddStream(streamConfig)
		if err != nil {
			return info, streamConfig, pkgerrors.MakeError(ErrAddStream, err)
		}
		js.namedLogger().Infow("Stream not found, created a new Stream",
			"stream-info", info)
		return info, streamConfig, nil
	}
	if err != nil {
		return info, streamConfig, pkgerrors.MakeError(ErrUnknown, err)
	}
	js.namedLogger().Infow("Reusing existing Stream", "stream-info", info)

	return info, streamConfig, nil
}

// ensureCorrectStreamConfiguration updates the config of the stream if the supplied streamConfig
// is different from the actual config of the stream (stored in nats.StreamInfo).
func (js *JetStream) ensureCorrectStreamConfiguration(streamInfo *nats.StreamInfo,
	streamConfig *nats.StreamConfig) error {
	if !streamIsConfiguredCorrectly(streamInfo.Config, *streamConfig) {
		newInfo, err := js.jsCtx.UpdateStream(streamConfig)
		if err != nil {
			return pkgerrors.MakeError(ErrUpdateStreamConfig, err)
		}
		js.namedLogger().Infow("Updated existing Stream:", "stream-streamInfo", newInfo)
		return nil
	}
	return nil
}

func streamIsConfiguredCorrectly(got nats.StreamConfig, want nats.StreamConfig) bool {
	return reflect.DeepEqual(got, want)
}

// handleReconnectEvent ensures that the stream exists and is configured correctly after
// a reconnection to the NATS server was established.
func (js *JetStream) handleReconnectEvent(_ *nats.Conn) {
	js.namedLogger().Infow("Called reconnect handler for JetStream")
	if streamInfo, streamConfig, err := js.ensureStreamExists(); err != nil {
		js.namedLogger().Errorw("Failed to ensure the stream exists", "error", err)
		// Try to update the stream configuration (if necessary).
		if err := js.ensureCorrectStreamConfiguration(streamInfo, streamConfig); err != nil {
			js.namedLogger().Errorw("Failed to ensure the stream configuration", "error", err)
			return
		}
		return
	}
}

func (js *JetStream) SyncSubscription(subscription *eventingv1alpha2.Subscription) error {
	subKeyPrefix := createKeyPrefix(subscription)
	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	if err := js.syncSubscriptionEventTypes(subscription); err != nil {
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
		jsSubject := js.GetJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
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
		result = append(result, js.GetJetStreamSubject(source, subject, typeMatching))
	}
	return result
}

// GetJetStreamContext returns the current JetStreamContext.
func (js *JetStream) GetJetStreamContext() nats.JetStreamContext {
	return js.jsCtx
}

// GetJetStreamSubject appends the prefix and the cleaned source to subject.
func (js *JetStream) GetJetStreamSubject(source, subject string, typeMatching eventingv1alpha2.TypeMatching) string {
	if typeMatching == eventingv1alpha2.TypeMatchingExact {
		return fmt.Sprintf("%s.%s", js.GetConfig().JSSubjectPrefix, subject)
	}
	cleanSource, _ := js.cleaner.CleanSource(source)
	return fmt.Sprintf("%s.%s.%s", js.GetConfig().JSSubjectPrefix, cleanSource, subject)
}

// DeleteInvalidConsumers deletes all JetStream consumers having no subscription event types in subscription resources.
func (js *JetStream) DeleteInvalidConsumers(subscriptions []eventingv1alpha2.Subscription) error {
	config := js.config
	consumers := js.jsCtx.Consumers(config.JSStreamName)
	for con := range consumers {
		// consumer should have no interest and no subscription types to delete it
		if !con.PushBound && !js.isConsumerUsedByKymaSub(con.Name, subscriptions) {
			if err := js.deleteConsumerFromJetStream(con.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (js *JetStream) isConsumerUsedByKymaSub(consumerName string, subscriptions []eventingv1alpha2.Subscription) bool {
	if len(subscriptions) == 0 {
		return false
	}
	for ix := range subscriptions {
		cleanedTypes := GetCleanEventTypes(&subscriptions[ix], js.cleaner)
		jsSubjects := js.GetJetStreamSubjects(
			subscriptions[ix].Spec.Source,
			GetCleanEventTypesFromEventTypes(cleanedTypes),
			subscriptions[ix].Spec.TypeMatching)

		for _, jsSubject := range jsSubjects {
			computedConsumerNameFromSubject := computeConsumerName(&subscriptions[ix], jsSubject)
			if consumerName == computedConsumerNameFromSubject {
				return true
			}
		}
	}
	return false
}

// getJetStreamSubject appends the prefix and the cleaned source to subject.
func (js *JetStream) getJetStreamSubject(source, subject string, typeMatching eventingv1alpha2.TypeMatching) string {
	if typeMatching == eventingv1alpha2.TypeMatchingExact {
		return fmt.Sprintf("%s.%s", js.GetConfig().JSSubjectPrefix, subject)
	}
	cleanSource, _ := js.cleaner.CleanSource(source)
	return fmt.Sprintf("%s.%s.%s", js.GetConfig().JSSubjectPrefix, cleanSource, subject)
}

// checkJetStreamConnection reconnects to the server if the server is not connected.
func (js *JetStream) checkJetStreamConnection() error {
	if !js.connection.IsConnected() {
		if err := js.Initialize(js.handleConnectionClosedEvent); err != nil {
			return err
		}
	}
	return nil
}

// syncSubscriptionEventTypes syncs the Kyma subscription types with NATS subscriptions.
func (js *JetStream) syncSubscriptionEventTypes(subscription *eventingv1alpha2.Subscription) error {
	for key, jsSub := range js.subscriptions {
		if isJsSubAssociatedWithKymaSub(key, subscription) {
			err := js.syncSubscriptionEventType(key, subscription, jsSub)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// syncSubscriptionEventType syncs controller runtime subscriptions to subscription CR event types and to JetStream
// subscriptions/consumers.
func (js *JetStream) syncSubscriptionEventType(key SubscriptionSubjectIdentifier,
	subscription *eventingv1alpha2.Subscription, subscriber Subscriber) error {
	// don't try to delete invalid subscriber and its consumer if subscriber has type in subscription CR it belongs to.
	// This means that it will be bound to the existing JetStream consumer in later steps.
	if !subscriber.IsValid() && js.runtimeSubscriptionExistsInKymaSub(key, subscription) {
		return nil
	}

	log := backendutilsv2.LoggerWithSubscription(js.namedLogger(), subscription)
	return js.cleanupUnnecessaryJetStreamSubscribers(subscriber, subscription, log, key)
}

func (js *JetStream) cleanupUnnecessaryJetStreamSubscribers(
	jsSub Subscriber,
	subscription *eventingv1alpha2.Subscription,
	log *zap.SugaredLogger,
	key SubscriptionSubjectIdentifier) error {
	config := js.config
	consumer, err := js.jsCtx.ConsumerInfo(config.JSStreamName, key.ConsumerName())
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
			"Deleting JetStream subscription because it was deleted from subscription types",
			"jetStreamSubject", consumer.Config.FilterSubject,
		)
		return js.deleteSubscriptionFromJetStream(jsSub, key)
	}
	return nil
}

// runtimeSubscriptionExistsInKymaSub returns true if runtime subscriber subject exists in subscription CR.
func (js *JetStream) runtimeSubscriptionExistsInKymaSub(runtimeSubscriptionKey SubscriptionSubjectIdentifier,
	subscription *eventingv1alpha2.Subscription) bool {
	for _, subject := range subscription.Status.Types {
		jsSubject := js.getJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
		jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
		if runtimeSubscriptionKey.consumerName == jsSubKey.consumerName {
			return true
		}
	}
	return false
}

// consumerSubjectExistsInKymaSub checks if the specified consumer is used by the subscription.
func (js *JetStream) consumerSubjectExistsInKymaSub(consumer *nats.ConsumerInfo,
	subscription *eventingv1alpha2.Subscription) bool {
	return utils.ContainsString(
		js.GetJetStreamSubjects(
			subscription.Spec.Source,
			GetCleanEventTypesFromEventTypes(subscription.Status.Types),
			subscription.Spec.TypeMatching),
		consumer.Config.FilterSubject,
	)
}

// deleteSubscriptionFromJetStream deletes subscription from NATS server and from in-memory db.
func (js *JetStream) deleteSubscriptionFromJetStream(jsSub Subscriber, jsSubKey SubscriptionSubjectIdentifier) error {
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

	delete(js.subscriptions, jsSubKey)
	return nil
}

func (js *JetStream) revertEventTypeToOriginal(event *cev2.Event, sugaredLogger *zap.SugaredLogger) {
	// check if original type header exists in the cloud event.
	if orgType, ok := event.Extensions()[originalTypeHeaderName]; ok && orgType != "" {
		event.SetType(fmt.Sprintf("%v", orgType))
		sugaredLogger.Debugf("type reverted to original type using %s header", originalTypeHeaderName)
		return
	}

	// otherwise, manually trim the prefixes from event type.
	ceType := strings.TrimPrefix(event.Type(), fmt.Sprintf("%s.%s.", js.config.JSSubjectPrefix, event.Source()))
	event.SetType(ceType)
	sugaredLogger.Debugw("type reverted to original type by trimming prefixes")
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

		// revert the event type to original form
		js.revertEventTypeToOriginal(ce, ceLogger)

		ceLogger.Debugw("Sending the CloudEvent")

		// dispatch the event to sink
		result := js.ceClient.Send(traceCtxWithCE, *ce)
		if !cev2protocol.IsACK(result) {
			js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusInternalServerError)
			ceLogger.Errorw("Failed to dispatch the CloudEvent", "error", result.Error())
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
	config := js.config
	if err := js.jsCtx.DeleteConsumer(config.JSStreamName, name); err != nil &&
		!errors.Is(err, nats.ErrConsumerNotFound) {
		// if it is not a Not Found error, then return error
		return utils.MakeConsumerError(ErrDeleteConsumer, err, name)
	}

	return nil
}

// syncConsumerAndSubscription makes sure there is a consumer and subscription created on the NATS Backend.
// these also must be bound to each other to ensure that NATS JetStream eventing logic works as expected.
func (js *JetStream) syncConsumerAndSubscription(subscription *eventingv1alpha2.Subscription,
	asyncCallback func(m *nats.Msg)) error {
	for _, eventType := range subscription.Status.Types {
		jsSubject := js.GetJetStreamSubject(subscription.Spec.Source, eventType.CleanType, subscription.Spec.TypeMatching)
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
			return pkgerrors.MakeError(ErrMissingSubscription, err)
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
	jsSubject := js.GetJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
	config := js.config

	consumerInfo, err := js.jsCtx.ConsumerInfo(config.JSStreamName, jsSubKey.ConsumerName())
	if err != nil {
		if errors.Is(err, nats.ErrConsumerNotFound) {
			consumerInfo, err = js.jsCtx.AddConsumer(
				config.JSStreamName,
				js.getConsumerConfig(jsSubKey, jsSubject, subscription.GetMaxInFlightMessages(&js.subsConfig)),
			)
			if err != nil {
				return nil, pkgerrors.MakeError(ErrAddConsumer, err)
			}
		} else {
			return nil, pkgerrors.MakeError(ErrGetConsumer, err)
		}
	}
	return consumerInfo, nil
}

// createNATSSubscription creates a NATS Subscription and binds it to the already existing consumer.
func (js *JetStream) createNATSSubscription(subscription *eventingv1alpha2.Subscription,
	subject eventingv1alpha2.EventType, asyncCallback func(m *nats.Msg)) error {
	jsSubject := js.GetJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)

	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		js.getDefaultSubscriptionOptions(jsSubKey, subscription.GetMaxInFlightMessages(&js.subsConfig))...,
	)
	if err != nil {
		return pkgerrors.MakeError(ErrFailedSubscribe, err)
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
	jsSubject := js.GetJetStreamSubject(subscription.Spec.Source, subject.CleanType, subscription.Spec.TypeMatching)
	config := js.config
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
	// bind the existing consumer to a new subscription on JetStream
	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		nats.Bind(config.JSStreamName, jsSubKey.ConsumerName()),
	)
	if err != nil {
		return pkgerrors.MakeError(ErrFailedSubscribe, err)
	}
	// save recreated JetStream subscription in storage
	js.subscriptions[jsSubKey] = &Subscription{Subscription: jsSubscription}
	return nil
}

// syncConsumerMaxInFlight checks that the latest Subscription's maxInFlight value
// is propagated to the NATS consumer as MaxAckPending.
func (js *JetStream) syncConsumerMaxInFlight(subscription *eventingv1alpha2.Subscription,
	consumerInfo nats.ConsumerInfo) error {
	config := js.config
	maxInFlight := subscription.GetMaxInFlightMessages(&js.subsConfig)

	if consumerInfo.Config.MaxAckPending == maxInFlight {
		return nil
	}

	// set the new maxInFlight value
	consumerConfig := consumerInfo.Config
	consumerConfig.MaxAckPending = maxInFlight

	// update the consumer
	if _, updateErr := js.jsCtx.UpdateConsumer(config.JSStreamName, &consumerConfig); updateErr != nil {
		return pkgerrors.MakeError(ErrUpdateConsumer, updateErr)
	}
	return nil
}

// GetNATSSubscriptions returns the map which contains details of all NATS subscriptions and consumers.
// Use this only for testing purposes.
func (js *JetStream) GetNATSSubscriptions() map[SubscriptionSubjectIdentifier]Subscriber {
	return js.subscriptions
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
