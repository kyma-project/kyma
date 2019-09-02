// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha2 "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeApis implements ApiInterface
type FakeApis struct {
	Fake *FakeGatewayV1alpha2
	ns   string
}

var apisResource = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha2", Resource: "apis"}

var apisKind = schema.GroupVersionKind{Group: "gateway.kyma-project.io", Version: "v1alpha2", Kind: "Api"}

// Get takes name of the api, and returns the corresponding api object, and an error if there is any.
func (c *FakeApis) Get(name string, options v1.GetOptions) (result *v1alpha2.Api, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(apisResource, c.ns, name), &v1alpha2.Api{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Api), err
}

// List takes label and field selectors, and returns the list of Apis that match those selectors.
func (c *FakeApis) List(opts v1.ListOptions) (result *v1alpha2.ApiList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(apisResource, apisKind, c.ns, opts), &v1alpha2.ApiList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.ApiList{ListMeta: obj.(*v1alpha2.ApiList).ListMeta}
	for _, item := range obj.(*v1alpha2.ApiList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested apis.
func (c *FakeApis) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(apisResource, c.ns, opts))

}

// Create takes the representation of a api and creates it.  Returns the server's representation of the api, and an error, if there is any.
func (c *FakeApis) Create(api *v1alpha2.Api) (result *v1alpha2.Api, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(apisResource, c.ns, api), &v1alpha2.Api{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Api), err
}

// Update takes the representation of a api and updates it. Returns the server's representation of the api, and an error, if there is any.
func (c *FakeApis) Update(api *v1alpha2.Api) (result *v1alpha2.Api, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(apisResource, c.ns, api), &v1alpha2.Api{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Api), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeApis) UpdateStatus(api *v1alpha2.Api) (*v1alpha2.Api, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(apisResource, "status", c.ns, api), &v1alpha2.Api{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Api), err
}

// Delete takes name of the api and deletes it. Returns an error if one occurs.
func (c *FakeApis) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(apisResource, c.ns, name), &v1alpha2.Api{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeApis) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(apisResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha2.ApiList{})
	return err
}

// Patch applies the patch and returns the patched api.
func (c *FakeApis) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha2.Api, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(apisResource, c.ns, name, pt, data, subresources...), &v1alpha2.Api{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.Api), err
}
