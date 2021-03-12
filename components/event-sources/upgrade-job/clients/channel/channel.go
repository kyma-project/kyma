package channel

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

type Client struct {
	DynamicClient dynamic.Interface
}

func NewClient(dInf dynamic.Interface) Client {
	return Client{DynamicClient: dInf}
}

func (c Client) List(namespace string, labelSelector string) (*messagingv1alpha1.ChannelList, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	channelsUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	return toChannelList(channelsUnstructured)
}

func (c Client) Get(namespace, name string) (*messagingv1alpha1.Channel, error) {
	unstructuredChannel, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toChannel(unstructuredChannel)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toChannel(unstructuredChannel *unstructured.Unstructured) (*messagingv1alpha1.Channel, error) {
	channel := new(messagingv1alpha1.Channel)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredChannel.Object, channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func toChannelList(unstructuredList *unstructured.UnstructuredList) (*messagingv1alpha1.ChannelList, error) {
	channelList := new(messagingv1alpha1.ChannelList)
	channelListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(channelListBytes, channelList)
	if err != nil {
		return nil, err
	}
	return channelList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  messagingv1alpha1.SchemeGroupVersion.Version,
		Group:    messagingv1alpha1.SchemeGroupVersion.Group,
		Resource: "channels",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: messagingv1alpha1.SchemeGroupVersion.Version,
		Group:   messagingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "Channel",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: messagingv1alpha1.SchemeGroupVersion.Version,
		Group:   messagingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "ChannelList",
	}
}
