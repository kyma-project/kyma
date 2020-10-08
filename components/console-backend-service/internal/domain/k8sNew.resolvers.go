package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

func (r *limitRangeResolver) JSON(ctx context.Context, obj *v1.LimitRange) (gqlschema.JSON, error) {
	return r.k8sNew.JsonField(ctx, obj)
}

func (r *limitRangeItemResolver) Max(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return r.k8sNew.GetResourceLimits(&obj.Max)
}

func (r *limitRangeItemResolver) Default(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return r.k8sNew.GetResourceLimits(&obj.Default)
}

func (r *limitRangeItemResolver) DefaultRequest(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return r.k8sNew.GetResourceLimits(&obj.DefaultRequest)
}

func (r *mutationResolver) UpdateLimitRange(ctx context.Context, namespace string, name string, generation int, json gqlschema.JSON) (*v1.LimitRange, error) {
	return r.k8sNew.UpdateLimitRange(ctx, namespace, name, int64(generation), json)
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	return r.k8sNew.LimitRangesQuery(ctx, namespace)
}

// LimitRange returns gqlschema.LimitRangeResolver implementation.
func (r *Resolver) LimitRange() gqlschema.LimitRangeResolver { return &limitRangeResolver{r} }

// LimitRangeItem returns gqlschema.LimitRangeItemResolver implementation.
func (r *Resolver) LimitRangeItem() gqlschema.LimitRangeItemResolver {
	return &limitRangeItemResolver{r}
}

type limitRangeResolver struct{ *Resolver }
type limitRangeItemResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
