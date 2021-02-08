package resource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Create(ctx context.Context, object Object) error
	CreateWithReference(ctx context.Context, parent Object, object Object) error
	Update(ctx context.Context, object Object) error
	Get(ctx context.Context, key ctrlclient.ObjectKey, object Object) error
	ListByLabel(ctx context.Context, namespace string, labels map[string]string, object runtime.Object) error
	DeleteAllBySelector(ctx context.Context, resourceType Object, namespace string, selector apilabels.Selector) error
	Delete(ctx context.Context, resourceType Object) error
	Status() ctrlclient.StatusWriter
}

//go:generate mockery -name=K8sClient -output=automock -outpkg=automock -case=underscore
type K8sClient interface {
	Create(context.Context, runtime.Object, ...ctrlclient.CreateOption) error
	Update(ctx context.Context, obj runtime.Object, opts ...ctrlclient.UpdateOption) error
	Get(ctx context.Context, key ctrlclient.ObjectKey, obj runtime.Object) error
	List(context.Context, runtime.Object, ...ctrlclient.ListOption) error
	DeleteAllOf(context.Context, runtime.Object, ...ctrlclient.DeleteAllOfOption) error
	Status() ctrlclient.StatusWriter
	Delete(ctx context.Context, obj runtime.Object, opts ...ctrlclient.DeleteOption) error
}

type Object interface {
	runtime.Object
	metav1.Object
}

var _ Client = &client{}

type client struct {
	k8sClient K8sClient
	schema    *runtime.Scheme
}

func (c *client) Delete(ctx context.Context, obj Object) error {
	propagationPolicy := metav1.DeletePropagationBackground
	return c.k8sClient.Delete(ctx, obj, &ctrlclient.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}

func New(k8sClient K8sClient, schema *runtime.Scheme) Client {
	return &client{
		k8sClient: k8sClient,
		schema:    schema,
	}
}

func (c *client) Create(ctx context.Context, object Object) error {
	return c.CreateWithReference(ctx, nil, object)
}

func (c *client) CreateWithReference(ctx context.Context, parent, object Object) error {
	if parent != nil {
		if err := controllerutil.SetControllerReference(parent, object, c.schema); err != nil {
			return err
		}
	}

	return c.k8sClient.Create(ctx, object)
}

func (c *client) Update(ctx context.Context, object Object) error {
	return c.k8sClient.Update(ctx, object)
}

func (c *client) Get(ctx context.Context, key ctrlclient.ObjectKey, object Object) error {
	return c.k8sClient.Get(ctx, key, object)
}

func (c *client) ListByLabel(ctx context.Context, namespace string, labels map[string]string, list runtime.Object) error {
	return c.k8sClient.List(ctx, list, &ctrlclient.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(labels),
		Namespace:     namespace,
	})
}

func (c *client) DeleteAllBySelector(ctx context.Context, resourceType Object, namespace string, selector apilabels.Selector) error {
	propagationPolicy := metav1.DeletePropagationForeground

	return c.k8sClient.DeleteAllOf(ctx, resourceType, &ctrlclient.DeleteAllOfOptions{
		ListOptions: ctrlclient.ListOptions{
			LabelSelector: selector,
			Namespace:     namespace,
		},
		DeleteOptions: ctrlclient.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		},
	})
}

func (c *client) Status() ctrlclient.StatusWriter {
	return c.k8sClient.Status()
}
