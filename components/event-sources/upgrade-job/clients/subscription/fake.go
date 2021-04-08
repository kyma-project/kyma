package subscription

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

func NewFakeClient(subscriptions *messagingv1alpha1.SubscriptionList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	subsUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(subscriptions)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(subscriptions)
	if err != nil {
		return Client{}, err
	}

	subscriptionsUnstructured := &unstructured.UnstructuredList{
		Object: subsUnstructuredMap,
		Items:  unstructuredItems,
	}
	subscriptionsUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, subscriptionsUnstructured)
	return Client{DynamicClient: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := messagingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func toUnstructuredItems(subs *messagingv1alpha1.SubscriptionList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, sub := range subs.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&sub)
		if err != nil {
			return nil, err
		}
		subUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		subUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *subUnstructured)
	}
	return unstructuredItems, nil
}
