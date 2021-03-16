package broker

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

type Client struct {
	DynamicClient dynamic.Interface
}

func NewClient(dInf dynamic.Interface) Client {
	return Client{DynamicClient: dInf}
}

func (c Client) List() (*eventingv1alpha1.BrokerList, error) {
	brokersUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toBrokerList(brokersUnstructured)
}

func (c Client) Get(namespace, name string) (*eventingv1alpha1.Broker, error) {
	unstructuredTrigger, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toBroker(unstructuredTrigger)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toBroker(unstructuredBroker *unstructured.Unstructured) (*eventingv1alpha1.Broker, error) {
	broker := new(eventingv1alpha1.Broker)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredBroker.Object, broker)
	if err != nil {
		return nil, err
	}
	return broker, nil
}

func toBrokerList(unstructuredList *unstructured.UnstructuredList) (*eventingv1alpha1.BrokerList, error) {
	brokerList := new(eventingv1alpha1.BrokerList)
	brokerListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(brokerListBytes, brokerList)
	if err != nil {
		return nil, err
	}
	return brokerList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.SchemeGroupVersion.Version,
		Group:    eventingv1alpha1.SchemeGroupVersion.Group,
		Resource: "brokers",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: eventingv1alpha1.SchemeGroupVersion.Version,
		Group:   eventingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "Broker",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: eventingv1alpha1.SchemeGroupVersion.Version,
		Group:   eventingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "BrokerList",
	}
}
