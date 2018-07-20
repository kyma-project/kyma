// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeUsageKinds implements UsageKindInterface
type FakeUsageKinds struct {
	Fake *FakeServicecatalogV1alpha1
}

var usagekindsResource = schema.GroupVersionResource{Group: "servicecatalog.kyma.cx", Version: "v1alpha1", Resource: "usagekinds"}

var usagekindsKind = schema.GroupVersionKind{Group: "servicecatalog.kyma.cx", Version: "v1alpha1", Kind: "UsageKind"}

// Get takes name of the usageKind, and returns the corresponding usageKind object, and an error if there is any.
func (c *FakeUsageKinds) Get(name string, options v1.GetOptions) (result *v1alpha1.UsageKind, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(usagekindsResource, name), &v1alpha1.UsageKind{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.UsageKind), err
}

// List takes label and field selectors, and returns the list of UsageKinds that match those selectors.
func (c *FakeUsageKinds) List(opts v1.ListOptions) (result *v1alpha1.UsageKindList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(usagekindsResource, usagekindsKind, opts), &v1alpha1.UsageKindList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.UsageKindList{}
	for _, item := range obj.(*v1alpha1.UsageKindList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested usageKinds.
func (c *FakeUsageKinds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(usagekindsResource, opts))
}

// Create takes the representation of a usageKind and creates it.  Returns the server's representation of the usageKind, and an error, if there is any.
func (c *FakeUsageKinds) Create(usageKind *v1alpha1.UsageKind) (result *v1alpha1.UsageKind, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(usagekindsResource, usageKind), &v1alpha1.UsageKind{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.UsageKind), err
}

// Update takes the representation of a usageKind and updates it. Returns the server's representation of the usageKind, and an error, if there is any.
func (c *FakeUsageKinds) Update(usageKind *v1alpha1.UsageKind) (result *v1alpha1.UsageKind, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(usagekindsResource, usageKind), &v1alpha1.UsageKind{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.UsageKind), err
}

// Delete takes name of the usageKind and deletes it. Returns an error if one occurs.
func (c *FakeUsageKinds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(usagekindsResource, name), &v1alpha1.UsageKind{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeUsageKinds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(usagekindsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.UsageKindList{})
	return err
}

// Patch applies the patch and returns the patched usageKind.
func (c *FakeUsageKinds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.UsageKind, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(usagekindsResource, name, data, subresources...), &v1alpha1.UsageKind{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.UsageKind), err
}
