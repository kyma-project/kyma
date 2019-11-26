package util

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
)

// SubscriptionBuilder represents the subscription builder that is used in the internal knative util package
// and the knative subscription controller tests.
type SubscriptionBuilder struct {
	*messagingv1alpha1.Subscription
}

// Build returns an v1alpha1.Subscription instance.
func (s *SubscriptionBuilder) Build() *messagingv1alpha1.Subscription {
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
	destination := apisv1alpha1.Destination{
		Ref: &channel,
	}
	s.Spec.Reply.Channel = &destination
	return s
}

// EmptyReply sets the SubscriptionBuilder Reply.
func (s *SubscriptionBuilder) EmptyReply() *SubscriptionBuilder {
	s.Spec.Reply = &messagingv1alpha1.ReplyStrategy{}
	return s
}

// ToK8sService sets the SubscriptionBuilder Subscriber to Kubernetes service.
func (s *SubscriptionBuilder) ToK8sService(k8sServiceName string) *SubscriptionBuilder {
	SubscriptionSpec := &eventingv1alpha1.SubscriberSpec{
		DeprecatedRef: &corev1.ObjectReference{
			Name:       k8sServiceName,
			Kind:       "Service",
			APIVersion: "v1",
		},
	}
	destination := apisv1alpha1.Destination{
		Ref: SubscriptionSpec.DeprecatedRef,
	}
	s.Spec.Subscriber = &destination
	return s
}

// ToKNService sets the SubscriptionBuilder Subscriber to Knative service.
func (s *SubscriptionBuilder) ToKNService(knServiceName string) *SubscriptionBuilder {
	SubscriptionSpec := &eventingv1alpha1.SubscriberSpec{
		DeprecatedRef: &corev1.ObjectReference{
			Name:       knServiceName,
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
	}
	destination := apisv1alpha1.Destination{
		Ref: SubscriptionSpec.DeprecatedRef,
	}
	s.Spec.Subscriber = &destination
	return s
}

// ToURI sets the SubscriptionBuilder Subscriber URI.
func (s *SubscriptionBuilder) ToURI(uri *string) *SubscriptionBuilder {
	url, err := apis.ParseURL(*uri)
	if err != nil {
		//TODO maybe not the best to panic here, instead return an error
		panic(fmt.Sprintf("Couldn't parse the subscriber URI: %+v", err))
	}
	destination := apisv1alpha1.Destination{
		URI: url,
	}
	s.Spec.Subscriber = &destination
	return s
}

var (
	emptyString = ""
)

// Subscription returns a new SubscriptionBuilder instance.
func Subscription(name string, namespace string, labels map[string]string) *SubscriptionBuilder {
	Subscriber := &eventingv1alpha1.SubscriberSpec{
		DeprecatedRef: &corev1.ObjectReference{
			Name:       "",
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		SubscriberURI: emptyString,
	}
	SubscriberDestination := apisv1alpha1.Destination{
		Ref: Subscriber.DeprecatedRef,
	}

	channel := corev1.ObjectReference{
		Name:       "",
		Kind:       "Channel",
		APIVersion: "messaging.knative.dev/v1alpha1",
	}

	channelDestination := apisv1alpha1.Destination{
		Ref: &channel,
	}

	subscription := &messagingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: messagingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: messagingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				Name:       "",
				Kind:       "Channel",
				APIVersion: "messaging.knative.dev/v1alpha1",
			},
			Subscriber: &SubscriberDestination,
			Reply: &messagingv1alpha1.ReplyStrategy{
				Channel: &channelDestination,
			},
		},
	}
	return &SubscriptionBuilder{
		Subscription: subscription,
	}
}
