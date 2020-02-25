package knative

import (
	"log"
	"regexp"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// maxPrefixLength for limiting the max length of the name prefix
	maxPrefixLength = 10

	// generatedNameSeparator for adding a separator after the generated name prefix
	generatedNameSeparator = "-"
)

type SubscriptionBuilder struct {
	subscription *messagingv1alpha1.Subscription
}

// Subscription function returns a Subscription Builder Object what the invoker can use to Build a Knative Subscription
// in a Fluent API manner.
func Subscription(prefix, namespace string) *SubscriptionBuilder {
	// format the name prefix
	prefix = formatPrefix(prefix, generatedNameSeparator, maxPrefixLength)

	// construct the Knative Subscription object
	subscription := &messagingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Namespace:    namespace,
		},
	}

	// init a new Subscription builder
	return &SubscriptionBuilder{subscription: subscription}
}

// Build a knative subscription builder from an existing knative subscription object
// Using the fluent API, the invoker can change the subscription specs and build its own Knative Subscription
func FromSubscription(subscription *messagingv1alpha1.Subscription) *SubscriptionBuilder {
	// init a new Subscription builder
	return &SubscriptionBuilder{subscription: subscription}
}

// Spec adds the specification to the Knative Subscription Object
func (b *SubscriptionBuilder) Spec(channel *messagingv1alpha1.Channel, subscriberURI string) *SubscriptionBuilder {
	url, err := apis.ParseURL(subscriberURI)
	if err != nil {
		panic("todo: nils, return error in build()")
	}
	b.subscription.Spec = messagingv1alpha1.SubscriptionSpec{
		Channel: apicorev1.ObjectReference{
			Name:       channel.Name,
			Kind:       channel.Kind,
			APIVersion: channel.APIVersion,
		},
		Subscriber: &duckv1.Destination{
			URI: url,
		},
	}
	return b
}

// Add Labels to the Knative Subscription Object
func (b *SubscriptionBuilder) Labels(labels map[string]string) *SubscriptionBuilder {
	if len(labels) == 0 {
		return b
	}
	if len(b.subscription.Labels) == 0 {
		b.subscription.Labels = labels
		return b
	}
	for k, v := range labels {
		b.subscription.Labels[k] = v
	}
	return b
}

// Build is a final function, as per the Fluent Interface Design and returns the invoker a Knative Subscription CR,
// which can be used in the knative eventing go client to perform a typical kubernetes operation.
func (b *SubscriptionBuilder) Build() *messagingv1alpha1.Subscription {
	return b.subscription
}

// formatPrefix returns a new string for the prefix that is limited in the length, not having special characters, and
// has the separator appended to it in the end.
func formatPrefix(prefix, separator string, length int) string {
	// prepare special characters regex
	reg, err := regexp.Compile("[^a-z0-9]+")
	if err != nil {
		log.Fatal(err)
	}

	// limit the prefix length
	if len(prefix) > length {
		prefix = prefix[:length]
	}

	// remove the special characters and append the separator
	prefix = reg.ReplaceAllString(prefix, "") + separator

	return prefix
}
