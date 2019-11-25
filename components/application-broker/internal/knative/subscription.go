package knative

import (
	"log"
	"regexp"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
)

const (
	// maxPrefixLength for limiting the max length of the name prefix
	maxPrefixLength = 10

	// generatedNameSeparator for adding a separator after the generated name prefix
	generatedNameSeparator = "-"
)

type SubscriptionBuilder struct {
	subscription *eventingv1alpha1.Subscription
}

func Subscription(prefix, namespace string) *SubscriptionBuilder {
	// format the name prefix
	prefix = formatPrefix(prefix, generatedNameSeparator, maxPrefixLength)

	// construct the Knative Subscription object
	subscription := &eventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Namespace:    namespace,
		},
	}

	// init a new Subscription builder
	return &SubscriptionBuilder{subscription: subscription}
}

func FromSubscription(subscription *eventingv1alpha1.Subscription) *SubscriptionBuilder {
	// init a new Subscription builder
	return &SubscriptionBuilder{subscription: subscription}
}

func (b *SubscriptionBuilder) Spec(channel *messagingv1alpha1.Channel, subscriberURI string) *SubscriptionBuilder {
	b.subscription.Spec = eventingv1alpha1.SubscriptionSpec{
		Channel: apicorev1.ObjectReference{
			Name:       channel.Name,
			Kind:       channel.Kind,
			APIVersion: channel.APIVersion,
		},
		Subscriber: &eventingv1alpha1.SubscriberSpec{
			URI: &subscriberURI,
		},
	}
	return b
}

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

func (b *SubscriptionBuilder) Build() *eventingv1alpha1.Subscription {
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
