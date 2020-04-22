package resource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Resource interface {
	Create(context.Context, Object, Object) error
	ListByLabel(context.Context, string, map[string]string, runtime.Object) error
	DeleteAllBySelector(context.Context, Object, string, apilabels.Selector) error
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Create(context.Context, runtime.Object, ...client.CreateOption) error
	List(context.Context, runtime.Object, ...client.ListOption) error
	DeleteAllOf(context.Context, runtime.Object, ...client.DeleteAllOfOption) error
}

type Object interface {
	runtime.Object
	metav1.Object
}

var _ Resource = &resourceSvc{}

type resourceSvc struct {
	client Client
	schema *runtime.Scheme
}

func New(client Client, schema *runtime.Scheme) *resourceSvc {
	return &resourceSvc{
		client: client,
		schema: schema,
	}
}

func (r *resourceSvc) Create(ctx context.Context, parent, object Object) error {
	if err := controllerutil.SetControllerReference(parent, object, r.schema); err != nil {
		return err
	}

	return r.client.Create(ctx, object)
}

func (r *resourceSvc) ListByLabel(ctx context.Context, namespace string, labels map[string]string, list runtime.Object) error {
	return r.client.List(ctx, list, &client.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(labels),
		Namespace:     namespace,
	})
}

func (r *resourceSvc) DeleteAllBySelector(ctx context.Context, resourceType Object, namespace string, selector apilabels.Selector) error {
	propagationPolicy := metav1.DeletePropagationBackground

	return r.client.DeleteAllOf(ctx, resourceType, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: selector,
			Namespace:     namespace,
		},
		DeleteOptions: client.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		},
	})
}
