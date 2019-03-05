// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceBindingUsages implements ServiceBindingUsageInterface
type FakeServiceBindingUsages struct {
	Fake *FakeServicecatalogV1alpha1
	ns   string
}

var servicebindingusagesResource = schema.GroupVersionResource{Group: "servicecatalog.kyma-project.io", Version: "v1alpha1", Resource: "servicebindingusages"}

var servicebindingusagesKind = schema.GroupVersionKind{Group: "servicecatalog.kyma-project.io", Version: "v1alpha1", Kind: "ServiceBindingUsage"}

// Get takes name of the serviceBindingUsage, and returns the corresponding serviceBindingUsage object, and an error if there is any.
func (c *FakeServiceBindingUsages) Get(name string, options v1.GetOptions) (result *v1alpha1.ServiceBindingUsage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicebindingusagesResource, c.ns, name), &v1alpha1.ServiceBindingUsage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceBindingUsage), err
}

// List takes label and field selectors, and returns the list of ServiceBindingUsages that match those selectors.
func (c *FakeServiceBindingUsages) List(opts v1.ListOptions) (result *v1alpha1.ServiceBindingUsageList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicebindingusagesResource, servicebindingusagesKind, c.ns, opts), &v1alpha1.ServiceBindingUsageList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ServiceBindingUsageList{}
	for _, item := range obj.(*v1alpha1.ServiceBindingUsageList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceBindingUsages.
func (c *FakeServiceBindingUsages) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicebindingusagesResource, c.ns, opts))

}

// Create takes the representation of a serviceBindingUsage and creates it.  Returns the server's representation of the serviceBindingUsage, and an error, if there is any.
func (c *FakeServiceBindingUsages) Create(serviceBindingUsage *v1alpha1.ServiceBindingUsage) (result *v1alpha1.ServiceBindingUsage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicebindingusagesResource, c.ns, serviceBindingUsage), &v1alpha1.ServiceBindingUsage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceBindingUsage), err
}

// Update takes the representation of a serviceBindingUsage and updates it. Returns the server's representation of the serviceBindingUsage, and an error, if there is any.
func (c *FakeServiceBindingUsages) Update(serviceBindingUsage *v1alpha1.ServiceBindingUsage) (result *v1alpha1.ServiceBindingUsage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicebindingusagesResource, c.ns, serviceBindingUsage), &v1alpha1.ServiceBindingUsage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceBindingUsage), err
}

// Delete takes name of the serviceBindingUsage and deletes it. Returns an error if one occurs.
func (c *FakeServiceBindingUsages) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicebindingusagesResource, c.ns, name), &v1alpha1.ServiceBindingUsage{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceBindingUsages) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicebindingusagesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ServiceBindingUsageList{})
	return err
}

// Patch applies the patch and returns the patched serviceBindingUsage.
func (c *FakeServiceBindingUsages) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceBindingUsage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicebindingusagesResource, c.ns, name, data, subresources...), &v1alpha1.ServiceBindingUsage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceBindingUsage), err
}
