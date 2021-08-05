package subscription

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(subscriptions *eventingv1alpha1.SubscriptionList) (Client, error) {
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
	return Client{client: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func toUnstructuredItems(subs *eventingv1alpha1.SubscriptionList) ([]unstructured.Unstructured, error) {
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
