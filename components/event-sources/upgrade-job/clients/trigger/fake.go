package trigger

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

func NewFakeClient(triggers *eventingv1alpha1.TriggerList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	triggersUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(triggers)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(triggers)
	if err != nil {
		return Client{}, err
	}

	triggersUnstructured := &unstructured.UnstructuredList{
		Object: triggersUnstructuredMap,
		Items:  unstructuredItems,
	}
	triggersUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, triggersUnstructured)
	return Client{DynamicClient: dynamicClient}, nil
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

func toUnstructuredItems(triggers *eventingv1alpha1.TriggerList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, trigger := range triggers.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&trigger)
		if err != nil {
			return nil, err
		}
		triggerUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		triggerUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *triggerUnstructured)
	}
	return unstructuredItems, nil
}
