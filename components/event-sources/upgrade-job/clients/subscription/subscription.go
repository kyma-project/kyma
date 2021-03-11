package subscription

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

func (c Client) List(namespace string) (*messagingv1alpha1.SubscriptionList, error) {
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}
	subscriptionsUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscriptionList(subscriptionsUnstructured)
}

func (c Client) Get(namespace, name string) (*messagingv1alpha1.Subscription, error) {
	unstructuredSubscription, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscription(unstructuredSubscription)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toSubscription(unstructuredSub *unstructured.Unstructured) (*messagingv1alpha1.Subscription, error) {
	subscription := new(messagingv1alpha1.Subscription)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSub.Object, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func toSubscriptionList(unstructuredList *unstructured.UnstructuredList) (*messagingv1alpha1.SubscriptionList, error) {
	triggerList := new(messagingv1alpha1.SubscriptionList)
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
		Version:  messagingv1alpha1.SchemeGroupVersion.Version,
		Group:    messagingv1alpha1.SchemeGroupVersion.Group,
		Resource: "subscriptions",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: messagingv1alpha1.SchemeGroupVersion.Version,
		Group:   messagingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "Subscription",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: messagingv1alpha1.SchemeGroupVersion.Version,
		Group:   messagingv1alpha1.SchemeGroupVersion.Group,
		Kind:    "SubscriptionList",
	}
}
