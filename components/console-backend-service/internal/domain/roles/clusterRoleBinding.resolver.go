package roles

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func (r *Resolver) CreateClusterRoleBinding(ctx context.Context, name string, params gqlschema.ClusterRoleBindingInput) (*v1.ClusterRoleBinding, error) {
	convertedSubjects := make([]v1.Subject, len(params.Subjects))

	for i, sub := range params.Subjects {
		convertedSubjects[i] = v1.Subject{
			Kind:     string(sub.Kind),
			APIGroup: clusterRoleGroupVersionResource.Group,
			Name:     sub.Name,
		}
	}

	clusterRoleBinding := &v1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: convertedSubjects,
		RoleRef: v1.RoleRef{
			Kind: "ClusterRole",
			Name: params.RoleName,
		},
	}

	result := &v1.ClusterRoleBinding{}
	err := r.ClusterRoleBindingService().Create(clusterRoleBinding, result)
	return result, err
}

func (r *Resolver) DeleteClusterRoleBinding(ctx context.Context, name string) (*v1.ClusterRoleBinding, error) {
	result := &v1.ClusterRoleBinding{}
	err := r.ClusterRoleBindingService().Delete(name, result)
	return result, err
}
