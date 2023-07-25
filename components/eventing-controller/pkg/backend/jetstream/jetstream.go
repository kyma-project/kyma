package jetstream

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	http2 "github.com/cloudevents/sdk-go/v2/protocol/http"

	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

var _ Backend = &JetStream{}

const (
	jsHandlerName          = "jetstream-handler"
	jsMaxStreamNameLength  = 32
	idleHeartBeatDuration  = 1 * time.Minute
	jsConsumerMaxRedeliver = 100
	jsConsumerNakDelay     = 30 * time.Second
	jsConsumerAckWait      = 30 * time.Second
	originalTypeHeaderName = "originaltype"
)

func NewJetStream(config env.NATSConfig, metricsCollector *backendmetrics.Collector,
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

func (js *JetStream) Initialize(connCloseHandler backendutils.ConnClosedHandler) error {
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

func (js *JetStream) DeleteSubscriptionOnly(subscription *eventingv1alpha2.Subscription) error {
	js.namedLogger().Infow(
		"Delete JetStream subscription only",
		"namespace", subscription.Namespace,
		"name", subscription.Name,
	)

	if err := js.checkJetStreamConnection(); err != nil {
		return err
	}

	for key, jsSub := range js.subscriptions {
		if !isJsSubAssociatedWithKymaSub(key, subscription) {
			continue
		}
		if err := js.deleteSubscriptionFromJetStreamOnly(jsSub, key); err != nil {
			return err
		}
	}

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
		return fmt.Sprintf("%s.%s", js.Config.JSSubjectPrefix, subject)
	}
	cleanSource, _ := js.cleaner.CleanSource(source)
	return fmt.Sprintf("%s.%s.%s", js.Config.JSSubjectPrefix, cleanSource, subject)
}

// DeleteInvalidConsumers deletes all JetStream consumers having no subscription event types in subscription resources.
func (js *JetStream) DeleteInvalidConsumers(subscriptions []eventingv1alpha2.Subscription) error {
	consumers := js.jsCtx.Consumers(js.Config.JSStreamName)
	for con := range consumers {
		// consumer should have no interest and no subscription types to delete it
		if !con.PushBound && !js.isConsumerUsedByKymaSub(con.Name, subscriptions) {
			if err := js.deleteConsumerFromJetStream(con.Name); err != nil {
				return err
			}
			js.namedLogger().Infow("Dangling JetStream consumer is deleted", "name", con.Name,
				"description", con.Config.Description)
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
		return fmt.Sprintf("%s.%s", js.Config.JSSubjectPrefix, subject)
	}
	cleanSource, _ := js.cleaner.CleanSource(source)
	return fmt.Sprintf("%s.%s.%s", js.Config.JSSubjectPrefix, cleanSource, subject)
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
	if _, err := toJetStreamDiscardPolicy(js.Config.JSStreamDiscardPolicy); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) initNATSConn(connCloseHandler backendutils.ConnClosedHandler) error {
	if js.Conn == nil || js.Conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.Config.MaxReconnects),
			nats.ReconnectWait(js.Config.ReconnectWait),
			nats.Name("Kyma Controller"),
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
	if err := js.ensureStreamExistsAndIsConfiguredCorrectly(); err != nil {
		js.namedLogger().Errorw("Failed to ensure the stream exists", "error", err)
	}
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
	// only comparing the fields which we define in stream config.
	if got.Name != want.Name ||
		got.Storage != want.Storage ||
		got.Replicas != want.Replicas ||
		got.Retention != want.Retention ||
		got.MaxMsgs != want.MaxMsgs ||
		got.MaxBytes != want.MaxBytes ||
		got.Discard != want.Discard {
		return false
	}
	return reflect.DeepEqual(got.Subjects, want.Subjects)
}

func (js *JetStream) initJSContext() error {
	jsCtx, err := js.Conn.JetStream()
	if err != nil {
		return fmt.Errorf("failed to create the JetStream context: %w", err)
	}
	js.jsCtx = jsCtx
	return nil
}

func (js *JetStream) initCloudEventClient(config env.NATSConfig) error {
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

	log := backendutils.LoggerWithSubscription(js.namedLogger(), subscription)
	return js.cleanupUnnecessaryJetStreamSubscribers(subscriber, subscription, log, key)
}

func (js *JetStream) cleanupUnnecessaryJetStreamSubscribers(
	jsSub Subscriber,
	subscription *eventingv1alpha2.Subscription,
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

// deleteSubscriptionFromJetStreamOnly deletes the subscription from NATS server and from in-memory db.
// Note: The consumer will not be deleted, meaning there should be no message loss.
func (js *JetStream) deleteSubscriptionFromJetStreamOnly(jsSub Subscriber, jsSubKey SubscriptionSubjectIdentifier) error {
	if jsSub.IsValid() {
		// The Unsubscribe function should not delete the consumer because it was added manually.
		if err := jsSub.Unsubscribe(); err != nil {
			return utils.MakeSubscriptionError(ErrFailedUnsubscribe, err, jsSub)
		}
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
	ceType := strings.TrimPrefix(event.Type(), fmt.Sprintf("%s.%s.", js.Config.JSSubjectPrefix, event.Source()))
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
		start := time.Now()
		result := js.client.Send(traceCtxWithCE, *ce)
		duration := time.Since(start)
		var res *http2.Result
		if !cev2protocol.IsACK(result) {
			status := http.StatusInternalServerError
			if cev2.ResultAs(result, &res) {
				status = res.StatusCode
			}

			js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, status)
			js.metricsCollector.RecordLatencyPerSubscription(duration, subscriptionName, ce.Type(), sink, status)

			// NAK the msg with a delay so it is redelivered after jsConsumerNakDelay period.
			if err := msg.NakWithDelay(jsConsumerNakDelay); err != nil {
				js.namedLogger().Errorw("failed to NAK an event on JetStream")
			}

			ceLogger.Errorw("Failed to dispatch the CloudEvent", "error", result.Error())
			return
		}

		// event was successfully dispatched, check if acknowledged by the NATS server
		// if not, the message is redelivered.
		if ackErr := msg.Ack(); ackErr != nil {
			ceLogger.Errorw("Failed to ACK an event on JetStream")
		}

		status := http.StatusOK
		if cev2.ResultAs(result, &res) {
			status = res.StatusCode
		}

		js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, status)
		js.metricsCollector.RecordLatencyPerSubscription(duration, subscriptionName, ce.Type(), sink, status)
		ceLogger.Debugw("CloudEvent was dispatched")
	}
}

