package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

func (r *applicationMappingResolver) Application(ctx context.Context, obj *model.ApplicationMapping) (*model.Application, error) {
	return r.Query().Application(ctx, obj.Name)
}

func (r *queryResolver) Applications(ctx context.Context) ([]*model.Application, error) {
	list := model.ApplicationList{}
	err := r.ApplicationConnectorServices.Applications.List(&list)
	return list, err
}

func (r *queryResolver) Application(ctx context.Context, name string) (*model.Application, error) {
	result := &model.Application{}
	err := r.ApplicationConnectorServices.Applications.Get(name, result)
	if err == resource.NotFound {
		return nil, nil
	}
	return result, err
}

func (r *queryResolver) Mappings(ctx context.Context, namespace string) ([]*model.ApplicationMapping, error) {
	mappings := model.ApplicationMappingList{}
	err := r.ApplicationMappings.ListInNamespace(namespace, &mappings)
	return mappings, err
}

// ApplicationMapping returns generated.ApplicationMappingResolver implementation.
func (r *Resolver) ApplicationMapping() generated.ApplicationMappingResolver {
	return &applicationMappingResolver{r}
}

type applicationMappingResolver struct{ *Resolver }
