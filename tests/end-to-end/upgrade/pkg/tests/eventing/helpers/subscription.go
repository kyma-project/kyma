package helpers

import (
	"fmt"

	"github.com/avast/retry-go"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	eventingv1alpha1 "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventing/subscription/api/v1alpha1"
)

const (
	prefix = "sap.kyma.custom"
)

func WaitForSubscriptionReady(dynamicClient dynamic.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(func() error {
		subscriptionUnstructured, err := dynamicClient.Resource(subscriptionGVR()).Namespace(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		subscription := &eventingv1alpha1.Subscription{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(subscriptionUnstructured.Object, subscription); err != nil {
			return err
		}

		if !subscription.Status.Ready {
			return fmt.Errorf("subscription %s/%s is not ready", namespace, name)
		}

		return nil
	}, retryOptions...)
}

type SubOption func(trigger *eventingv1alpha1.Subscription)

func CreateSubscription(dynamicClient dynamic.Interface, name, namespace string, subOptions ...SubOption) error {
	labelSelector := map[string]string{
		"function": name,
	}
	sub := &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labelSelector,
		},
		Spec:   eventingv1alpha1.SubscriptionSpec{
			ProtocolSettings: &eventingv1alpha1.ProtocolSettings{},
		},
		Status: eventingv1alpha1.SubscriptionStatus{},
	}

	for _, option := range subOptions {
		option(sub)
	}
	subUnstructuredObj, err := toUnstructured(sub)

	_, err = dynamicClient.Resource(subscriptionGVR()).Namespace(namespace).Create(subUnstructuredObj, metav1.CreateOptions{})

	return err
}

func WithFilter(eventVersion, eventTypeRaw, appName string) SubOption {
	eventName := fmt.Sprintf("%s.%s.%s.%s", prefix, appName, eventTypeRaw, eventVersion)
	eventSource := &eventingv1alpha1.Filter{
		Type:     "exact",
		Property: "source",
		Value:    "", // For NATS there is no source needed.
	}
	eventType := &eventingv1alpha1.Filter{
		Type:     "exact",
		Property: "type",
		Value:    eventName,
	}

	bebFilter := &eventingv1alpha1.BebFilter{
		EventSource: eventSource,
		EventType:   eventType,
	}
	bebFilters := &eventingv1alpha1.BebFilters{
		Dialect: "",
		Filters: []*eventingv1alpha1.BebFilter{
			bebFilter,
		},
	}

	return func(sub *eventingv1alpha1.Subscription) {
		sub.Spec.Filter = bebFilters
	}
}

func WithSink(sink string) SubOption {
	return func(sub *eventingv1alpha1.Subscription) {
		sub.Spec.Sink = sink
	}
}

func toUnstructured(sub *eventingv1alpha1.Subscription) (*unstructured.Unstructured, error) {
	object, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(&sub)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func subscriptionGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}
