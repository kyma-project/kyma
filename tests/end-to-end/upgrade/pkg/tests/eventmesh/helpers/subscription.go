package helpers

import (
	"fmt"

	"github.com/avast/retry-go"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	eventingv1alpha1 "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/eventmesh/subscription/api/v1alpha1"
)

func WaitForSubscriptionReady(dynamicClient dynamic.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(func() error {
		subscriptionUnstructured, err := dynamicClient.Resource(subscriptionGVR()).Namespace(namespace).Get(name, v1.GetOptions{})
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

func subscriptionGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}
