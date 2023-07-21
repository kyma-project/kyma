package jetstream

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	"github.com/pkg/errors"

	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"
	"github.com/nats-io/nats.go"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
)

const (
	StorageTypeMemory = "memory"
	StorageTypeFile   = "file"

	RetentionPolicyLimits   = "limits"
	RetentionPolicyInterest = "interest"

	DiscardPolicyNew = "new"
	DiscardPolicyOld = "old"

	ConsumerDeliverPolicyAll            = "all"
	ConsumerDeliverPolicyLast           = "last"
	ConsumerDeliverPolicyLastPerSubject = "last_per_subject"
	ConsumerDeliverPolicyNew            = "new"
)

// getDefaultSubscriptionOptions builds the default nats.SubOpts by using the subscription/consumer configuration.
func (js *JetStream) getDefaultSubscriptionOptions(consumer SubscriptionSubjectIdentifier,
	maxInFlightMessages int) DefaultSubOpts {
	return DefaultSubOpts{
		nats.Durable(consumer.consumerName),
		nats.Description(consumer.namespacedSubjectName),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.IdleHeartbeat(idleHeartBeatDuration),
		nats.EnableFlowControl(),
		toJetStreamConsumerDeliverPolicyOptOrDefault(js.Config.JSConsumerDeliverPolicy),
		nats.MaxAckPending(maxInFlightMessages),
		nats.MaxDeliver(jsConsumerMaxRedeliver),
		nats.AckWait(jsConsumerAckWait),
		nats.Bind(js.Config.JSStreamName, consumer.ConsumerName()),
	}
}

var ErrInvalidStorageType = pkgerrors.NewArgumentError("invalid stream storage type: %q")
var ErrInvalidRetentionPolicy = pkgerrors.NewArgumentError("invalid stream retention policy: %q")
var ErrInvalidDiscardPolicy = pkgerrors.NewArgumentError("invalid stream discard policy: %q")

// toJetStreamStorageType converts a string to a nats.StorageType.
func toJetStreamStorageType(s string) (nats.StorageType, error) {
	switch s {
	case StorageTypeMemory:
		return nats.MemoryStorage, nil
	case StorageTypeFile:
		return nats.FileStorage, nil
	}
	return nats.MemoryStorage, ErrInvalidStorageType.WithArg(s)
}

func toJetStreamRetentionPolicy(s string) (nats.RetentionPolicy, error) {
	switch s {
	case RetentionPolicyLimits:
		return nats.LimitsPolicy, nil
	case RetentionPolicyInterest:
		return nats.InterestPolicy, nil
	}
	return nats.LimitsPolicy, ErrInvalidRetentionPolicy.WithArg(s)
}

func toJetStreamDiscardPolicy(s string) (nats.DiscardPolicy, error) {
	switch s {
	case DiscardPolicyOld:
		return nats.DiscardOld, nil
	case DiscardPolicyNew:
		return nats.DiscardNew, nil
	}
	return nats.DiscardNew, ErrInvalidDiscardPolicy.WithArg(s)
}

// toJetStreamConsumerDeliverPolicyOpt returns a nats.DeliverPolicy opt based on the given deliver policy string value.
// It returns "DeliverNew" as the default nats.DeliverPolicy opt, if the given deliver policy value is not supported.
// Supported deliver policy values are ("all", "last", "last_per_subject" and "new").
func toJetStreamConsumerDeliverPolicyOptOrDefault(deliverPolicy string) nats.SubOpt {
	switch deliverPolicy {
	case ConsumerDeliverPolicyAll:
		return nats.DeliverAll()
	case ConsumerDeliverPolicyLast:
		return nats.DeliverLast()
	case ConsumerDeliverPolicyLastPerSubject:
		return nats.DeliverLastPerSubject()
	case ConsumerDeliverPolicyNew:
		return nats.DeliverNew()
	}
	return nats.DeliverNew()
}

// toJetStreamConsumerDeliverPolicy returns a nats.DeliverPolicy based on the given deliver policy string value.
// It returns "DeliverNew" as the default nats.DeliverPolicy, if the given deliver policy value is not supported.
// Supported deliver policy values are ("all", "last", "last_per_subject" and "new").
func toJetStreamConsumerDeliverPolicy(deliverPolicy string) nats.DeliverPolicy {
	switch deliverPolicy {
	case ConsumerDeliverPolicyAll:
		return nats.DeliverAllPolicy
	case ConsumerDeliverPolicyLast:
		return nats.DeliverLastPolicy
	case ConsumerDeliverPolicyLastPerSubject:
		return nats.DeliverLastPerSubjectPolicy
	case ConsumerDeliverPolicyNew:
		return nats.DeliverNewPolicy
	}
	return nats.DeliverNewPolicy
}

func getStreamConfig(natsConfig env.NATSConfig) (*nats.StreamConfig, error) {
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
		Name:              natsConfig.JSStreamName,
		Storage:           storage,
		Replicas:          natsConfig.JSStreamReplicas,
		Retention:         retentionPolicy,
		MaxMsgs:           natsConfig.JSStreamMaxMessages,
		MaxBytes:          maxBytes.Value(),
		Discard:           discardPolicy,
		MaxMsgsPerSubject: natsConfig.JSStreamMaxMsgsPerTopic,
		// Since one stream is used to store events of all types, the stream has to match all event types, and therefore
		// we use the wildcard char >. However, to avoid matching internal JetStream and non-Kyma-related subjects, we
		// use a prefix. This prefix is handled only on the JetStream level (i.e. JetStream handler
		// and EPP) and should not be exposed in the Kyma subscription. Any Kyma event type gets appended with the
		// configured stream's subject prefix.
		Subjects: []string{fmt.Sprintf("%s.>", natsConfig.JSSubjectPrefix)},
	}
	return streamConfig, nil
}

