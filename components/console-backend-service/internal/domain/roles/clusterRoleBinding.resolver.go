package roles

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
)

type ClusterRoleBindingList []*v1.ClusterRoleBinding

func (l *ClusterRoleBindingList) Append() interface{} {
	e := &v1.ClusterRoleBinding{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) ClusterRoleBindingsQuery(ctx context.Context) ([]*v1.ClusterRoleBinding, error) {
	items := ClusterRoleBindingList{}
	var err error
	err = r.ClusterRoleBindingService().List(&items)
	return items, err
}

func (r *Resolver) DeleteClusterRoleBinding(ctx context.Context, name string) (*v1.ClusterRoleBinding, error) {
	result := &v1.ClusterRoleBinding{}
	err := r.ClusterRoleBindingService().Delete(name, result)
	return result, err
}
