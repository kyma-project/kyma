package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
)

func (r *queryResolver) APIRules(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]*model.APIRule, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) APIRule(ctx context.Context, name string, namespace string) (*model.APIRule, error) {
	panic(fmt.Errorf("not implemented"))
}
