package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *namespaceResolver) IsSystemNamespace(ctx context.Context, obj *model.Namespace) (bool, error) {
	return false, nil
}

func (r *namespaceResolver) Applications(ctx context.Context, obj *model.Namespace) ([]*model.Application, error) {
	mapps, err := r.ApplicationMappings.Client.Namespace(obj.Name).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := &model.ApplicationList{}
	for _, mapp := range mapps.Items {
		err := r.ApplicationConnectorServices.Applications.Get(mapp.GetName(), result.Append())
		if err != nil {
			return nil, err
		}
	}
	return *result, nil
}

// Namespace returns generated.NamespaceResolver implementation.
func (r *Resolver) Namespace() generated.NamespaceResolver { return &namespaceResolver{r} }

type namespaceResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *namespaceResolver) Metadata(ctx context.Context, obj *model.Namespace) (*v1.ObjectMeta, error) {
	panic(fmt.Errorf("not implemented"))
}
