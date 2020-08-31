package roles

import (
	"context"
	"fmt"

	"k8s.io/api/rbac/v1alpha1"
)

type RolesList []*v1alpha1.Role

func (l *RolesList) Append() interface{} {
	e := &v1alpha1.Role{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) RolesQuery(ctx context.Context, namespace string) ([]*v1alpha1.Role, error) {
	items := RolesList{}
	var err error
	fmt.Println(r.Service())
	err = r.Service().ListInNamespace(namespace, &items)
	return items, err
}
