package roles

import (
	"context"
	v1 "k8s.io/api/rbac/v1"
	"sort"
)

func RolesAsPointers(objs []v1.Role) []*v1.Role {
	var pointers []*v1.Role
	for i := range objs {
		pointers = append(pointers, &objs[i])
	}
	return pointers
}

func (r *Resolver) RolesQuery(ctx context.Context, namespace string) ([]*v1.Role, error) {
	roles := &v1.RoleList{}
	err := r.ListInNamespace(ctx, namespace, roles)
	sort.Slice(roles.Items, func(i, j int) bool {
		return roles.Items[i].Name < roles.Items[j].Name
	})
	return RolesAsPointers(roles.Items), err
}

func (r *Resolver) RoleQuery(ctx context.Context, namespace string, name string) (*v1.Role, error) {
	role := &v1.Role{}
	err := r.GetInNamespace(ctx, namespace, name, role)
	return role, err
}
