// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	scheme "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CompassConnectionsGetter has a method to return a CompassConnectionInterface.
// A group's client should implement this interface.
type CompassConnectionsGetter interface {
	CompassConnections() CompassConnectionInterface
}

// CompassConnectionInterface has methods to work with CompassConnection resources.
type CompassConnectionInterface interface {
	Create(*v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
	Update(*v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
	UpdateStatus(*v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CompassConnection, error)
	List(opts v1.ListOptions) (*v1alpha1.CompassConnectionList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CompassConnection, err error)
	CompassConnectionExpansion
}

// compassConnections implements CompassConnectionInterface
type compassConnections struct {
	client rest.Interface
}

// newCompassConnections returns a CompassConnections
func newCompassConnections(c *CompassV1alpha1Client) *compassConnections {
	return &compassConnections{
		client: c.RESTClient(),
	}
}

// Get takes name of the compassConnection, and returns the corresponding compassConnection object, and an error if there is any.
func (c *compassConnections) Get(name string, options v1.GetOptions) (result *v1alpha1.CompassConnection, err error) {
	result = &v1alpha1.CompassConnection{}
	err = c.client.Get().
		Resource("compassconnections").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CompassConnections that match those selectors.
func (c *compassConnections) List(opts v1.ListOptions) (result *v1alpha1.CompassConnectionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.CompassConnectionList{}
	err = c.client.Get().
		Resource("compassconnections").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested compassConnections.
func (c *compassConnections) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("compassconnections").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a compassConnection and creates it.  Returns the server's representation of the compassConnection, and an error, if there is any.
func (c *compassConnections) Create(compassConnection *v1alpha1.CompassConnection) (result *v1alpha1.CompassConnection, err error) {
	result = &v1alpha1.CompassConnection{}
	err = c.client.Post().
		Resource("compassconnections").
		Body(compassConnection).
		Do().
		Into(result)
	return
}

// Update takes the representation of a compassConnection and updates it. Returns the server's representation of the compassConnection, and an error, if there is any.
func (c *compassConnections) Update(compassConnection *v1alpha1.CompassConnection) (result *v1alpha1.CompassConnection, err error) {
	result = &v1alpha1.CompassConnection{}
	err = c.client.Put().
		Resource("compassconnections").
		Name(compassConnection.Name).
		Body(compassConnection).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *compassConnections) UpdateStatus(compassConnection *v1alpha1.CompassConnection) (result *v1alpha1.CompassConnection, err error) {
	result = &v1alpha1.CompassConnection{}
	err = c.client.Put().
		Resource("compassconnections").
		Name(compassConnection.Name).
		SubResource("status").
		Body(compassConnection).
		Do().
		Into(result)
	return
}

// Delete takes name of the compassConnection and deletes it. Returns an error if one occurs.
func (c *compassConnections) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("compassconnections").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *compassConnections) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("compassconnections").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched compassConnection.
func (c *compassConnections) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CompassConnection, err error) {
	result = &v1alpha1.CompassConnection{}
	err = c.client.Patch(pt).
		Resource("compassconnections").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
