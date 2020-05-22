package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *applicationConnectorQueryResolver) Applications(ctx context.Context, obj *model.ApplicationConnectorQuery) ([]*model.Application, error) {
	list := model.ApplicationList{}
	err := r.ApplicationConnectorServices.Applications.List(&list)
	return list, err
}

// ApplicationConnectorQuery returns generated.ApplicationConnectorQueryResolver implementation.
func (r *Resolver) ApplicationConnectorQuery() generated.ApplicationConnectorQueryResolver {
	return &applicationConnectorQueryResolver{r}
}

type applicationConnectorQueryResolver struct{ *Resolver }
