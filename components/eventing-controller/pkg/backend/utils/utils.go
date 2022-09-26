//nolint:gosec
package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/pkg/errors"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
)

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

// GetInternalView4Ev2 returns the BEB subscription equivalent of Kyma Subscription
// Will be depreciated when Subscription v1alpha2 is active.
func GetInternalView4Ev2(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1beta1.APIRule,
	defaultWebhookAuth *types.WebhookAuth, defaultProtocolSettings *eventingv1alpha1.ProtocolSettings,
	defaultNamespace string, nameMapper NameMapper) (*types.Subscription, error) {
	emsSubscription, err := getDefaultSubscription(defaultProtocolSettings)
	if err != nil {
		return nil, errors.Wrap(err, "apply default protocol settings failed")
	}
	// Name
	emsSubscription.Name = nameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace)

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
	urlTobeRegistered, err := getExposedURLFromAPIRule(apiRule, subscription.Spec.Sink)
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

func ResetStatusToDefaults(sub eventingv1alpha1.Subscription) *eventingv1alpha1.Subscription {
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
		Version:  apigatewayv1beta1.GroupVersion.Version,
		Group:    apigatewayv1beta1.GroupVersion.Group,
		Resource: "apirules",
	}
}

// LoggerWithSubscription returns a logger with the given subscription details.
func LoggerWithSubscription(log *zap.SugaredLogger, subscription *eventingv1alpha1.Subscription) *zap.SugaredLogger {
	return log.With(
		"kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"version", subscription.GetGeneration(),
		"namespace", subscription.GetNamespace(),
		"name", subscription.GetName(),
	)
}
