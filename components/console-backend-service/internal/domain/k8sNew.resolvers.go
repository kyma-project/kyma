package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

func (r *limitRangeResolver) JSON(ctx context.Context, obj *v1.LimitRange) (gqlschema.JSON, error) {
	return r.k8sNew.LimitRangeJSONfield(ctx, obj)
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

func (r *mutationResolver) CreateLimitRange(ctx context.Context, namespace string, name string, limitRange gqlschema.LimitRangeInput) (*v1.LimitRange, error) {
	return r.k8sNew.CreateLimitRange(ctx, namespace, name, limitRange)
}

func (r *mutationResolver) UpdateLimitRange(ctx context.Context, namespace string, name string, json gqlschema.JSON) (*v1.LimitRange, error) {
	return r.k8sNew.UpdateLimitRange(ctx, namespace, name, json)
}

func (r *mutationResolver) CreateResourceQuota(ctx context.Context, namespace string, name string, resourceQuota gqlschema.ResourceQuotaInput) (*v1.ResourceQuota, error) {
	return r.k8sNew.CreateResourceQuota(ctx, namespace, name, resourceQuota)
}

func (r *mutationResolver) UpdateResourceQuota(ctx context.Context, namespace string, name string, json gqlschema.JSON) (*v1.ResourceQuota, error) {
	return r.k8sNew.UpdateResourceQuota(ctx, namespace, name, json)
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	return r.k8sNew.LimitRangesQuery(ctx, namespace)
}

func (r *queryResolver) ResourceQuotas(ctx context.Context, namespace string) ([]*v1.ResourceQuota, error) {
	return r.k8sNew.ResourceQuotasQuery(ctx, namespace)
}

func (r *resourceQuotaResolver) JSON(ctx context.Context, obj *v1.ResourceQuota) (gqlschema.JSON, error) {
	return r.k8sNew.ResourceQuotaJSONfield(ctx, obj)
}

func (r *resourceQuotaSpecResolver) Hard(ctx context.Context, obj *v1.ResourceQuotaSpec) (*gqlschema.ResourceQuotaHard, error) {
	return r.k8sNew.GetHardField(obj)
}

// LimitRange returns gqlschema.LimitRangeResolver implementation.
func (r *Resolver) LimitRange() gqlschema.LimitRangeResolver { return &limitRangeResolver{r} }

// LimitRangeItem returns gqlschema.LimitRangeItemResolver implementation.
func (r *Resolver) LimitRangeItem() gqlschema.LimitRangeItemResolver {
	return &limitRangeItemResolver{r}
}

// ResourceQuota returns gqlschema.ResourceQuotaResolver implementation.
func (r *Resolver) ResourceQuota() gqlschema.ResourceQuotaResolver { return &resourceQuotaResolver{r} }

// ResourceQuotaSpec returns gqlschema.ResourceQuotaSpecResolver implementation.
func (r *Resolver) ResourceQuotaSpec() gqlschema.ResourceQuotaSpecResolver {
	return &resourceQuotaSpecResolver{r}
}

type limitRangeResolver struct{ *Resolver }
type limitRangeItemResolver struct{ *Resolver }
type resourceQuotaResolver struct{ *Resolver }
type resourceQuotaSpecResolver struct{ *Resolver }
