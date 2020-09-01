package roles

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
)

type RolesList []*v1.Role

func (l *RolesList) Append() interface{} {
	e := &v1.Role{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) RolesQuery(ctx context.Context, namespace string) ([]*v1.Role, error) {
	items := RolesList{}
	var err error
	err = r.RoleService().ListInNamespace(namespace, &items)
	return items, err
}

func (r *Resolver) RoleQuery(ctx context.Context, name string, namespace string) (*v1.Role, error) {
	var result *v1.Role
	err := r.RoleService().GetInNamespace(name, namespace, &result)
	return result, err
}
