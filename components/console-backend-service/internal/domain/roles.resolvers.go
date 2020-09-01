package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
)

func (r *mutationResolver) DeleteRoleBinding(ctx context.Context, namespace string, name string) (*v1.RoleBinding, error) {
	return r.roles.DeleteRoleBinding(ctx, namespace, name)
}

func (r *mutationResolver) DeleteClusterRoleBinding(ctx context.Context, name string) (*v1.ClusterRoleBinding, error) {
	return r.roles.DeleteClusterRoleBinding(ctx, name)
}

func (r *queryResolver) Roles(ctx context.Context, namespace string) ([]*v1.Role, error) {
	return r.roles.RolesQuery(ctx, namespace)
}

func (r *queryResolver) Role(ctx context.Context, namespace string, name string) (*v1.Role, error) {
	return r.roles.RoleQuery(ctx, name, namespace)
}

func (r *queryResolver) ClusterRoles(ctx context.Context) ([]*v1.ClusterRole, error) {
	return r.roles.ClusterRolesQuery(ctx)
}

func (r *queryResolver) ClusterRole(ctx context.Context, name string) (*v1.ClusterRole, error) {
	return r.roles.ClusterRoleQuery(ctx, name)
}

func (r *queryResolver) RoleBindings(ctx context.Context, namespace string) ([]*v1.RoleBinding, error) {
	return r.roles.RoleBindingsQuery(ctx, namespace)
}

func (r *queryResolver) ClusterRoleBindings(ctx context.Context) ([]*v1.ClusterRoleBinding, error) {
	return r.roles.ClusterRoleBindingsQuery(ctx)
}
