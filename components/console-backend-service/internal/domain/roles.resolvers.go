package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/rbac/v1"
)

func (r *mutationResolver) CreateRoleBinding(ctx context.Context, name string, namespace string, params gqlschema.RoleBindingInput) (*v1.RoleBinding, error) {
	return r.roles.CreateRoleBinding(ctx, namespace, name, params)
}

func (r *mutationResolver) DeleteRoleBinding(ctx context.Context, namespace string, name string) (*v1.RoleBinding, error) {
	return r.roles.DeleteRoleBinding(ctx, namespace, name)
}

func (r *mutationResolver) CreateClusterRoleBinding(ctx context.Context, name string, params gqlschema.ClusterRoleBindingInput) (*v1.ClusterRoleBinding, error) {
	return r.roles.CreateClusterRoleBinding(ctx, name, params)
}

func (r *mutationResolver) DeleteClusterRoleBinding(ctx context.Context, name string) (*v1.ClusterRoleBinding, error) {
	return r.roles.DeleteClusterRoleBinding(ctx, name)
}

func (r *queryResolver) Roles(ctx context.Context, namespace string) ([]*v1.Role, error) {
	return r.roles.RolesQuery(ctx, namespace)
}

func (r *queryResolver) Role(ctx context.Context, namespace string, name string) (*v1.Role, error) {
	return r.roles.RoleQuery(ctx, namespace, name)
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

func (r *subscriptionResolver) RoleBindingEvent(ctx context.Context, namespace string) (<-chan *gqlschema.RoleBindingEvent, error) {
	return r.roles.RoleBindingSubscription(ctx, namespace)
}

func (r *subscriptionResolver) ClusterRoleBindingEvent(ctx context.Context) (<-chan *gqlschema.ClusterRoleBindingEvent, error) {
	return r.roles.ClusterRoleBindingSubscription(ctx)
}
