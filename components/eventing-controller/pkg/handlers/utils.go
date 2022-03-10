//nolint:gosec
package handlers

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// NameMapper is used to map Kyma-specific resource names to their corresponding name on other
// (external) systems, e.g. on different eventing backends, the same Kyma subscription name
// could map to a different name.
type NameMapper interface {
	MapSubscriptionName(sub *eventingv1alpha1.Subscription) string
}

// bebSubscriptionNameMapper maps a Kyma subscription to an ID that can be used on the BEB backend,
// which has a max length. Domain name is used to make the names on BEB unique.
type bebSubscriptionNameMapper struct {
	domainName string
	maxLength  int
}

func NewBEBSubscriptionNameMapper(domainName string, maxNameLength int) NameMapper {
	return &bebSubscriptionNameMapper{
		domainName: domainName,
		maxLength:  maxNameLength,
	}
}

func (m *bebSubscriptionNameMapper) MapSubscriptionName(sub *eventingv1alpha1.Subscription) string {
	hash := hashSubscriptionFullName(m.domainName, sub.Namespace, sub.Name)
	return shortenNameAndAppendHash(sub.Name, hash, m.maxLength)
}

func hashSubscriptionFullName(domainName, namespace, name string) string {
	hash := sha1.Sum([]byte(domainName + namespace + name))
	return fmt.Sprintf("%x", hash)
}

func encodeString(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func decodeString(value string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// produces a name+hash which is not longer than maxLength. If necessary, shortens name, not the hash.
// Requires maxLength >= len(hash)
func shortenNameAndAppendHash(name string, hash string, maxLength int) string {
	if len(hash) > maxLength {
		// This shouldn't happen!
		panic(fmt.Sprintf("max name length (%d) used for BEB subscription mapper is not large enough to hold the hash (%s)", maxLength, hash))
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

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func getHash(subscription *types.Subscription) (int64, error) {
	hash, err := hashstructure.Hash(subscription, hashstructure.FormatV2, nil)
	if err != nil {
		return 0, err
	}
	return int64(hash), nil
}

func getDefaultSubscription(protocolSettings *eventingv1alpha1.ProtocolSettings) (*types.Subscription, error) {
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

func getInternalView4Ev2(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule,
	defaultWebhookAuth *types.WebhookAuth, defaultProtocolSettings *eventingv1alpha1.ProtocolSettings,
	defaultNamespace string, nameMapper NameMapper) (*types.Subscription, error) {
	emsSubscription, err := getDefaultSubscription(defaultProtocolSettings)
	if err != nil {
		return nil, errors.Wrap(err, "apply default protocol settings failed")
	}
	// Name
	emsSubscription.Name = nameMapper.MapSubscriptionName(subscription)

	// Applying protocol settings if provided in subscription CR
	if subscription.Spec.ProtocolSettings != nil {
		if subscription.Spec.ProtocolSettings.ContentMode != nil {
			emsSubscription.ContentMode = *subscription.Spec.ProtocolSettings.ContentMode
		}
		// ExemptHandshake
		if subscription.Spec.ProtocolSettings.ExemptHandshake != nil {
			emsSubscription.ExemptHandshake = *subscription.Spec.ProtocolSettings.ExemptHandshake
		}
		// Qos
		if subscription.Spec.ProtocolSettings.Qos != nil {
			qos, err := getQos(*subscription.Spec.ProtocolSettings.Qos)
			if err != nil {
				return nil, err
			}
			emsSubscription.Qos = qos
		}
	}

	// WebhookURL
	urlTobeRegistered, err := getExposedURLFromAPIRule(apiRule, subscription)
	if err != nil {
		return nil, errors.Wrap(err, "get APIRule exposed URL failed")
	}
	emsSubscription.WebhookURL = urlTobeRegistered

	// Events
	uniqueFilters, err := subscription.Spec.Filter.Deduplicate()
	if err != nil {
		return nil, errors.Wrap(err, "deduplicate subscription filters failed")
	}
	for _, e := range uniqueFilters.Filters {
		s := defaultNamespace
		if e.EventSource.Value != "" {
			s = e.EventSource.Value
		}
		t := e.EventType.Value
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

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
	emsSubscription.WebhookAuth = auth
	return emsSubscription, nil
}

func getExposedURLFromAPIRule(apiRule *apigatewayv1alpha1.APIRule, sub *eventingv1alpha1.Subscription) (string, error) {
	scheme := "https://"
	path := ""

	sURL, err := url.ParseRequestURI(sub.Spec.Sink)
	if err != nil {
		return "", err
	}
	sURLPath := sURL.Path
	if sURL.Path == "" {
		sURLPath = "/"
	}
	for _, rule := range apiRule.Spec.Rules {
		if rule.Path == sURLPath {
			path = rule.Path
			break
		}
	}
	return fmt.Sprintf("%s%s%s", scheme, *apiRule.Spec.Service.Host, path), nil
}

func getInternalView4Ems(subscription *types.Subscription) *types.Subscription {
	emsSubscription := &types.Subscription{}

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.ContentMode
	emsSubscription.ExemptHandshake = subscription.ExemptHandshake

	// Qos
	emsSubscription.Qos = subscription.Qos

	// WebhookURL
	emsSubscription.WebhookURL = subscription.WebhookURL

	// Events
	for _, e := range subscription.Events {
		s := e.Source
		t := e.Type
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

	return emsSubscription
}

// GetRandString returns a random string of the given length
func GetRandString(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RemoveStatus(sub eventingv1alpha1.Subscription) *eventingv1alpha1.Subscription {
	desiredSub := sub.DeepCopy()
	desiredSub.Status = eventingv1alpha1.SubscriptionStatus{}
	return desiredSub
}

func SetStatusAsNotReady(sub eventingv1alpha1.Subscription) *eventingv1alpha1.Subscription {
	desiredSub := sub.DeepCopy()
	desiredSub.Status.Ready = false
	return desiredSub
}

func UpdateSubscriptionStatus(ctx context.Context, dClient dynamic.Interface, sub *eventingv1alpha1.Subscription) error {
	unstructuredObj, err := toUnstructuredSub(sub)
	if err != nil {
		return errors.Wrap(err, "convert subscription to unstructured failed")
	}
	_, err = dClient.
		Resource(SubscriptionGroupVersionResource()).
		Namespace(sub.Namespace).
		UpdateStatus(ctx, unstructuredObj, metav1.UpdateOptions{})

	return err
}

func ToSubscriptionList(unstructuredList *unstructured.UnstructuredList) (*eventingv1alpha1.SubscriptionList, error) {
	subscriptionList := new(eventingv1alpha1.SubscriptionList)
	subscriptionListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(subscriptionListBytes, subscriptionList)
	if err != nil {
		return nil, err
	}
	return subscriptionList, nil
}

func toUnstructuredSub(sub *eventingv1alpha1.Subscription) (*unstructured.Unstructured, error) {
	object, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(&sub)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

func APIRuleGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  apigatewayv1alpha1.GroupVersion.Version,
		Group:    apigatewayv1alpha1.GroupVersion.Group,
		Resource: "apirules",
	}
}
