package broker

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

func NewFakeClient(brokers *eventingv1alpha1.BrokerList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	brokersUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(brokers)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(brokers)
	if err != nil {
		return Client{}, err
	}

	brokersUnstructured := &unstructured.UnstructuredList{
		Object: brokersUnstructuredMap,
		Items:  unstructuredItems,
	}
	brokersUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, brokersUnstructured)
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

func toUnstructuredItems(brokerList *eventingv1alpha1.BrokerList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, broker := range brokerList.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&broker)
		if err != nil {
			return nil, err
		}
		brokerUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		brokerUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *brokerUnstructured)
	}
	return unstructuredItems, nil
}
