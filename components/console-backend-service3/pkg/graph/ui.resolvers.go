package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
)

func (r *queryResolver) BackendModules(ctx context.Context) ([]*model.BackendModule, error) {
	return r.ui.Resolver.BackendModulesQuery(ctx)
}
