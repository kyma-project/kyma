package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *mutationResolver) CreateNamespace(ctx context.Context, in *model.NamespaceInput) (*model.Namespace, error) {
	ns := &model.Namespace{}
	err := r.CoreServices.Namespaces.Create(in, ns)
	return ns, err
}

func (r *namespaceResolver) IsSystemNamespace(ctx context.Context, obj *model.Namespace) (bool, error) {
	return strings.HasSuffix(obj.Name, "-system"), nil
}

func (r *namespaceResolver) ApplicationMappings(ctx context.Context, obj *model.Namespace) ([]*model.ApplicationMapping, error) {
	return r.Query().Mappings(ctx, obj.Name)
}

func (r *queryResolver) Namespaces(ctx context.Context) ([]*model.Namespace, error) {
	nss := model.NamespaceList{}
	err := r.CoreServices.Namespaces.List(&nss)
	return nss, err
}

func (r *queryResolver) Namespace(ctx context.Context, name string) (*model.Namespace, error) {
	ns := &model.Namespace{}
	err := r.CoreServices.Namespaces.Get(name, ns)
	return ns, err
}

func (r *queryResolver) Pods(ctx context.Context, namespace string) ([]*model.Pod, error) {
	pods := model.PodList{}
	err := r.CoreServices.Pods.ListInNamespace(namespace, &pods)
	return pods, err
}

func (r *queryResolver) Pod(ctx context.Context, namespace string, name string) (*model.Pod, error) {
	ns := &model.Pod{}
	err := r.CoreServices.Pods.GetInNamespace(name, namespace, ns)
	return ns, err
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Namespace returns generated.NamespaceResolver implementation.
func (r *Resolver) Namespace() generated.NamespaceResolver { return &namespaceResolver{r} }

type mutationResolver struct{ *Resolver }
type namespaceResolver struct{ *Resolver }
