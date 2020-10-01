package roles

import (
	"context"
	"sort"

	v1 "k8s.io/api/rbac/v1"
)

func ClusterRolesAsPointers(objs []v1.ClusterRole) []*v1.ClusterRole {
	var pointers []*v1.ClusterRole
	for i := range objs {
		pointers = append(pointers, &objs[i])
	}
	return pointers
}

func (r *Resolver) ClusterRolesQuery(ctx context.Context) ([]*v1.ClusterRole, error) {
	roles := &v1.ClusterRoleList{}
	err := r.List(ctx, roles)
	sort.Slice(roles.Items, func(i, j int) bool {
		return roles.Items[i].Name < roles.Items[j].Name
	})
	return ClusterRolesAsPointers(roles.Items), err
}

func (r *Resolver) ClusterRoleQuery(ctx context.Context, name string) (*v1.ClusterRole, error) {
	role := &v1.ClusterRole{}
	err := r.Get(ctx, name, role)
	return role, err
}
