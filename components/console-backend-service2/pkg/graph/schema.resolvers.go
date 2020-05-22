package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *queryResolver) Core(ctx context.Context) (*model.CoreQuery, error) {
	return &model.CoreQuery{}, nil
}

func (r *queryResolver) UI(ctx context.Context) (*model.UiQuery, error) {
	return &model.UiQuery{}, nil
}

func (r *queryResolver) ApplicationConnector(ctx context.Context) (*model.ApplicationConnectorQuery, error) {
	return &model.ApplicationConnectorQuery{}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
