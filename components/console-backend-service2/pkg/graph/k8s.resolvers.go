package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
