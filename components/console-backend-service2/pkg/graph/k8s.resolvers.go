package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *namespaceResolver) IsSystemNamespace(ctx context.Context, obj *model.Namespace) (bool, error) {
	return false, nil
}

func (r *namespaceResolver) Applications(ctx context.Context, obj *model.Namespace) ([]*model.Application, error) {
	mappings := model.ApplicationMappingList{}
	err := r.ApplicationMappings.ListInNamespace(obj.Name, &mappings)
	if err != nil {
		return nil, err
	}

	result := &model.ApplicationList{}
	for _, mapping := range mappings {
		err := r.ApplicationConnectorServices.Applications.Get(mapping.GetName(), result.Append())
		if err != nil {
			return nil, err
		}
	}
	return *result, nil
}

// Namespace returns generated.NamespaceResolver implementation.
func (r *Resolver) Namespace() generated.NamespaceResolver { return &namespaceResolver{r} }

type namespaceResolver struct{ *Resolver }
