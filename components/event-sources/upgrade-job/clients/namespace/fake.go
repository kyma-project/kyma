package namespace

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

func NewFakeClient(namespaces *corev1.NamespaceList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(namespaces)
	if err != nil {
		return Client{}, err
	}

	nsUnstructuredItems := &unstructured.UnstructuredList{
		Items: unstructuredItems,
	}
	nsUnstructuredItems.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, nsUnstructuredItems)
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

func toUnstructuredItems(namespaces *corev1.NamespaceList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, ns := range namespaces.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ns)
		if err != nil {
			return nil, err
		}
		nsUnstructured := unstructured.Unstructured{Object: unstructuredMap}
		nsUnstructured.SetGroupVersionKind(GroupVersionKind())

		unstructuredItems = append(unstructuredItems, nsUnstructured)
	}
	return unstructuredItems, nil
}
