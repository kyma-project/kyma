// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeHTTPSources implements HTTPSourceInterface
type FakeHTTPSources struct {
	Fake *FakeSourcesV1alpha1
	ns   string
}

var httpsourcesResource = schema.GroupVersionResource{Group: "sources.kyma-project.io", Version: "v1alpha1", Resource: "httpsources"}

var httpsourcesKind = schema.GroupVersionKind{Group: "sources.kyma-project.io", Version: "v1alpha1", Kind: "HTTPSource"}

// Get takes name of the hTTPSource, and returns the corresponding hTTPSource object, and an error if there is any.
func (c *FakeHTTPSources) Get(name string, options v1.GetOptions) (result *v1alpha1.HTTPSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(httpsourcesResource, c.ns, name), &v1alpha1.HTTPSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HTTPSource), err
}

// List takes label and field selectors, and returns the list of HTTPSources that match those selectors.
func (c *FakeHTTPSources) List(opts v1.ListOptions) (result *v1alpha1.HTTPSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(httpsourcesResource, httpsourcesKind, c.ns, opts), &v1alpha1.HTTPSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.HTTPSourceList{ListMeta: obj.(*v1alpha1.HTTPSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.HTTPSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested hTTPSources.
func (c *FakeHTTPSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(httpsourcesResource, c.ns, opts))

}

// Create takes the representation of a hTTPSource and creates it.  Returns the server's representation of the hTTPSource, and an error, if there is any.
func (c *FakeHTTPSources) Create(hTTPSource *v1alpha1.HTTPSource) (result *v1alpha1.HTTPSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(httpsourcesResource, c.ns, hTTPSource), &v1alpha1.HTTPSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HTTPSource), err
}

// Update takes the representation of a hTTPSource and updates it. Returns the server's representation of the hTTPSource, and an error, if there is any.
func (c *FakeHTTPSources) Update(hTTPSource *v1alpha1.HTTPSource) (result *v1alpha1.HTTPSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(httpsourcesResource, c.ns, hTTPSource), &v1alpha1.HTTPSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HTTPSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeHTTPSources) UpdateStatus(hTTPSource *v1alpha1.HTTPSource) (*v1alpha1.HTTPSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(httpsourcesResource, "status", c.ns, hTTPSource), &v1alpha1.HTTPSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HTTPSource), err
}

// Delete takes name of the hTTPSource and deletes it. Returns an error if one occurs.
func (c *FakeHTTPSources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(httpsourcesResource, c.ns, name), &v1alpha1.HTTPSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHTTPSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(httpsourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.HTTPSourceList{})
	return err
}

// Patch applies the patch and returns the patched hTTPSource.
func (c *FakeHTTPSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.HTTPSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(httpsourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.HTTPSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HTTPSource), err
}
