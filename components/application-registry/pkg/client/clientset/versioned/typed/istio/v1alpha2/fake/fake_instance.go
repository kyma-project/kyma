// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeInstances implements InstanceInterface
type FakeInstances struct {
	Fake *FakeIstioV1alpha2
	ns   string
}

var instancesResource = schema.GroupVersionResource{Group: "istio", Version: "v1alpha2", Resource: "instances"}

var instancesKind = schema.GroupVersionKind{Group: "istio", Version: "v1alpha2", Kind: "Instance"}

// Get takes name of the instance, and returns the corresponding instance object, and an error if there is any.
func (c *FakeInstances) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.Instance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(instancesResource, c.ns, name), &v1alpha2.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Instance), err
}

// List takes label and field selectors, and returns the list of Instances that match those selectors.
func (c *FakeInstances) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.InstanceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(instancesResource, instancesKind, c.ns, opts), &v1alpha2.InstanceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.InstanceList{ListMeta: obj.(*v1alpha2.InstanceList).ListMeta}
	for _, item := range obj.(*v1alpha2.InstanceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested instances.
func (c *FakeInstances) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(instancesResource, c.ns, opts))

}

// Create takes the representation of a instance and creates it.  Returns the server's representation of the instance, and an error, if there is any.
func (c *FakeInstances) Create(ctx context.Context, instance *v1alpha2.Instance, opts v1.CreateOptions) (result *v1alpha2.Instance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(instancesResource, c.ns, instance), &v1alpha2.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Instance), err
}

// Update takes the representation of a instance and updates it. Returns the server's representation of the instance, and an error, if there is any.
func (c *FakeInstances) Update(ctx context.Context, instance *v1alpha2.Instance, opts v1.UpdateOptions) (result *v1alpha2.Instance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(instancesResource, c.ns, instance), &v1alpha2.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Instance), err
}

// Delete takes name of the instance and deletes it. Returns an error if one occurs.
func (c *FakeInstances) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(instancesResource, c.ns, name), &v1alpha2.Instance{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeInstances) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(instancesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.InstanceList{})
	return err
}

// Patch applies the patch and returns the patched instance.
func (c *FakeInstances) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Instance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(instancesResource, c.ns, name, pt, data, subresources...), &v1alpha2.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Instance), err
}
