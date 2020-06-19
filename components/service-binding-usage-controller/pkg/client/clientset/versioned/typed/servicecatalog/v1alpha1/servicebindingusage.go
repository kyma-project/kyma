// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	scheme "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServiceBindingUsagesGetter has a method to return a ServiceBindingUsageInterface.
// A group's client should implement this interface.
type ServiceBindingUsagesGetter interface {
	ServiceBindingUsages(namespace string) ServiceBindingUsageInterface
}

// ServiceBindingUsageInterface has methods to work with ServiceBindingUsage resources.
type ServiceBindingUsageInterface interface {
	Create(ctx context.Context, serviceBindingUsage *v1alpha1.ServiceBindingUsage, opts v1.CreateOptions) (*v1alpha1.ServiceBindingUsage, error)
	Update(ctx context.Context, serviceBindingUsage *v1alpha1.ServiceBindingUsage, opts v1.UpdateOptions) (*v1alpha1.ServiceBindingUsage, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ServiceBindingUsage, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ServiceBindingUsageList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ServiceBindingUsage, err error)
	ServiceBindingUsageExpansion
}

// serviceBindingUsages implements ServiceBindingUsageInterface
type serviceBindingUsages struct {
	client rest.Interface
	ns     string
}

// newServiceBindingUsages returns a ServiceBindingUsages
func newServiceBindingUsages(c *ServicecatalogV1alpha1Client, namespace string) *serviceBindingUsages {
	return &serviceBindingUsages{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the serviceBindingUsage, and returns the corresponding serviceBindingUsage object, and an error if there is any.
func (c *serviceBindingUsages) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ServiceBindingUsage, err error) {
	result = &v1alpha1.ServiceBindingUsage{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicebindingusages").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceBindingUsages that match those selectors.
func (c *serviceBindingUsages) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ServiceBindingUsageList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ServiceBindingUsageList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicebindingusages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceBindingUsages.
func (c *serviceBindingUsages) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("servicebindingusages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a serviceBindingUsage and creates it.  Returns the server's representation of the serviceBindingUsage, and an error, if there is any.
func (c *serviceBindingUsages) Create(ctx context.Context, serviceBindingUsage *v1alpha1.ServiceBindingUsage, opts v1.CreateOptions) (result *v1alpha1.ServiceBindingUsage, err error) {
	result = &v1alpha1.ServiceBindingUsage{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("servicebindingusages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serviceBindingUsage).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a serviceBindingUsage and updates it. Returns the server's representation of the serviceBindingUsage, and an error, if there is any.
func (c *serviceBindingUsages) Update(ctx context.Context, serviceBindingUsage *v1alpha1.ServiceBindingUsage, opts v1.UpdateOptions) (result *v1alpha1.ServiceBindingUsage, err error) {
	result = &v1alpha1.ServiceBindingUsage{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicebindingusages").
		Name(serviceBindingUsage.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serviceBindingUsage).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the serviceBindingUsage and deletes it. Returns an error if one occurs.
func (c *serviceBindingUsages) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicebindingusages").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceBindingUsages) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicebindingusages").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched serviceBindingUsage.
func (c *serviceBindingUsages) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ServiceBindingUsage, err error) {
	result = &v1alpha1.ServiceBindingUsage{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("servicebindingusages").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
