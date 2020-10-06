package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"k8s.io/api/core/v1"
)

func (r *limitRangeResolver) Limits(ctx context.Context, obj *v1.LimitRange) ([]*gqlschema.LimitRangeItem, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	panic(fmt.Errorf("not implemented"))
}

// LimitRange returns gqlschema.LimitRangeResolver implementation.
func (r *Resolver) LimitRange() gqlschema.LimitRangeResolver { return &limitRangeResolver{r} }

type limitRangeResolver struct{ *Resolver }