// getConsumerConfig return the consumerConfig according to the default configuration.
func (js *JetStream) getConsumerConfig(jsSubKey SubscriptionSubjectIdentifier,
	jsSubject string, maxInFlight int) *nats.ConsumerConfig {
	return &nats.ConsumerConfig{
		Durable:        jsSubKey.ConsumerName(),
		Description:    jsSubKey.namespacedSubjectName,
		DeliverPolicy:  toJetStreamConsumerDeliverPolicy(js.Config.JSConsumerDeliverPolicy),
		FlowControl:    true,
		MaxAckPending:  maxInFlight,
		AckPolicy:      nats.AckExplicitPolicy,
		AckWait:        jsConsumerAckWait,
		MaxDeliver:     jsConsumerMaxRedeliver,
		FilterSubject:  jsSubject,
		ReplayPolicy:   nats.ReplayInstantPolicy,
		DeliverSubject: nats.NewInbox(),
		Heartbeat:      idleHeartBeatDuration,
	}
}

func createKeyPrefix(sub *eventingv1alpha2.Subscription) string {
	namespacedName := types.NamespacedName{
		Namespace: sub.Namespace,
		Name:      sub.Name,
	}
	return namespacedName.String()
}

func GetCleanEventTypesFromEventTypes(eventTypes []eventingv1alpha2.EventType) []string {
	var cleantypes []string
	for _, eventtypes := range eventTypes {
		cleantypes = append(cleantypes, eventtypes.CleanType)
	}
	return cleantypes
}

// TODO: to be moved to subscription types
func getUniqueEventTypes(eventTypes []string) []string {
	unique := make([]string, 0, len(eventTypes))
	mapper := make(map[string]bool)

	for _, val := range eventTypes {
		if _, ok := mapper[val]; !ok {
			mapper[val] = true
			unique = append(unique, val)
		}
	}

	return unique
}

// GetCleanEventTypes returns a list of clean eventTypes from the unique types in the subscription.
func GetCleanEventTypes(sub *eventingv1alpha2.Subscription, cleaner cleaner.Cleaner) []eventingv1alpha2.EventType {
	uniqueTypes := getUniqueEventTypes(sub.Spec.Types)
	var cleanEventTypes []eventingv1alpha2.EventType
	for _, eventType := range uniqueTypes {
		cleanType := eventType
		if sub.Spec.TypeMatching != eventingv1alpha2.TypeMatchingExact {
			cleanType, _ = cleaner.CleanEventType(eventType)
		}
		newEventType := eventingv1alpha2.EventType{
			OriginalType: eventType,
			CleanType:    cleanType,
		}
		cleanEventTypes = append(cleanEventTypes, newEventType)
	}
	return cleanEventTypes
}

// GetBackendJetStreamTypes gets the original event type and the consumer name for all the subscriptions
// and this slice is set as the backend specific status for JetStream.
func GetBackendJetStreamTypes(subscription *eventingv1alpha2.Subscription,
	jsSubjects []string) ([]eventingv1alpha2.JetStreamTypes, error) {
	if len(jsSubjects) != len(subscription.Spec.Types) {
		return nil, errors.New("length of JetStream subjects do not match with eventTypes from spec")
	}

	var jsTypes []eventingv1alpha2.JetStreamTypes
	for i, ot := range subscription.Spec.Types {
		jt := eventingv1alpha2.JetStreamTypes{OriginalType: ot,
			ConsumerName: computeConsumerName(subscription, jsSubjects[i])}
		jsTypes = append(jsTypes, jt)
	}
	return jsTypes, nil
}

// isJsSubAssociatedWithKymaSub returns true if the given SubscriptionSubjectIdentifier and Kyma subscription
// have the same namespaced name, otherwise returns false.
func isJsSubAssociatedWithKymaSub(jsSubKey SubscriptionSubjectIdentifier,
	subscription *eventingv1alpha2.Subscription) bool {
	return createKeyPrefix(subscription) == jsSubKey.NamespacedName()
}

// ----------------------------------------
// SubscriptionSubjectIdentifier utils
// ----------------------------------------

// NamespacedName returns the Kubernetes namespaced name.
func (s SubscriptionSubjectIdentifier) NamespacedName() string {
	return s.namespacedSubjectName[:strings.LastIndex(s.namespacedSubjectName, separator)]
}

// ConsumerName returns the JetStream consumer name.
func (s SubscriptionSubjectIdentifier) ConsumerName() string {
	return s.consumerName
}

// NewSubscriptionSubjectIdentifier returns a new SubscriptionSubjectIdentifier instance.
func NewSubscriptionSubjectIdentifier(subscription *eventingv1alpha2.Subscription,
	subject string) SubscriptionSubjectIdentifier {
	cn := computeConsumerName(subscription, subject)          // compute the consumer name once
	nn := computeNamespacedSubjectName(subscription, subject) // compute the namespaced name with the subject once
	return SubscriptionSubjectIdentifier{consumerName: cn, namespacedSubjectName: nn}
}

// computeConsumerName returns JetStream consumer name of the given subscription and subject.
// It uses the crypto/md5 lib to return a string of 32 characters as recommended by the JetStream
// documentation https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming.
func computeConsumerName(subscription *eventingv1alpha2.Subscription, subject string) string {
	cn := subscription.Namespace + separator + subscription.Name + separator + subject
	h := md5.Sum([]byte(cn)) // #nosec
	return hex.EncodeToString(h[:])
}

// computeNamespacedSubjectName returns Kubernetes namespaced name of the given subscription along with the subject.
func computeNamespacedSubjectName(subscription *eventingv1alpha2.Subscription, subject string) string {
	return subscription.Namespace + separator + subscription.Name + separator + subject
}
