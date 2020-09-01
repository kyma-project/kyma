package roles

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
)

type RoleBindingList []*v1.RoleBinding

func (l *RoleBindingList) Append() interface{} {
	e := &v1.RoleBinding{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) RoleBindingsQuery(ctx context.Context, namespace string) ([]*v1.RoleBinding, error) {
	items := RoleBindingList{}
	var err error
	err = r.RoleBindingService().List(&items)
	return items, err
}

func (r *Resolver) DeleteRoleBinding(ctx context.Context, namespace string, name string) (*v1.RoleBinding, error) {
	result := &v1.RoleBinding{}
	err := r.RoleBindingService().DeleteInNamespace(namespace, name, result)
	return result, err
}
