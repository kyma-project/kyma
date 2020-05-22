package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *uiQueryResolver) BackendModules(ctx context.Context, obj *model.UiQuery) ([]*model.BackendModule, error) {
	list := model.BackendModuleList{}
	err := r.backendModules.List(&list)
	return list, err
}

// UiQuery returns generated.UiQueryResolver implementation.
func (r *Resolver) UiQuery() generated.UiQueryResolver { return &uiQueryResolver{r} }

type uiQueryResolver struct{ *Resolver }
