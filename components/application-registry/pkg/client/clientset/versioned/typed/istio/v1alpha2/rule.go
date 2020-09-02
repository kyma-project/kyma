// Code generated by client-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	"time"

	v1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	scheme "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// RulesGetter has a method to return a RuleInterface.
// A group's client should implement this interface.
type RulesGetter interface {
	Rules(namespace string) RuleInterface
}

// RuleInterface has methods to work with Rule resources.
type RuleInterface interface {
	Create(ctx context.Context, rule *v1alpha2.Rule, opts v1.CreateOptions) (*v1alpha2.Rule, error)
	Update(ctx context.Context, rule *v1alpha2.Rule, opts v1.UpdateOptions) (*v1alpha2.Rule, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.Rule, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.RuleList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Rule, err error)
	RuleExpansion
}

// rules implements RuleInterface
type rules struct {
	client rest.Interface
	ns     string
}

// newRules returns a Rules
func newRules(c *IstioV1alpha2Client, namespace string) *rules {
	return &rules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the rule, and returns the corresponding rule object, and an error if there is any.
func (c *rules) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Rules that match those selectors.
func (c *rules) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.RuleList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.RuleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested rules.
func (c *rules) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a rule and creates it.  Returns the server's representation of the rule, and an error, if there is any.
func (c *rules) Create(ctx context.Context, rule *v1alpha2.Rule, opts v1.CreateOptions) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(rule).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a rule and updates it. Returns the server's representation of the rule, and an error, if there is any.
func (c *rules) Update(ctx context.Context, rule *v1alpha2.Rule, opts v1.UpdateOptions) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("rules").
		Name(rule.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(rule).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the rule and deletes it. Returns an error if one occurs.
func (c *rules) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *rules) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched rule.
func (c *rules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.Rule, err error) {
	result = &v1alpha2.Rule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