// deleteConsumerFromJS deletes consumer on NATS Server.
func (js *JetStream) deleteConsumerFromJetStream(name string) error {
	if err := js.jsCtx.DeleteConsumer(js.Config.JSStreamName, name); err != nil &&
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

	consumerInfo, err := js.jsCtx.ConsumerInfo(js.Config.JSStreamName, jsSubKey.ConsumerName())
	if err != nil {
		if errors.Is(err, nats.ErrConsumerNotFound) {
			consumerInfo, err = js.jsCtx.AddConsumer(
				js.Config.JSStreamName,
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
	jsSubKey := NewSubscriptionSubjectIdentifier(subscription, jsSubject)
	// bind the existing consumer to a new subscription on JetStream
	jsSubscription, err := js.jsCtx.Subscribe(
		jsSubject,
		asyncCallback,
		nats.Bind(js.Config.JSStreamName, jsSubKey.ConsumerName()),
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
	maxInFlight := subscription.GetMaxInFlightMessages(&js.subsConfig)

	if consumerInfo.Config.MaxAckPending == maxInFlight {
		return nil
	}

	// set the new maxInFlight value
	consumerConfig := consumerInfo.Config
	consumerConfig.MaxAckPending = maxInFlight

	// update the consumer
	if _, updateErr := js.jsCtx.UpdateConsumer(js.Config.JSStreamName, &consumerConfig); updateErr != nil {
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
