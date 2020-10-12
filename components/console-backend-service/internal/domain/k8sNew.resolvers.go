package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"k8s.io/api/core/v1"
)

func (r *limitRangeResolver) JSON(ctx context.Context, obj *v1.LimitRange) (gqlschema.JSON, error) {
	return r.k8sNew.JsonField(ctx, obj)
}

func (r *limitRangeItemResolver) Max(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return r.k8sNew.GetLimitRangeResources(&obj.Max)
}

func (r *limitRangeItemResolver) Default(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return r.k8sNew.GetLimitRangeResources(&obj.Default)
}

func (r *limitRangeItemResolver) DefaultRequest(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return r.k8sNew.GetLimitRangeResources(&obj.DefaultRequest)
}

func (r *mutationResolver) UpdateLimitRange(ctx context.Context, namespace string, name string, json gqlschema.JSON) (*v1.LimitRange, error) {
	return r.k8sNew.UpdateLimitRange(ctx, namespace, name, json)
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	return r.k8sNew.LimitRangesQuery(ctx, namespace)
}

func (r *queryResolver) ResourceQuotas(ctx context.Context, namespace string) ([]*v1.ResourceQuota, error) {
	return r.k8sNew.ResourceQuotasQuery(ctx, namespace)
}

func (r *resourceQuotaSpecResolver) Hard(ctx context.Context, obj *v1.ResourceQuotaSpec) (*gqlschema.ResourceQuotaHard, error) {
	return r.k8sNew.GetHardField(obj.Hard)
}

// LimitRange returns gqlschema.LimitRangeResolver implementation.
func (r *Resolver) LimitRange() gqlschema.LimitRangeResolver { return &limitRangeResolver{r} }

// LimitRangeItem returns gqlschema.LimitRangeItemResolver implementation.
func (r *Resolver) LimitRangeItem() gqlschema.LimitRangeItemResolver {
	return &limitRangeItemResolver{r}
}

// ResourceQuotaSpec returns gqlschema.ResourceQuotaSpecResolver implementation.
func (r *Resolver) ResourceQuotaSpec() gqlschema.ResourceQuotaSpecResolver {
	return &resourceQuotaSpecResolver{r}
}

type limitRangeResolver struct{ *Resolver }
type limitRangeItemResolver struct{ *Resolver }
type resourceQuotaSpecResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *resourceQuotaSpecResolver) Limits(ctx context.Context, obj *v1.ResourceQuotaSpec) (*gqlschema.ResourceLimits, error) {
	fmt.Println(obj)
	panic(fmt.Errorf("not implemented"))
}
func (r *resourceQuotaSpecResolver) Requests(ctx context.Context, obj *v1.ResourceQuotaSpec) (*gqlschema.ResourceLimits, error) {
	panic(fmt.Errorf("not implemented"))
}
