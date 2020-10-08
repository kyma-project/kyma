package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type resourceLimitsItem interface {
	Memory() *resource.Quantity
	Cpu() *resource.Quantity
}

func GetResourceLimits(item resourceLimitsItem) *gqlschema.ResourceLimits {
	mem := item.Memory().String()
	cpu := item.Cpu().String()

	return &gqlschema.ResourceLimits{
		Memory: &mem,
		CPU:    &cpu,
	}
}

func (r *limitRangeItemResolver) Max(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return GetResourceLimits(&obj.Max), nil
}

func (r *limitRangeItemResolver) Default(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return GetResourceLimits(&obj.Default), nil
}

func (r *limitRangeItemResolver) DefaultRequest(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceLimits, error) {
	return GetResourceLimits(&obj.DefaultRequest), nil
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	return r.k8sNew.LimitRangesQuery(ctx, namespace)
}

// LimitRangeItem returns gqlschema.LimitRangeItemResolver implementation.
func (r *Resolver) LimitRangeItem() gqlschema.LimitRangeItemResolver {
	return &limitRangeItemResolver{r}
}

type limitRangeItemResolver struct{ *Resolver }
