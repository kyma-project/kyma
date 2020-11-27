package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	// "fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

func (r *queryResolver) EventSubscription(ctx context.Context, name string, namespace string) (*gqlschema.EventSubscription, error) {
	return r.bebEventing.EventSubscriptionQuery(ctx, namespace, name)
}
