package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *queryResolver) Namespaces(ctx context.Context) ([]*model.Namespace, error) {
	nss := model.NamespaceList{}
	err := r.namespaces.List(&nss)
	return nss, err
}

func (r *queryResolver) Pods(ctx context.Context, namespace string) ([]*model.Pod, error) {
	pods := model.PodList{}
	err := r.pods.ListInNamespace(namespace, &pods)
	return pods, err
}

func (r *queryResolver) Applications(ctx context.Context) ([]*model.Application, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
