package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"k8s.io/api/rbac/v1alpha1"
)

func (r *queryResolver) Roles(ctx context.Context, namespace string) ([]*v1alpha1.Role, error) {
	return r.roles.RolesQuery(ctx, namespace)
}
