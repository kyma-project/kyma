package deployment

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/dynamic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Client struct for deployment client
type Client struct {
	client dynamic.Interface
}

// NewClient creates and returns new client for k8s deployments
func NewClient(client dynamic.Interface) Client {
	return Client{client}
}

// Get returns the deployment for specified name and namespace.
// or returns an error if the deployment is not found or other issues
func (c Client) Get(namespace, name string) (*appsv1.Deployment, error) {
	unstructuredDeployment, err := c.client.Resource(GroupVersionResource()).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toDeployment(unstructuredDeployment)
}

// Update updated the specified deployment to desired state.
// It return the new updated deployment
// or returns an error if it fails to update the deployment
func (c Client) Update(namespace string, desiredDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	// Unmarshal from typed to unstructured
	data, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(desiredDeployment)
	if err != nil {
		return nil, err
	}
	unstructuredObj := &unstructured.Unstructured{
		Object: data,
	}

	unstructuredDeployment, err := c.client.Resource(GroupVersionResource()).Namespace(namespace).Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return toDeployment(unstructuredDeployment)
}

// Delete deletes the specified deployment.
// It return nil if the deletion is successful
// or returns an error if it fails to delete the deployment
func (c Client) Delete(namespace, name string) error {
	err := c.client.Resource(GroupVersionResource()).Namespace(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// GroupVersionResource return the GVR for Deployment k8s resource
func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  appsv1.SchemeGroupVersion.Version,
		Group:    appsv1.SchemeGroupVersion.Group,
		Resource: "deployments",
	}
}

// toDeployment converts unstructured deployment object to typed deployment object
func toDeployment(unstructuredDeployment *unstructured.Unstructured) (*appsv1.Deployment, error) {
	deployment := new(appsv1.Deployment)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}