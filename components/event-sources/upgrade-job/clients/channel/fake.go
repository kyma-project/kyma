package channel

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

func NewFakeClient(channels *messagingv1alpha1.ChannelList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	channelsUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(channels)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(channels)
	if err != nil {
		return Client{}, err
	}

	channelsUnstructured := &unstructured.UnstructuredList{
		Object: channelsUnstructuredMap,
		Items:  unstructuredItems,
	}
	channelsUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, channelsUnstructured)
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

func toUnstructuredItems(channelList *messagingv1alpha1.ChannelList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, channel := range channelList.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&channel)
		if err != nil {
			return nil, err
		}
		channelUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		channelUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *channelUnstructured)
	}
	return unstructuredItems, nil
}
