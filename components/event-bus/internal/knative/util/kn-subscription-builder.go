package util

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SubscriptionBuilder represents the subscription builder that is used in the internal knative util package
// and the knative subscription controller tests.
type SubscriptionBuilder struct {
	*eventingv1alpha1.Subscription
}

// Build returns an v1alpha1.Subscription instance.
func (s *SubscriptionBuilder) Build() *eventingv1alpha1.Subscription {
	return s.Subscription
}

// ToChannel sets SubscriptionBuilder Channel.
func (s *SubscriptionBuilder) ToChannel(name string) *SubscriptionBuilder {
	channel := corev1.ObjectReference{
		Name:       name,
		Kind:       "Channel",
		APIVersion: "messaging.knative.dev/v1alpha1",
	}
	s.Spec.Channel = channel
	s.Spec.Reply.Channel = &channel
	return s
}

// EmptyReply sets the SubscriptionBuilder Reply.
func (s *SubscriptionBuilder) EmptyReply() *SubscriptionBuilder {
	s.Spec.Reply = &eventingv1alpha1.ReplyStrategy{}
	return s
}

// ToK8sService sets the SubscriptionBuilder Subscriber to Kubernetes service.
func (s *SubscriptionBuilder) ToK8sService(k8sServiceName string) *SubscriptionBuilder {
	s.Spec.Subscriber = &eventingv1alpha1.SubscriberSpec{
		Ref: &corev1.ObjectReference{
			Name:       k8sServiceName,
			Kind:       "Service",
			APIVersion: "v1",
		},
	}
	return s
}

// ToKNService sets the SubscriptionBuilder Subscriber to Knative service.
func (s *SubscriptionBuilder) ToKNService(knServiceName string) *SubscriptionBuilder {
	s.Spec.Subscriber = &eventingv1alpha1.SubscriberSpec{
		Ref: &corev1.ObjectReference{
			Name:       knServiceName,
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
	}
	return s
}

// ToURI sets the SubscriptionBuilder Subscriber URI.
func (s *SubscriptionBuilder) ToURI(uri *string) *SubscriptionBuilder {
	s.Spec.Subscriber = &eventingv1alpha1.SubscriberSpec{
		URI: uri,
	}
	return s
}

var (
	emptyString = ""
)

// Subscription returns a new SubscriptionBuilder instance.
func Subscription(name string, namespace string, labels map[string]string) *SubscriptionBuilder {
	subscription := &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Name:       "",
				Kind:       "Channel",
				APIVersion: "messaging.knative.dev/v1alpha1",
			},
			Subscriber: &eventingv1alpha1.SubscriberSpec{
				Ref: &corev1.ObjectReference{
					Name:       "",
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1alpha1",
				},
				URI: &emptyString,
			},
			Reply: &eventingv1alpha1.ReplyStrategy{
				Channel: &corev1.ObjectReference{
					Name:       "",
					Kind:       "Channel",
					APIVersion: "messaging.knative.dev/v1alpha1",
				},
			},
		},
	}
	return &SubscriptionBuilder{
		Subscription: subscription,
	}
}
