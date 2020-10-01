package roles

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resolver struct {
	Client *client.Client
	notifier *resource.NotifierXX
}

func New(k8sClient *client.Client) *Resolver {
	return &Resolver{Client: k8sClient}
}



type Unsubscribe func()

func (r *Resolver) ListInNamespace(ctx context.Context, namespace string, object runtime.Object) error {
	return (*r.Client).List(ctx, object, &client.ListOptions{Namespace: namespace})
}

func (r *Resolver) GetInNamespace(ctx context.Context, namespace, name string, object runtime.Object) error {
	return (*r.Client).Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, object)
}

func (r *Resolver) List(ctx context.Context, object runtime.Object) error {
	return r.ListInNamespace(ctx, "", object)
}

func (r *Resolver) Get(ctx context.Context, name string, object runtime.Object) error {
	return r.GetInNamespace(ctx, "", name, object)
}

func (r *Resolver) Create(ctx context.Context, object runtime.Object) error {
	return (*r.Client).Create(ctx, object)
}

func (r *Resolver) DeleteInNamespace(ctx context.Context, namespace, name string, object runtime.Object) error {
	err := r.GetInNamespace(ctx, namespace, name, object)
	if err != nil {
		return err
	}

	object = object.DeepCopyObject()
	return (*r.Client).Delete(ctx, object)
}

func (r *Resolver) Delete(ctx context.Context, name string, object runtime.Object) error {
	return r.DeleteInNamespace(ctx, "", name, object)
}

func (r *Resolver) Subscribe(handler resource.EventHandlerProviderXX) (Unsubscribe, error) {
	panic("NOPE")
	//listener := resource.NewListenerXX(handler)
	//r.notifier.AddListener(listener)
	//return func() {
	//	r.deleteListener(listener)
	//}, nil
}