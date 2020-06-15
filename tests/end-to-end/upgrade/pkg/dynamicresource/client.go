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
	resCli dynamic.NamespaceableResourceInterface
}

// NewClient function creates a new instance of DynamicResource
func NewClient(dynamicCli dynamic.Interface, s schema.GroupVersionResource) *DynamicResource {
	return &DynamicResource{
		resCli: dynamicCli.Resource(s),
	}
}

// Create method creates a new instance of appropriate k8s resource
func (r *DynamicResource) Create(res metav1.Object) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return err
	}

	_, err = r.resCli.Namespace(res.GetNamespace()).Create(&unstructured.Unstructured{
		Object: u,
	}, metav1.CreateOptions{})

	return err
}

// Get method gets a existing instance of appropriate k8s resource
func (r *DynamicResource) Get(namespace, name string, obj interface{}) error {
	u, err := r.resCli.Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, obj)
}

// Delete method deletes a existing instance of appropriate k8s resource
func (r *DynamicResource) Delete(namespace, name string) error {
	return r.resCli.Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
}
