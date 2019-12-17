package testing

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
)

const FakeChannelName = "fake-chan"
const FakeSubscriptionName = "fake-sub"

// redefine here to avoid cyclic dependency
const (
	integrationNamespace                      = "kyma-integration"
	applicationNameLabelKey                   = "application-name"
	brokerNamespaceLabelKey                   = "broker-namespace"
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"
	knSubscriptionNamePrefix                  = "brokersub"
)

func NewAppSubscription(appNs, appName string, opts ...SubscriptionOption) *messagingv1alpha1.Subscription {
	sub := &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", knSubscriptionNamePrefix),
			Namespace:    integrationNamespace,
			Labels: map[string]string{
				brokerNamespaceLabelKey: appNs,
				applicationNameLabelKey: appName,
			},
		},
	}

	for _, opt := range opts {
		opt(sub)
	}

	return sub
}

// SubscriptionOption is a functional option for Subscription objects.
type SubscriptionOption func(*messagingv1alpha1.Subscription)

// WithSpec sets the spec of a Subscription.
func WithSpec(subscriberURI string) SubscriptionOption {
	url, err := apis.ParseURL(subscriberURI)
	if err != nil {
		panic("todo: nils: throw error during build()")
	}
	return func(s *messagingv1alpha1.Subscription) {
		s.Spec = messagingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Name: FakeChannelName,
			},
			Subscriber: &apisv1alpha1.Destination{
				URI: url,
			},
		}
	}
}

// WithNameSuffix generates the name of a Subscription using its GenerateName prefix.
func WithNameSuffix(nameSuffix string) SubscriptionOption {
	return func(s *messagingv1alpha1.Subscription) {
		if s.GenerateName != "" && s.Name == "" {
			s.Name = s.GenerateName + nameSuffix
		}
	}
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
