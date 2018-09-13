// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeInstallations implements InstallationInterface
type FakeInstallations struct {
	Fake *FakeInstallerV1alpha1
	ns   string
}

var installationsResource = schema.GroupVersionResource{Group: "installer.kyma-project.io", Version: "v1alpha1", Resource: "installations"}

var installationsKind = schema.GroupVersionKind{Group: "installer.kyma-project.io", Version: "v1alpha1", Kind: "Installation"}

// Get takes name of the installation, and returns the corresponding installation object, and an error if there is any.
func (c *FakeInstallations) Get(name string, options v1.GetOptions) (result *v1alpha1.Installation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(installationsResource, c.ns, name), &v1alpha1.Installation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Installation), err
}

// List takes label and field selectors, and returns the list of Installations that match those selectors.
func (c *FakeInstallations) List(opts v1.ListOptions) (result *v1alpha1.InstallationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(installationsResource, installationsKind, c.ns, opts), &v1alpha1.InstallationList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.InstallationList{}
	for _, item := range obj.(*v1alpha1.InstallationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested installations.
func (c *FakeInstallations) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(installationsResource, c.ns, opts))

}

// Create takes the representation of a installation and creates it.  Returns the server's representation of the installation, and an error, if there is any.
func (c *FakeInstallations) Create(installation *v1alpha1.Installation) (result *v1alpha1.Installation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(installationsResource, c.ns, installation), &v1alpha1.Installation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Installation), err
}

// Update takes the representation of a installation and updates it. Returns the server's representation of the installation, and an error, if there is any.
func (c *FakeInstallations) Update(installation *v1alpha1.Installation) (result *v1alpha1.Installation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(installationsResource, c.ns, installation), &v1alpha1.Installation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Installation), err
}

// Delete takes name of the installation and deletes it. Returns an error if one occurs.
func (c *FakeInstallations) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(installationsResource, c.ns, name), &v1alpha1.Installation{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeInstallations) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(installationsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.InstallationList{})
	return err
}

// Patch applies the patch and returns the patched installation.
func (c *FakeInstallations) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Installation, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(installationsResource, c.ns, name, data, subresources...), &v1alpha1.Installation{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Installation), err
}
