// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	scheme "github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// BackendModulesGetter has a method to return a BackendModuleInterface.
// A group's client should implement this interface.
type BackendModulesGetter interface {
	BackendModules() BackendModuleInterface
}

// BackendModuleInterface has methods to work with BackendModule resources.
type BackendModuleInterface interface {
	Create(*v1alpha1.BackendModule) (*v1alpha1.BackendModule, error)
	Update(*v1alpha1.BackendModule) (*v1alpha1.BackendModule, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.BackendModule, error)
	List(opts v1.ListOptions) (*v1alpha1.BackendModuleList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BackendModule, err error)
	BackendModuleExpansion
}

// backendModules implements BackendModuleInterface
type backendModules struct {
	client rest.Interface
}

// newBackendModules returns a BackendModules
func newBackendModules(c *UiV1alpha1Client) *backendModules {
	return &backendModules{
		client: c.RESTClient(),
	}
}

// Get takes name of the backendModule, and returns the corresponding backendModule object, and an error if there is any.
func (c *backendModules) Get(name string, options v1.GetOptions) (result *v1alpha1.BackendModule, err error) {
	result = &v1alpha1.BackendModule{}
	err = c.client.Get().
		Resource("backendmodules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BackendModules that match those selectors.
func (c *backendModules) List(opts v1.ListOptions) (result *v1alpha1.BackendModuleList, err error) {
	result = &v1alpha1.BackendModuleList{}
	err = c.client.Get().
		Resource("backendmodules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested backendModules.
func (c *backendModules) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("backendmodules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a backendModule and creates it.  Returns the server's representation of the backendModule, and an error, if there is any.
func (c *backendModules) Create(backendModule *v1alpha1.BackendModule) (result *v1alpha1.BackendModule, err error) {
	result = &v1alpha1.BackendModule{}
	err = c.client.Post().
		Resource("backendmodules").
		Body(backendModule).
		Do().
		Into(result)
	return
}

// Update takes the representation of a backendModule and updates it. Returns the server's representation of the backendModule, and an error, if there is any.
func (c *backendModules) Update(backendModule *v1alpha1.BackendModule) (result *v1alpha1.BackendModule, err error) {
	result = &v1alpha1.BackendModule{}
	err = c.client.Put().
		Resource("backendmodules").
		Name(backendModule.Name).
		Body(backendModule).
		Do().
		Into(result)
	return
}

// Delete takes name of the backendModule and deletes it. Returns an error if one occurs.
func (c *backendModules) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("backendmodules").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *backendModules) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("backendmodules").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched backendModule.
func (c *backendModules) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BackendModule, err error) {
	result = &v1alpha1.BackendModule{}
	err = c.client.Patch(pt).
		Resource("backendmodules").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
