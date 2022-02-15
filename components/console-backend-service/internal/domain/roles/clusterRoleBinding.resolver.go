package roles

import (
	"context"
	"sort"

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
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items, err
}

func (r *Resolver) CreateClusterRoleBinding(ctx context.Context, name string, params gqlschema.ClusterRoleBindingInput) (*v1.ClusterRoleBinding, error) {
	convertedSubjects := make([]v1.Subject, len(params.Subjects))

	for i, sub := range params.Subjects {
		convertedSubjects[i] = v1.Subject{
			Kind:     string(sub.Kind),
			APIGroup: clusterRoleBindingGroupVersionResource.Group,
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

func (r *Resolver) ClusterRoleBindingSubscription(ctx context.Context) (<-chan *gqlschema.ClusterRoleBindingEvent, error) {
	channel := make(chan *gqlschema.ClusterRoleBindingEvent, 1)
	filter := func(apiRule v1.ClusterRoleBinding) bool { return true }

	unsubscribe, err := r.ClusterRoleBindingService().Subscribe(NewClusterRoleBindingEventHandler(channel, filter))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(channel)
		defer unsubscribe()
		<-ctx.Done()
	}()

	return channel, nil
}
