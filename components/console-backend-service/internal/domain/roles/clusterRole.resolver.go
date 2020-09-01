package roles

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
)

type ClusterRolesList []*v1.ClusterRole

func (l *ClusterRolesList) Append() interface{} {
	e := &v1.ClusterRole{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) ClusterRolesQuery(ctx context.Context) ([]*v1.ClusterRole, error) {
	items := ClusterRolesList{}
	var err error
	err = r.ClusterRoleService().List(&items)
	return items, err
}

func (r *Resolver) ClusterRoleQuery(ctx context.Context, name string) (*v1.ClusterRole, error) {
	var result *v1.ClusterRole
	err := r.ClusterRoleService().Get(name, &result)
	return result, err
}
