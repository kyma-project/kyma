// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCompassConnections implements CompassConnectionInterface
type FakeCompassConnections struct {
	Fake *FakeCompassV1alpha1
}

var compassconnectionsResource = schema.GroupVersionResource{Group: "compass.kyma-project.io", Version: "v1alpha1", Resource: "compassconnections"}

var compassconnectionsKind = schema.GroupVersionKind{Group: "compass.kyma-project.io", Version: "v1alpha1", Kind: "CompassConnection"}

// Get takes name of the compassConnection, and returns the corresponding compassConnection object, and an error if there is any.
func (c *FakeCompassConnections) Get(name string, options v1.GetOptions) (result *v1alpha1.CompassConnection, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(compassconnectionsResource, name), &v1alpha1.CompassConnection{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CompassConnection), err
}

// List takes label and field selectors, and returns the list of CompassConnections that match those selectors.
func (c *FakeCompassConnections) List(opts v1.ListOptions) (result *v1alpha1.CompassConnectionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(compassconnectionsResource, compassconnectionsKind, opts), &v1alpha1.CompassConnectionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CompassConnectionList{ListMeta: obj.(*v1alpha1.CompassConnectionList).ListMeta}
	for _, item := range obj.(*v1alpha1.CompassConnectionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested compassConnections.
func (c *FakeCompassConnections) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(compassconnectionsResource, opts))
}

// Create takes the representation of a compassConnection and creates it.  Returns the server's representation of the compassConnection, and an error, if there is any.
func (c *FakeCompassConnections) Create(compassConnection *v1alpha1.CompassConnection) (result *v1alpha1.CompassConnection, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(compassconnectionsResource, compassConnection), &v1alpha1.CompassConnection{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CompassConnection), err
}

// Update takes the representation of a compassConnection and updates it. Returns the server's representation of the compassConnection, and an error, if there is any.
func (c *FakeCompassConnections) Update(compassConnection *v1alpha1.CompassConnection) (result *v1alpha1.CompassConnection, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(compassconnectionsResource, compassConnection), &v1alpha1.CompassConnection{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CompassConnection), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCompassConnections) UpdateStatus(compassConnection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(compassconnectionsResource, "status", compassConnection), &v1alpha1.CompassConnection{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CompassConnection), err
}

// Delete takes name of the compassConnection and deletes it. Returns an error if one occurs.
func (c *FakeCompassConnections) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(compassconnectionsResource, name), &v1alpha1.CompassConnection{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCompassConnections) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(compassconnectionsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CompassConnectionList{})
	return err
}

// Patch applies the patch and returns the patched compassConnection.
func (c *FakeCompassConnections) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CompassConnection, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(compassconnectionsResource, name, pt, data, subresources...), &v1alpha1.CompassConnection{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CompassConnection), err
}
