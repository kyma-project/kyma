package secret

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Client struct for k8s Secret client
type Client struct {
	client dynamic.Interface
}

// NewClient creates and returns new client for k8s secrets
func NewClient(client dynamic.Interface) Client {
	return Client{client}
}

// ListByMatchingLabels returns the list of k8s secrets in specified namespace and matching labels.
// or returns an error if it fails for any reason
func (c Client) ListByMatchingLabels(namespace string, labelSelector string) (*corev1.SecretList, error) {

	subscriptionsUnstructured, err := c.client.Resource(GroupVersionResource()).Namespace(namespace).List(
		context.Background(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		return nil, err
	}
	return toSecretList(subscriptionsUnstructured)
}

// GroupVersionResource returns the GVR for secret k8s resource
func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  corev1.SchemeGroupVersion.Version,
		Group:    corev1.SchemeGroupVersion.Group,
		Resource: "secrets",
	}
}

// toSecretList converts unstructured Secret list object to typed object
func toSecretList(unstructuredList *unstructured.UnstructuredList) (*corev1.SecretList, error) {
	triggerList := new(corev1.SecretList)
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
