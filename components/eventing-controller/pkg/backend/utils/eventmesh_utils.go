package utils

import (
	"crypto/sha1" //nolint:gosec
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"

    apigatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/internal/featureflags"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

// eventMeshSubscriptionNameMapper maps a Kyma subscription to an ID that can be used on the EventMesh backend,
// which has a max length. Domain name is used to make the names on EventMesh unique.
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

func GetWebhookAuthHash(webhookAuth *types.WebhookAuth) (int64, error) {
	hash, err := hashstructure.Hash(webhookAuth, hashstructure.FormatV2, nil)
	if err != nil {
		return 0, err
	}
	return int64(hash), nil
}

func hashSubscriptionFullName(domainName, namespace, name string) string {
	hash := sha1.Sum([]byte(domainName + namespace + name)) //nolint:gosec
	return fmt.Sprintf("%x", hash)
}

func getDefaultSubscriptionV1Alpha2(protocolSettings *ProtocolSettings) (*types.Subscription, error) {
	// @TODO: Rename this method to getDefaultSubscription once old BEB backend is depreciated
	emsSubscription := &types.Subscription{}
	emsSubscription.ContentMode = *protocolSettings.ContentMode
	emsSubscription.ExemptHandshake = *protocolSettings.ExemptHandshake
	emsSubscription.Qos = types.GetQos(*protocolSettings.Qos)
	return emsSubscription, nil
}

func getEventMeshEvents(typeInfos []EventTypeInfo, typeMatching eventingv1alpha2.TypeMatching,
	defaultNamespace, source string) types.Events {
	eventMeshNamespace := defaultNamespace
	if typeMatching == eventingv1alpha2.TypeMatchingExact && source != "" {
		eventMeshNamespace = source
	}

	events := make(types.Events, 0, len(typeInfos))
	for _, typeInfo := range typeInfos {
		eventType := typeInfo.ProcessedType
		if typeMatching == eventingv1alpha2.TypeMatchingExact {
			eventType = typeInfo.OriginalType
		}
		events = append(
			events,
			types.Event{Source: eventMeshNamespace, Type: eventType},
		)
	}
	return events
}

func ConvertKymaSubToEventMeshSub(subscription *eventingv1alpha2.Subscription, typeInfos []EventTypeInfo,
	apiRule *apigatewayv1beta1.APIRule, defaultWebhookAuth *types.WebhookAuth,
	defaultProtocolSettings *ProtocolSettings,
	defaultNamespace string, nameMapper NameMapper) (*types.Subscription, error) { //nolint:gocognit
	// get default EventMesh subscription object
	eventMeshSubscription, err := getDefaultSubscriptionV1Alpha2(defaultProtocolSettings)
	if err != nil {
		return nil, errors.Wrap(err, "apply default protocol settings failed")
	}
	// set Name of EventMesh subscription
	eventMeshSubscription.Name = nameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace)

	// Applying protocol settings if provided in subscription CR
	if setErr := setEventMeshProtocolSettings(subscription, eventMeshSubscription); setErr != nil {
		return nil, setErr
	}

	// Events
	// set the event types in EventMesh subscription instance
	eventMeshSubscription.Events = getEventMeshEvents(typeInfos, subscription.Spec.TypeMatching,
		defaultNamespace, subscription.Spec.Source)

	// WebhookURL
	// set WebhookURL of EventMesh subscription where the events will be dispatched to.
	urlTobeRegistered, err := GetExposedURLFromAPIRule(apiRule, subscription.Spec.Sink)
	if err != nil {
		return nil, errors.Wrap(err, "get APIRule exposed URL failed")
	}
	eventMeshSubscription.WebhookURL = urlTobeRegistered

	// set webhook auth
	eventMeshSubscription.WebhookAuth, err = getEventMeshWebhookAuth(subscription, defaultWebhookAuth)
	if err != nil {
		return nil, err
	}

	return eventMeshSubscription, nil
}

func setEventMeshProtocolSettings(subscription *eventingv1alpha2.Subscription, eventMeshSub *types.Subscription) error {
	// Applying protocol settings if provided in subscription CR
	// qos
	if qosStr, ok := subscription.Spec.Config[eventingv1alpha2.ProtocolSettingsQos]; ok {
		eventMeshSub.Qos = types.GetQos(qosStr)
	}
	// content mode
	if contentMode, ok := subscription.Spec.Config[eventingv1alpha2.ProtocolSettingsContentMode]; ok && contentMode != "" {
		eventMeshSub.ContentMode = contentMode
	}
	// handshake
	if exemptHandshake, ok := subscription.Spec.Config[eventingv1alpha2.ProtocolSettingsExemptHandshake]; ok {
		handshake, err := strconv.ParseBool(exemptHandshake)
		if err != nil {
			handshake = true
		}
		eventMeshSub.ExemptHandshake = handshake
	}
	return nil
}

// getEventMeshWebhookAuth uses default webhook auth unless specified in Subscription CR.
func getEventMeshWebhookAuth(subscription *eventingv1alpha2.Subscription,
	defaultWebhookAuth *types.WebhookAuth) (*types.WebhookAuth, error) {
	auth := &types.WebhookAuth{}
	// extract auth info from subscription CR if any
	if authType, ok := subscription.Spec.Config[eventingv1alpha2.WebhookAuthType]; ok {
		auth.Type = types.GetAuthType(authType)
	} else {
		// if auth type was not provided then use default webhook auth
		return defaultWebhookAuth, nil
	}

	if grantType, ok := subscription.Spec.Config[eventingv1alpha2.WebhookAuthGrantType]; ok {
		auth.GrantType = types.GetGrantType(grantType)
	}

	if tokenURL, ok := subscription.Spec.Config[eventingv1alpha2.WebhookAuthTokenURL]; ok {
		auth.TokenURL = tokenURL
	}

	if clientID, ok := subscription.Spec.Config[eventingv1alpha2.WebhookAuthClientID]; ok {
		auth.ClientID = clientID
	}

	if clientSecret, ok := subscription.Spec.Config[eventingv1alpha2.WebhookAuthClientSecret]; ok {
		auth.ClientSecret = clientSecret
	}

	return auth, nil
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
	if featureflags.IsEventingWebhookAuthEnabled() {
		cleanedEventMeshSub := GetCleanedEventMeshSubscription(subscription)
		newHash, err := GetHash(cleanedEventMeshSub)
		if err != nil {
			return false, err
		}
		return newHash != hash, nil
	}

	// generate has of new subscription
	newHash, err := GetHash(subscription)
	if err != nil {
		return false, err
	}

	// compare hashes
	return newHash != hash, nil
}
