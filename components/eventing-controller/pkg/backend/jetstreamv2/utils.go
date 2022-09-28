package jetstreamv2

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	StorageTypeMemory = "memory"
	StorageTypeFile   = "file"

	RetentionPolicyLimits   = "limits"
	RetentionPolicyInterest = "interest"

	ConsumerDeliverPolicyAll            = "all"
	ConsumerDeliverPolicyLast           = "last"
	ConsumerDeliverPolicyLastPerSubject = "last_per_subject"
	ConsumerDeliverPolicyNew            = "new"
)

func toJetStreamStorageType(s string) (nats.StorageType, error) {
	switch s {
	case StorageTypeMemory:
		return nats.MemoryStorage, nil
	case StorageTypeFile:
		return nats.FileStorage, nil
	}
	return nats.MemoryStorage, fmt.Errorf("invalid stream storage type %q", s)
}

func toJetStreamRetentionPolicy(s string) (nats.RetentionPolicy, error) {
	switch s {
	case RetentionPolicyLimits:
		return nats.LimitsPolicy, nil
	case RetentionPolicyInterest:
		return nats.InterestPolicy, nil
	}
	return nats.LimitsPolicy, fmt.Errorf("invalid stream retention policy %q", s)
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
		Subjects: []string{fmt.Sprintf("%s.>", env.JetStreamSubjectPrefix)},
	}
	return streamConfig, nil
}

func createKeyPrefix(sub *eventingv1alpha2.Subscription) string {
	namespacedName := types.NamespacedName{
		Namespace: sub.Namespace,
		Name:      sub.Name,
	}
	return namespacedName.String()
}

func getCleanEventTypesFromStatus(subscriptionStatus eventingv1alpha2.SubscriptionStatus) []string {
	var cleantypes []string
	for _, eventtypes := range subscriptionStatus.Types {
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

// getCleanEventTypes returns a list of clean eventTypes from the unique types in the subscription.
func getCleanEventTypes(sub *eventingv1alpha2.Subscription, cleaner cleaner.Cleaner) ([]eventingv1alpha2.EventType, error) {
	// TODO: Put this in the validation webhook
	if sub.Spec.Types == nil || sub.Spec.Source == "" {
		return []eventingv1alpha2.EventType{}, errors.New("event source and types must be provided")
	}

	uniqueTypes := getUniqueEventTypes(sub.Spec.Types)
	var cleanEventTypes []eventingv1alpha2.EventType
	for _, eventType := range uniqueTypes {
		cleanType := eventType
		var err error
		if sub.Spec.TypeMatching != eventingv1alpha2.TypeMatchingExact {
			cleanType, err = getCleanEventType(eventType, cleaner)
			if err != nil {
				return []eventingv1alpha2.EventType{}, err
			}
		}
		newEventType := eventingv1alpha2.EventType{
			OriginalType: eventType,
			CleanType:    cleanType,
		}
		cleanEventTypes = append(cleanEventTypes, newEventType)
	}
	return cleanEventTypes, nil
}

// isJsSubAssociatedWithKymaSub returns true if the given SubscriptionSubjectIdentifier and Kyma subscription
// have the same namespaced name, otherwise returns false.
func isJsSubAssociatedWithKymaSub(jsSubKey SubscriptionSubjectIdentifier, subscription *eventingv1alpha2.Subscription) bool {
	return createKeyPrefix(subscription) == jsSubKey.NamespacedName()
}

//----------------------------------------
// SubscriptionSubjectIdentifier utils
//----------------------------------------

// NamespacedName returns the Kubernetes namespaced name.
func (s SubscriptionSubjectIdentifier) NamespacedName() string {
	return s.namespacedSubjectName[:strings.LastIndex(s.namespacedSubjectName, separator)]
}

// ConsumerName returns the JetStream consumer name.
func (s SubscriptionSubjectIdentifier) ConsumerName() string {
	return s.consumerName
}

// NewSubscriptionSubjectIdentifier returns a new SubscriptionSubjectIdentifier instance.
func NewSubscriptionSubjectIdentifier(subscription *eventingv1alpha2.Subscription, subject string) SubscriptionSubjectIdentifier {
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

//----------------------------------------
// CloudEvent utils
//----------------------------------------

func convertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	if err := event.Validate(); err != nil {
		return nil, err
	}
	return &event, nil
}

//----------------------------------------
// Testing utils
//----------------------------------------

func getCleanEventType(eventType string, cleaner cleaner.Cleaner) (string, error) {
	if len(eventType) == 0 {
		return "", nats.ErrBadSubject
	}
	if segments := strings.Split(eventType, "."); len(segments) < 2 {
		return "", nats.ErrBadSubject
	}
	return cleaner.CleanEventType(eventType)
}
