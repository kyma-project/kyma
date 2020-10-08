package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

func (r *limitRangeItemResolver) Max(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	fmt.Println(obj.Max.Memory())
	fmt.Println(obj.Max.Cpu())
	a := "ej"
	return &gqlschema.ResourceLimits{
		Memory: &a,
		CPU:    &a,
	}, nil
}

func (r *limitRangeItemResolver) Default(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	//return &gqlschema.ResourceLimits{
	//	Memory: obj.Default.Memory().String(),
	//	CPU:    obj.Default.Cpu().String(),
	//}, nil
	panic("default")
}

func (r *limitRangeItemResolver) DefaultRequest(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	//return &gqlschema.ResourceLimits{
	//	//	Memory: obj.DefaultRequest.Memory().String(),
	//	//	CPU:    obj.DefaultRequest.Cpu().String(),
	//	//}, nil
	panic("defaultrequest")
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	return r.k8sNew.LimitRangesQuery(ctx, namespace)
}

// LimitRangeItem returns gqlschema.LimitRangeItemResolver implementation.
func (r *Resolver) LimitRangeItem() gqlschema.LimitRangeItemResolver {
	return &limitRangeItemResolver{r}
}

type limitRangeItemResolver struct{ *Resolver }
