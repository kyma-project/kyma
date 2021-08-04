package subscription

import (
	"context"
	"encoding/json"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Client struct for Kyma Subscription client
type Client struct {
	client dynamic.Interface
}

// NewClient creates and returns new client for Kyma Subscriptions
func NewClient(client dynamic.Interface) Client {
	return Client{client}
}

// List returns the list of kyma subscriptions in specified namespace
// or returns an error if it fails for any reason
func (c Client) List(namespace string) (*eventingv1alpha1.SubscriptionList, error) {

	subscriptionsUnstructured, err := c.client.Resource(GroupVersionResource()).Namespace(namespace).List(
		context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}
	return toSubscriptionList(subscriptionsUnstructured)
}

// GroupVersionResource returns the GVR for Subscription resource
func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

// toSecretList converts unstructured Subscription list object to typed object
func toSubscriptionList(unstructuredList *unstructured.UnstructuredList) (*eventingv1alpha1.SubscriptionList, error) {
	triggerList := new(eventingv1alpha1.SubscriptionList)
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
