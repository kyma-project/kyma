// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	scheme "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ClusterMicroFrontendsGetter has a method to return a ClusterMicroFrontendInterface.
// A group's client should implement this interface.
type ClusterMicroFrontendsGetter interface {
	ClusterMicroFrontends() ClusterMicroFrontendInterface
}

// ClusterMicroFrontendInterface has methods to work with ClusterMicroFrontend resources.
type ClusterMicroFrontendInterface interface {
	Create(*v1alpha1.ClusterMicroFrontend) (*v1alpha1.ClusterMicroFrontend, error)
	Update(*v1alpha1.ClusterMicroFrontend) (*v1alpha1.ClusterMicroFrontend, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ClusterMicroFrontend, error)
	List(opts v1.ListOptions) (*v1alpha1.ClusterMicroFrontendList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterMicroFrontend, err error)
	ClusterMicroFrontendExpansion
}

// clusterMicroFrontends implements ClusterMicroFrontendInterface
type clusterMicroFrontends struct {
	client rest.Interface
}

// newClusterMicroFrontends returns a ClusterMicroFrontends
func newClusterMicroFrontends(c *UiV1alpha1Client) *clusterMicroFrontends {
	return &clusterMicroFrontends{
		client: c.RESTClient(),
	}
}

// Get takes name of the clusterMicroFrontend, and returns the corresponding clusterMicroFrontend object, and an error if there is any.
func (c *clusterMicroFrontends) Get(name string, options v1.GetOptions) (result *v1alpha1.ClusterMicroFrontend, err error) {
	result = &v1alpha1.ClusterMicroFrontend{}
	err = c.client.Get().
		Resource("clustermicrofrontends").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterMicroFrontends that match those selectors.
func (c *clusterMicroFrontends) List(opts v1.ListOptions) (result *v1alpha1.ClusterMicroFrontendList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ClusterMicroFrontendList{}
	err = c.client.Get().
		Resource("clustermicrofrontends").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested clusterMicroFrontends.
func (c *clusterMicroFrontends) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("clustermicrofrontends").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a clusterMicroFrontend and creates it.  Returns the server's representation of the clusterMicroFrontend, and an error, if there is any.
func (c *clusterMicroFrontends) Create(clusterMicroFrontend *v1alpha1.ClusterMicroFrontend) (result *v1alpha1.ClusterMicroFrontend, err error) {
	result = &v1alpha1.ClusterMicroFrontend{}
	err = c.client.Post().
		Resource("clustermicrofrontends").
		Body(clusterMicroFrontend).
		Do().
		Into(result)
	return
}

// Update takes the representation of a clusterMicroFrontend and updates it. Returns the server's representation of the clusterMicroFrontend, and an error, if there is any.
func (c *clusterMicroFrontends) Update(clusterMicroFrontend *v1alpha1.ClusterMicroFrontend) (result *v1alpha1.ClusterMicroFrontend, err error) {
	result = &v1alpha1.ClusterMicroFrontend{}
	err = c.client.Put().
		Resource("clustermicrofrontends").
		Name(clusterMicroFrontend.Name).
		Body(clusterMicroFrontend).
		Do().
		Into(result)
	return
}

// Delete takes name of the clusterMicroFrontend and deletes it. Returns an error if one occurs.
func (c *clusterMicroFrontends) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("clustermicrofrontends").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *clusterMicroFrontends) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("clustermicrofrontends").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched clusterMicroFrontend.
func (c *clusterMicroFrontends) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterMicroFrontend, err error) {
	result = &v1alpha1.ClusterMicroFrontend{}
	err = c.client.Patch(pt).
		Resource("clustermicrofrontends").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
