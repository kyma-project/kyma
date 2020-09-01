package roles

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	err = r.RoleBindingService().ListInNamespace(namespace, &items)
	return items, err
}

func (r *Resolver) CreateRoleBinding(ctx context.Context, name string, namespace string, params gqlschema.RoleBindingInput) (*v1.RoleBinding, error) {
	convertedSubjects := make([]v1.Subject, len(params.Subjects))

	for i, sub := range params.Subjects {
		convertedSubjects[i] = v1.Subject{
			Kind:     string(sub.Kind),
			APIGroup: clusterRoleGroupVersionResource.Group,
			Name:     sub.Name,
		}
	}

	roleBinding := &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subjects: convertedSubjects,
		RoleRef: v1.RoleRef{
			Kind: string(params.RoleKind),
			Name: params.RoleName,
		},
	}

	result := &v1.RoleBinding{}
	err := r.RoleBindingService().Create(roleBinding, result)
	return result, err
}

func (r *Resolver) DeleteRoleBinding(ctx context.Context, namespace string, name string) (*v1.RoleBinding, error) {
	result := &v1.RoleBinding{}
	err := r.RoleBindingService().DeleteInNamespace(namespace, name, result)
	return result, err
}
