package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"k8s.io/api/core/v1"
)

func (r *limitRangeItemResolver) Type(ctx context.Context, obj *v1.LimitRangeItem) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) LimitRanges(ctx context.Context, namespace string) ([]*v1.LimitRange, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *resourceListResolver) CPU(ctx context.Context, obj *v1.ResourceList) (*v1.ResourceName, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *resourceListResolver) Memory(ctx context.Context, obj *v1.ResourceList) (*v1.ResourceName, error) {
	panic(fmt.Errorf("not implemented"))
}

// LimitRangeItem returns gqlschema.LimitRangeItemResolver implementation.
func (r *Resolver) LimitRangeItem() gqlschema.LimitRangeItemResolver {
	return &limitRangeItemResolver{r}
}

// ResourceList returns gqlschema.ResourceListResolver implementation.
func (r *Resolver) ResourceList() gqlschema.ResourceListResolver { return &resourceListResolver{r} }

type limitRangeItemResolver struct{ *Resolver }
type resourceListResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *limitRangeItemResolver) Max(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceType, error) {
	panic(fmt.Errorf("not implemented"))
}
func (r *limitRangeItemResolver) Default(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceType, error) {
	panic(fmt.Errorf("not implemented"))
}
func (r *limitRangeItemResolver) DefaultRequest(ctx context.Context, obj *v1.LimitRangeItem) (*gqlschema.ResourceType, error) {
	panic(fmt.Errorf("not implemented"))
}
