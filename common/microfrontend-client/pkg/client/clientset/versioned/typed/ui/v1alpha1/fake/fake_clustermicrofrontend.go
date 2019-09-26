// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterMicroFrontends implements ClusterMicroFrontendInterface
type FakeClusterMicroFrontends struct {
	Fake *FakeUiV1alpha1
}

var clustermicrofrontendsResource = schema.GroupVersionResource{Group: "ui.kyma-project.io", Version: "v1alpha1", Resource: "clustermicrofrontends"}

var clustermicrofrontendsKind = schema.GroupVersionKind{Group: "ui.kyma-project.io", Version: "v1alpha1", Kind: "ClusterMicroFrontend"}

// Get takes name of the clusterMicroFrontend, and returns the corresponding clusterMicroFrontend object, and an error if there is any.
func (c *FakeClusterMicroFrontends) Get(name string, options v1.GetOptions) (result *v1alpha1.ClusterMicroFrontend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(clustermicrofrontendsResource, name), &v1alpha1.ClusterMicroFrontend{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterMicroFrontend), err
}

// List takes label and field selectors, and returns the list of ClusterMicroFrontends that match those selectors.
func (c *FakeClusterMicroFrontends) List(opts v1.ListOptions) (result *v1alpha1.ClusterMicroFrontendList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(clustermicrofrontendsResource, clustermicrofrontendsKind, opts), &v1alpha1.ClusterMicroFrontendList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ClusterMicroFrontendList{ListMeta: obj.(*v1alpha1.ClusterMicroFrontendList).ListMeta}
	for _, item := range obj.(*v1alpha1.ClusterMicroFrontendList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterMicroFrontends.
func (c *FakeClusterMicroFrontends) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(clustermicrofrontendsResource, opts))
}

// Create takes the representation of a clusterMicroFrontend and creates it.  Returns the server's representation of the clusterMicroFrontend, and an error, if there is any.
func (c *FakeClusterMicroFrontends) Create(clusterMicroFrontend *v1alpha1.ClusterMicroFrontend) (result *v1alpha1.ClusterMicroFrontend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(clustermicrofrontendsResource, clusterMicroFrontend), &v1alpha1.ClusterMicroFrontend{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterMicroFrontend), err
}

// Update takes the representation of a clusterMicroFrontend and updates it. Returns the server's representation of the clusterMicroFrontend, and an error, if there is any.
func (c *FakeClusterMicroFrontends) Update(clusterMicroFrontend *v1alpha1.ClusterMicroFrontend) (result *v1alpha1.ClusterMicroFrontend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(clustermicrofrontendsResource, clusterMicroFrontend), &v1alpha1.ClusterMicroFrontend{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterMicroFrontend), err
}

// Delete takes name of the clusterMicroFrontend and deletes it. Returns an error if one occurs.
func (c *FakeClusterMicroFrontends) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(clustermicrofrontendsResource, name), &v1alpha1.ClusterMicroFrontend{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterMicroFrontends) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(clustermicrofrontendsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ClusterMicroFrontendList{})
	return err
}

// Patch applies the patch and returns the patched clusterMicroFrontend.
func (c *FakeClusterMicroFrontends) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterMicroFrontend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(clustermicrofrontendsResource, name, pt, data, subresources...), &v1alpha1.ClusterMicroFrontend{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterMicroFrontend), err
}
