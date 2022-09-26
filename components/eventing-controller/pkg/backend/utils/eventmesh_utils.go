package utils

import (
	"crypto/sha1" //nolint:gosec
	"fmt"
	"strings"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
)

// bebSubscriptionNameMapper maps a Kyma subscription to an ID that can be used on the BEB backend,
// which has a max length. Domain name is used to make the names on BEB unique.
type eventMeshSubscriptionNameMapper struct {
	domainName string
	maxLength  int
}

func NewBEBSubscriptionNameMapper(domainName string, maxNameLength int) NameMapper {
	return &eventMeshSubscriptionNameMapper{
		domainName: domainName,
		maxLength:  maxNameLength,
	}
}

func (m *eventMeshSubscriptionNameMapper) MapSubscriptionName(subscriptionName, subscriptionNamespace string) string {
	hash := hashSubscriptionFullName(m.domainName, subscriptionNamespace, subscriptionName)
	return shortenNameAndAppendHash(subscriptionName, hash, m.maxLength)
}

// produces a name+hash which is not longer than maxLength. If necessary, shortens name, not the hash.
// Requires maxLength >= len(hash).
func shortenNameAndAppendHash(name string, hash string, maxLength int) string {
	if len(hash) > maxLength {
		// This shouldn't happen!
		panic(fmt.Sprintf("max name length (%d) used for EventMesh subscription mapper"+
			" is not large enough to hold the hash (%s)", maxLength, hash))
	}
	maxNameLen := maxLength - len(hash)
	// keep the first maxNameLen characters of the name
	if maxNameLen <= 0 {
		return hash
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	return name + hash
}

func GetHash(subscription *types.Subscription) (int64, error) {
	hash, err := hashstructure.Hash(subscription, hashstructure.FormatV2, nil)
	if err != nil {
		return 0, err
	}
	return int64(hash), nil
}

func hashSubscriptionFullName(domainName, namespace, name string) string {
	hash := sha1.Sum([]byte(domainName + namespace + name)) //nolint:gosec
	return fmt.Sprintf("%x", hash)
}

func getDefaultSubscriptionV1Alpha2(protocolSettings *eventingv1alpha2.ProtocolSettings) (*types.Subscription, error) {
	emsSubscription := &types.Subscription{}
	emsSubscription.ContentMode = *protocolSettings.ContentMode
	emsSubscription.ExemptHandshake = *protocolSettings.ExemptHandshake
	qos, err := getQos(*protocolSettings.Qos)
	if err != nil {
		return nil, err
	}
	emsSubscription.Qos = qos
	return emsSubscription, nil
}

func getQos(qosStr string) (types.Qos, error) {
	qosStr = strings.ReplaceAll(qosStr, "-", "_")
	switch qosStr {
	case string(types.QosAtLeastOnce):
		return types.QosAtLeastOnce, nil
	case string(types.QosAtMostOnce):
		return types.QosAtMostOnce, nil
	default:
		return "", fmt.Errorf("invalid Qos: %s", qosStr)
	}
}

func ConvertKymaSubToEventMeshSub(subscription *eventingv1alpha2.Subscription, typeInfos []EventTypeInfo,
	apiRule *apigatewayv1beta1.APIRule, defaultWebhookAuth *types.WebhookAuth,
	defaultProtocolSettings *eventingv1alpha2.ProtocolSettings,
	defaultNamespace string, nameMapper NameMapper) (*types.Subscription, error) { //nolint:gocognit

	// get default EventMesh subscription object
	eventMeshSubscription, err := getDefaultSubscriptionV1Alpha2(defaultProtocolSettings)
	if err != nil {
		return nil, errors.Wrap(err, "apply default protocol settings failed")
	}
	// set Name of EventMesh subscription
	eventMeshSubscription.Name = nameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace)

	// @TODO: Check how the protocol settings would work in new CRD
	//// Applying protocol settings if provided in subscription CR
	// if subscription.Spec.ProtocolSettings != nil {
	//	if subscription.Spec.ProtocolSettings.ContentMode != nil {
	//		eventMeshSubscription.ContentMode = *subscription.Spec.ProtocolSettings.ContentMode
	//	}
	//	// ExemptHandshake
	//	if subscription.Spec.ProtocolSettings.ExemptHandshake != nil {
	//		eventMeshSubscription.ExemptHandshake = *subscription.Spec.ProtocolSettings.ExemptHandshake
	//	}
	//	// Qos
	//	if subscription.Spec.ProtocolSettings.Qos != nil {
	//		qos, err := getQos(*subscription.Spec.ProtocolSettings.Qos)
	//		if err != nil {
	//			return nil, err
	//		}
	//		eventMeshSubscription.Qos = qos
	//	}
	// }

	// Events
	// set the event types in EventMesh subscription instance

	eventMeshNamespace := defaultNamespace
	if subscription.Spec.TypeMatching == eventingv1alpha2.EXACT {
		eventMeshNamespace = subscription.Spec.Source
	}

	for _, typeInfo := range typeInfos {
		eventType := typeInfo.ProcessedType
		if subscription.Spec.TypeMatching == eventingv1alpha2.EXACT {
			eventType = typeInfo.OriginalType
		}
		eventMeshSubscription.Events = append(
			eventMeshSubscription.Events,
			types.Event{Source: eventMeshNamespace, Type: eventType},
		)
	}

	// WebhookURL
	// set WebhookURL of EventMesh subscription where the events will be dispatched to.
	urlTobeRegistered, err := getExposedURLFromAPIRule(apiRule, subscription.Spec.Sink)
	if err != nil {
		return nil, errors.Wrap(err, "get APIRule exposed URL failed")
	}
	eventMeshSubscription.WebhookURL = urlTobeRegistered

	// Using default webhook auth unless specified in Subscription CR
	auth := defaultWebhookAuth
	if subscription.Spec.ProtocolSettings != nil && subscription.Spec.ProtocolSettings.WebhookAuth != nil {
		auth = &types.WebhookAuth{}
		auth.ClientID = subscription.Spec.ProtocolSettings.WebhookAuth.ClientID
		auth.ClientSecret = subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret
		if subscription.Spec.ProtocolSettings.WebhookAuth.GrantType == string(types.GrantTypeClientCredentials) {
			auth.GrantType = types.GrantTypeClientCredentials
		} else {
			return nil, fmt.Errorf("invalid GrantType: %v", subscription.Spec.ProtocolSettings.WebhookAuth.GrantType)
		}
		if subscription.Spec.ProtocolSettings.WebhookAuth.Type == string(types.AuthTypeClientCredentials) {
			auth.Type = types.AuthTypeClientCredentials
		} else {
			return nil, fmt.Errorf("invalid Type: %v", subscription.Spec.ProtocolSettings.WebhookAuth.Type)
		}
		auth.TokenURL = subscription.Spec.ProtocolSettings.WebhookAuth.TokenURL
	}
	eventMeshSubscription.WebhookAuth = auth
	return eventMeshSubscription, nil
}

func GetCleanedEventMeshSubscription(subscription *types.Subscription) *types.Subscription {
	eventMeshSubscription := &types.Subscription{}

	// Name
	eventMeshSubscription.Name = subscription.Name
	eventMeshSubscription.ContentMode = subscription.ContentMode
	eventMeshSubscription.ExemptHandshake = subscription.ExemptHandshake

	// Qos
	eventMeshSubscription.Qos = subscription.Qos

	// WebhookURL
	eventMeshSubscription.WebhookURL = subscription.WebhookURL

	// Events
	for _, e := range subscription.Events {
		s := e.Source
		t := e.Type
		eventMeshSubscription.Events = append(eventMeshSubscription.Events, types.Event{Source: s, Type: t})
	}

	return eventMeshSubscription
}

func IsEventMeshSubModified(subscription *types.Subscription, hash int64) (bool, error) {
	// generate has of new subscription
	newHash, err := GetHash(subscription)
	if err != nil {
		return false, err
	}

	// compare hashes
	return newHash != hash, nil
}
