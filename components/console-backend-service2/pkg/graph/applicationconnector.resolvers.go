package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

func (r *applicationConnectorQueryResolver) Applications(ctx context.Context, obj *model.ApplicationConnectorQuery) ([]*model.Application, error) {
	list := model.ApplicationList{}
	err := r.ApplicationConnectorServices.Applications.List(&list)
	return list, err
}

func (r *applicationConnectorQueryResolver) Application(ctx context.Context, obj *model.ApplicationConnectorQuery, name string) (*model.Application, error) {
	result := &model.Application{}
	err := r.ApplicationConnectorServices.Applications.Get(name, result)
	if err == resource.NotFound {
		return nil, nil
	}
	return result, err
}

func (r *applicationConnectorQueryResolver) Mappings(ctx context.Context, obj *model.ApplicationConnectorQuery, namespace string) ([]*model.ApplicationMapping, error) {
	mappings := model.ApplicationMappingList{}
	err := r.ApplicationMappings.ListInNamespace(namespace, &mappings)
	return mappings, err
}

// ApplicationConnectorQuery returns generated.ApplicationConnectorQueryResolver implementation.
func (r *Resolver) ApplicationConnectorQuery() generated.ApplicationConnectorQueryResolver {
	return &applicationConnectorQueryResolver{r}
}

type applicationConnectorQueryResolver struct{ *Resolver }
