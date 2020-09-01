package roles

import (
	"context"
	"sort"

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
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items, err
}

func (r *Resolver) RoleQuery(ctx context.Context, namespace string, name string) (*v1.Role, error) {
	var result *v1.Role
	err := r.RoleService().GetInNamespace(name, namespace, &result)
	return result, err
}
