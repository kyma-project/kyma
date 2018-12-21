package knutil

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// Subscriptions
// var uri "bla-bla"
// usage: subscription := Subscription("my-sub", "default").ToChannel("my-channel").ToUri(&uri).EmptyReply().Build()
//

type SubscriptionBuilder struct {
	*eventingv1alpha1.Subscription
}

func (s *SubscriptionBuilder) Build() *eventingv1alpha1.Subscription {
	return s.Subscription
}

func (s *SubscriptionBuilder) ToChannel(name  string) *SubscriptionBuilder {
	s.Spec.Channel = corev1.ObjectReference{
		Name:       name,
		Kind:       "Channel",
		APIVersion: "eventing.knative.dev/v1alpha1",
	}
	return s
}

func (s *SubscriptionBuilder) EmptyReply() *SubscriptionBuilder {
	s.Spec.Reply = &eventingv1alpha1.ReplyStrategy{}
	return s
}

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

func (s *SubscriptionBuilder) ToUri(uri *string) *SubscriptionBuilder {
	s.Spec.Subscriber = &eventingv1alpha1.SubscriberSpec{
		DNSName: uri,
	}
	return s
}

var (
	emptyString = ""
)

func Subscription(name string, namespace string) *SubscriptionBuilder {
	subscription := &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Name:       "",
				Kind:       "Channel",
				APIVersion: "eventing.knative.dev/v1alpha1",
			},
			Subscriber: &eventingv1alpha1.SubscriberSpec{
				Ref: &corev1.ObjectReference{
					Name:       "",
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1alpha1",
				},
				DNSName: &emptyString,
			},
			Reply: &eventingv1alpha1.ReplyStrategy{
				Channel: &corev1.ObjectReference{
					Name:       "",
					Kind:       "Channel",
					APIVersion: "serving.knative.dev/v1alpha1",
				},
			},
		},
	}
	return &SubscriptionBuilder{
		Subscription: subscription,
	}
}
