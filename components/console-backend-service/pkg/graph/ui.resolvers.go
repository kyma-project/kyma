package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
)

func (r *queryResolver) BackendModules(ctx context.Context) ([]*model.BackendModule, error) {
	return r.ui.BackendModulesQuery(ctx)
}

func (r *queryResolver) MicroFrontends(ctx context.Context, namespace string) ([]*model.MicroFrontend, error) {
	return r.ui.MicroFrontendsQuery(ctx, namespace)
}

func (r *queryResolver) ClusterMicroFrontends(ctx context.Context) ([]*model.ClusterMicroFrontend, error) {
	return r.ui.ClusterMicroFrontendsQuery(ctx)
}

func (r *subscriptionResolver) BackendModules(ctx context.Context) (<-chan *model.BackendModuleEvent, error) {
	return r.ui.BackendModulesEvents(ctx)
}
