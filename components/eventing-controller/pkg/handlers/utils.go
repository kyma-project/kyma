package handlers

import (
	"context"
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
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
)

// MessagingBackend exposes a common handler interface for different messaging backend systems
type MessagingBackend interface {
	// Initialize should initialize the communication layer with the messaging backend system
	Initialize(cfg env.Config) error

	// SyncSubscription should synchronize the Kyma eventing susbscription with the susbcriber infrastructure of messaging backend system.
	// It should return true if Kyma eventing subscription status was changed during this synchronization process.
	// It sets subscription.status.config with configurations that were applied on the messaging backend when creating the subscription.
	// TODO: Give up the usage of variadic parameters in the favor of using only subscription as input parameter.
	// TODO: This should contain all the infos necessary for the handler to do its job.
	SyncSubscription(subscription *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, params ...interface{}) (bool, error)

	// DeleteSubscription should delete the corresponding subscriber data of messaging backend
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func getHash(subscription *types.Subscription) (int64, error) {
	if hash, err := hashstructure.Hash(subscription, hashstructure.FormatV2, nil); err != nil {
		return 0, err
	} else {
		return int64(hash), nil
	}
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

func getInternalView4Ev2(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule, defaultWebhookAuth *types.WebhookAuth, defaultProtocolSettings *eventingv1alpha1.ProtocolSettings, defaultNamespace string) (*types.Subscription, error) {
	emsSubscription, err := getDefaultSubscription(defaultProtocolSettings)
	if err != nil {
		return nil, errors.Wrap(err, "apply default protocol settings failed")
	}
	// Name
	emsSubscription.Name = subscription.Name

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

	// WebhookUrl
	urlTobeRegistered, err := getExposedURLFromAPIRule(apiRule, subscription)
	if err != nil {
		return nil, errors.Wrap(err, "get APIRule exposed URL failed")
	}
	emsSubscription.WebhookUrl = urlTobeRegistered

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
		auth.ClientID = subscription.Spec.ProtocolSettings.WebhookAuth.ClientId
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
		auth.TokenURL = subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl
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

func getInternalView4Ems(subscription *types.Subscription) (*types.Subscription, error) {
	emsSubscription := &types.Subscription{}

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.ContentMode
	emsSubscription.ExemptHandshake = subscription.ExemptHandshake

	// Qos
	emsSubscription.Qos = subscription.Qos

	// WebhookUrl
	emsSubscription.WebhookUrl = subscription.WebhookUrl

	// Events
	for _, e := range subscription.Events {
		s := e.Source
		t := e.Type
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

	return emsSubscription, nil
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
