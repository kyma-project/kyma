package testing

import (
	"fmt"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const FakeChannelName = "fake-chan"

// redefine here to avoid cyclic dependency
const (
	integrationNamespace                      = "kyma-integration"
	applicationNameLabelKey                   = "application-name"
	brokerNamespaceLabelKey                   = "broker-namespace"
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"
	knSubscriptionNamePrefix                  = "brokersub"
)

func NewAppSubscription(appNs, appName string) *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", knSubscriptionNamePrefix),
			Namespace:    integrationNamespace,
			Labels: map[string]string{
				brokerNamespaceLabelKey: appNs,
				applicationNameLabelKey: appName,
			},
		},
	}
}

func NewAppSubscriptionWithSpec(appNs, appName, subscriberURI string) *eventingv1alpha1.Subscription {
	subscription := NewAppSubscription(appNs, appName)
	subscription.Spec = eventingv1alpha1.SubscriptionSpec{
		Channel: corev1.ObjectReference{
			Name: FakeChannelName,
		},
		Subscriber: &eventingv1alpha1.SubscriberSpec{
			URI: &subscriberURI,
		},
	}
	return subscription
}

func NewAppNamespace(name string, brokerInjection bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if brokerInjection {
		ns.Labels = map[string]string{
			knativeEventingInjectionLabelKey: knativeEventingInjectionLabelValueEnabled,
		}
	}
	return ns
}

func NewDefaultBroker(ns string) *eventingv1alpha1.Broker {
	return &eventingv1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: ns,
		},
	}
}

func NewAppChannel(appName string) *messagingv1alpha1.Channel {
	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FakeChannelName,
			Namespace: integrationNamespace,
			Labels: map[string]string{
				applicationNameLabelKey: appName,
			},
		},
	}
}
