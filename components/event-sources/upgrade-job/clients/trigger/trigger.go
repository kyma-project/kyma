package trigger

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

func (c Client) List(namespace string) (*eventingv1alpha1.TriggerList, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	triggersUnstructuredList, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toTriggerList(triggersUnstructuredList)
}

func (c Client) Get(namespace, name string) (*eventingv1alpha1.Trigger, error) {
	unstructuredTrigger, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toTrigger(unstructuredTrigger)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toTrigger(unstructuredDeployment *unstructured.Unstructured) (*eventingv1alpha1.Trigger, error) {
	trigger := new(eventingv1alpha1.Trigger)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, trigger)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func toTriggerList(unstructuredList *unstructured.UnstructuredList) (*eventingv1alpha1.TriggerList, error) {
	triggerList := new(eventingv1alpha1.TriggerList)
	triggerListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(triggerListBytes, triggerList)
	if err != nil {
		return nil, err
	}
	return triggerList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.SchemeGroupVersion.Version,
		Group:    eventingv1alpha1.SchemeGroupVersion.Group,
		Resource: "triggers",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: eventingv1alpha1.SchemeGroupVersion.Version,
		Group:   eventingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "Trigger",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: eventingv1alpha1.SchemeGroupVersion.Version,
		Group:   eventingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "TriggerList",
	}
}
