package configmap

import (
	"context"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Client struct for k8s configMap client
type Client struct {
	client dynamic.Interface
}

// NewClient creates and returns new client for k8s configMaps
func NewClient(client dynamic.Interface) Client {
	return Client{client}
}

// Get returns the k8s configMap in specified namespace
// or returns an error if it fails for any reason
func (c Client) Get(namespace string, name string) (*corev1.ConfigMap, error) {

	subscriptionsUnstructured, err := c.client.Resource(GroupVersionResource()).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	return toConfigMap(subscriptionsUnstructured)
}

// GroupVersionResource returns the GVR for secret k8s resource
func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  corev1.SchemeGroupVersion.Version,
		Group:    corev1.SchemeGroupVersion.Group,
		Resource: "configmaps",
	}
}

// GroupVersionKind return the GVK for Secret k8s resource
func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: corev1.SchemeGroupVersion.Version,
		Group:   corev1.SchemeGroupVersion.Group,
		Kind:    "ConfigMap",
	}
}

// GroupVersionKindList return the GVK list for Secret k8s resource
func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: corev1.SchemeGroupVersion.Version,
		Group:   corev1.SchemeGroupVersion.Group,
		Kind:    "ConfigMapList",
	}
}

// toConfigMap converts unstructured deployment object to typed ConfigMap object
func toConfigMap(unstructuredDeployment *unstructured.Unstructured) (*corev1.ConfigMap, error) {
	cm := new(corev1.ConfigMap)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, cm)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
