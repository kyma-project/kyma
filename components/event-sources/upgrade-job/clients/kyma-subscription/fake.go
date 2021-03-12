package kyma_subscription

import (
	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(subscriptions *kymaeventingv1alpha1.SubscriptionList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	if subscriptions == nil {
		dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
		return Client{DynamicClient: dynamicClient}, nil
	}

	subUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(subscriptions)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(subscriptions)
	if err != nil {
		return Client{}, err
	}

	subUnstructured := &unstructured.UnstructuredList{
		Object: subUnstructuredMap,
		Items:  unstructuredItems,
	}
	subUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, subUnstructured)
	return Client{DynamicClient: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := kymaeventingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func toUnstructuredItems(subList *kymaeventingv1alpha1.SubscriptionList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, sub := range subList.Items {
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
