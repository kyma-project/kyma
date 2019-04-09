package dynamicresource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// DynamicResource is a struct for dynamic resources
type DynamicResource struct {
	resCli    dynamic.ResourceInterface
	namespace string
	kind      string
}

// NewClient function creates a new instance of DynamicResource
func NewClient(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string) *DynamicResource {
	return &DynamicResource{
		resCli:    dynamicCli.Resource(s).Namespace(namespace),
		namespace: namespace,
		kind:      s.Resource,
	}
}

// Create method creates a new instance of appropriate k8s resource
func (r *DynamicResource) Create(res interface{}) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return err
	}

	_, err = r.resCli.Create(&unstructured.Unstructured{
		Object: u,
	}, metav1.CreateOptions{})

	return err
}

// Get method gets a existing instance of appropriate k8s resource
func (r *DynamicResource) Get(name string) (*unstructured.Unstructured, error) {
	return r.resCli.Get(name, metav1.GetOptions{})
}

// Delete method deletes a existing instance of appropriate k8s resource
func (r *DynamicResource) Delete(name string) error {
	return r.resCli.Delete(name, &metav1.DeleteOptions{})
}
