package eventingbackend

import (
	"context"
	"encoding/json"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// Client struct for EventingBackend client
type Client struct {
	client              dynamic.Interface
	apiExtensionsClient *apixv1beta1client.ApiextensionsV1beta1Client
}

// NewClient creates and returns new client for EventingBackend
func NewClient(client dynamic.Interface, apiExtensionsClient *apixv1beta1client.ApiextensionsV1beta1Client) Client {
	return Client{client, apiExtensionsClient}
}

// Get returns the EventingBackend for specified name and namespace.
// or returns an error if the EventingBackend was not found or other issues
func (c Client) Get(namespace string, name string) (*eventingv1alpha1.EventingBackend, error) {

	ebUnstructured, err := c.client.Resource(GroupVersionResource()).Namespace(namespace).Get(
		context.Background(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	return toEventingBackend(ebUnstructured)
}

// GetCRD returns the EventingBackend CRD
// or returns an error if the EventingBackend CRD was not found or other issues
func (c Client) GetCRD() (*v1beta1.CustomResourceDefinition, error) {
	crd, err := c.apiExtensionsClient.CustomResourceDefinitions().Get(context.Background(), "eventingbackends.eventing.kyma-project.io", metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	return crd, nil
}

// GroupVersionResource return the GVR for EventingBackend k8s resource
func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "eventingbackends",
	}
}

// toEventingBackend converts unstructured EventingBackend object to typed EventingBackend object
func toEventingBackend(unstructured *unstructured.Unstructured) (*eventingv1alpha1.EventingBackend, error) {
	triggerList := new(eventingv1alpha1.EventingBackend)
	triggerListBytes, err := unstructured.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(triggerListBytes, triggerList)
	if err != nil {
		return nil, err
	}
	return triggerList, nil
}
