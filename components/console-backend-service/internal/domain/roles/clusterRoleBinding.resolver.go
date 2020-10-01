package roles

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"

	v1 "k8s.io/api/rbac/v1"
)

func ClusterRoleBindingsAsPointers(objs []v1.ClusterRoleBinding) []*v1.ClusterRoleBinding {
	var pointers []*v1.ClusterRoleBinding
	for i := range objs {
		pointers = append(pointers, &objs[i])
	}
	return pointers
}

func (r *Resolver) ClusterRoleBindingsQuery(ctx context.Context) ([]*v1.ClusterRoleBinding, error) {
	bindings := &v1.ClusterRoleBindingList{}
	err := r.List(ctx, bindings)
	sort.Slice(bindings.Items, func(i, j int) bool {
		return bindings.Items[i].Name < bindings.Items[j].Name
	})
	return ClusterRoleBindingsAsPointers(bindings.Items), err
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

	err := r.Create(ctx, clusterRoleBinding)
	return clusterRoleBinding, err
}

func (r *Resolver) DeleteClusterRoleBinding(ctx context.Context, name string) (*v1.ClusterRoleBinding, error) {
	result := &v1.ClusterRoleBinding{}
	err := r.Delete(ctx, name, result)
	return result, err
}

func (r *Resolver) ClusterRoleBindingSubscription(ctx context.Context) (<-chan *gqlschema.ClusterRoleBindingEvent, error) {
	channel := make(chan *gqlschema.ClusterRoleBindingEvent, 1)
	filter := func(apiRule v1.ClusterRoleBinding) bool { return true }

	unsubscribe, err := r.Subscribe(NewClusterRoleBindingEventHandler(channel, filter))
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
