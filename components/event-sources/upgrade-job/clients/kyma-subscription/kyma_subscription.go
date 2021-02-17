package kyma_subscription

import (
	"encoding/json"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	DynamicClient dynamic.Interface
}

func NewClient(dInf dynamic.Interface) Client {
	return Client{DynamicClient: dInf}
}

func (c Client) List() (*kymaeventingv1alpha1.SubscriptionList, error) {
	subscriptionsUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscriptionList(subscriptionsUnstructured)
}

func (c Client) Get(namespace, name string) (*kymaeventingv1alpha1.Subscription, error) {
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

func (c Client) Create(sub *kymaeventingv1alpha1.Subscription) (*kymaeventingv1alpha1.Subscription, error) {
	unstructuredObj, err := toUnstructured(sub)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert cm to unstructured")
	}
	unstructuredSub, err := c.DynamicClient.
		Resource(GroupVersionResource()).
		Namespace(sub.Namespace).
		Create(unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return toSubscription(unstructuredSub)
}

func toUnstructured(sub *kymaeventingv1alpha1.Subscription) (*unstructured.Unstructured, error) {
	object, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(&sub)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func toSubscription(unstructuredDeployment *unstructured.Unstructured) (*kymaeventingv1alpha1.Subscription, error) {
	subscription := new(kymaeventingv1alpha1.Subscription)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func toSubscriptionList(unstructuredList *unstructured.UnstructuredList) (*kymaeventingv1alpha1.SubscriptionList, error) {
	subscriptionList := new(kymaeventingv1alpha1.SubscriptionList)
	subscriptionListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(subscriptionListBytes, subscriptionList)
	if err != nil {
		return nil, err
	}
	return subscriptionList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  kymaeventingv1alpha1.GroupVersion.Version,
		Group:    kymaeventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: kymaeventingv1alpha1.GroupVersion.Version,
		Group:   kymaeventingv1alpha1.GroupVersion.Group,
		Kind:    "Subscription",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: kymaeventingv1alpha1.GroupVersion.Version,
		Group:   kymaeventingv1alpha1.GroupVersion.Group,
		Kind:    "SubscriptionList",
	}
}
