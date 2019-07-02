// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	scheme "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AddonsConfigurationsGetter has a method to return a AddonsConfigurationInterface.
// A group's client should implement this interface.
type AddonsConfigurationsGetter interface {
	AddonsConfigurations(namespace string) AddonsConfigurationInterface
}

// AddonsConfigurationInterface has methods to work with AddonsConfiguration resources.
type AddonsConfigurationInterface interface {
	Create(*v1alpha1.AddonsConfiguration) (*v1alpha1.AddonsConfiguration, error)
	Update(*v1alpha1.AddonsConfiguration) (*v1alpha1.AddonsConfiguration, error)
	UpdateStatus(*v1alpha1.AddonsConfiguration) (*v1alpha1.AddonsConfiguration, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.AddonsConfiguration, error)
	List(opts v1.ListOptions) (*v1alpha1.AddonsConfigurationList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AddonsConfiguration, err error)
	AddonsConfigurationExpansion
}

// addonsConfigurations implements AddonsConfigurationInterface
type addonsConfigurations struct {
	client rest.Interface
	ns     string
}

// newAddonsConfigurations returns a AddonsConfigurations
func newAddonsConfigurations(c *AddonsV1alpha1Client, namespace string) *addonsConfigurations {
	return &addonsConfigurations{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the addonsConfiguration, and returns the corresponding addonsConfiguration object, and an error if there is any.
func (c *addonsConfigurations) Get(name string, options v1.GetOptions) (result *v1alpha1.AddonsConfiguration, err error) {
	result = &v1alpha1.AddonsConfiguration{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AddonsConfigurations that match those selectors.
func (c *addonsConfigurations) List(opts v1.ListOptions) (result *v1alpha1.AddonsConfigurationList, err error) {
	result = &v1alpha1.AddonsConfigurationList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested addonsConfigurations.
func (c *addonsConfigurations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a addonsConfiguration and creates it.  Returns the server's representation of the addonsConfiguration, and an error, if there is any.
func (c *addonsConfigurations) Create(addonsConfiguration *v1alpha1.AddonsConfiguration) (result *v1alpha1.AddonsConfiguration, err error) {
	result = &v1alpha1.AddonsConfiguration{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		Body(addonsConfiguration).
		Do().
		Into(result)
	return
}

// Update takes the representation of a addonsConfiguration and updates it. Returns the server's representation of the addonsConfiguration, and an error, if there is any.
func (c *addonsConfigurations) Update(addonsConfiguration *v1alpha1.AddonsConfiguration) (result *v1alpha1.AddonsConfiguration, err error) {
	result = &v1alpha1.AddonsConfiguration{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		Name(addonsConfiguration.Name).
		Body(addonsConfiguration).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *addonsConfigurations) UpdateStatus(addonsConfiguration *v1alpha1.AddonsConfiguration) (result *v1alpha1.AddonsConfiguration, err error) {
	result = &v1alpha1.AddonsConfiguration{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		Name(addonsConfiguration.Name).
		SubResource("status").
		Body(addonsConfiguration).
		Do().
		Into(result)
	return
}

// Delete takes name of the addonsConfiguration and deletes it. Returns an error if one occurs.
func (c *addonsConfigurations) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *addonsConfigurations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("addonsconfigurations").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched addonsConfiguration.
func (c *addonsConfigurations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AddonsConfiguration, err error) {
	result = &v1alpha1.AddonsConfiguration{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("addonsconfigurations").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
