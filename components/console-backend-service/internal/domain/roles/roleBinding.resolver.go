package roles

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"

	v1 "k8s.io/api/rbac/v1"
)

func RoleBindingsAsPointers(objs []v1.RoleBinding) []*v1.RoleBinding {
	var pointers []*v1.RoleBinding
	for i := range objs {
		pointers = append(pointers, &objs[i])
	}
	return pointers
}

func (r *Resolver) RoleBindingsQuery(ctx context.Context, namespace string) ([]*v1.RoleBinding, error) {
	bindings := &v1.RoleBindingList{}
	err := r.ListInNamespace(ctx, namespace, bindings)
	sort.Slice(bindings.Items, func(i, j int) bool {
		return bindings.Items[i].Name < bindings.Items[j].Name
	})
	return RoleBindingsAsPointers(bindings.Items), err
}

func (r *Resolver) CreateRoleBinding(ctx context.Context, namespace string, name string, params gqlschema.RoleBindingInput) (*v1.RoleBinding, error) {
	convertedSubjects := make([]v1.Subject, len(params.Subjects))

	for i, sub := range params.Subjects {
		convertedSubjects[i] = v1.Subject{
			Kind:     string(sub.Kind),
			APIGroup: roleGroupVersionResource.Group,
			Name:     sub.Name,
		}
	}

	roleBinding := &v1.RoleBinding{
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

	err := r.Create(ctx, roleBinding)
	return roleBinding, err
}

func (r *Resolver) DeleteRoleBinding(ctx context.Context, namespace string, name string) (*v1.RoleBinding, error) {
	result := &v1.RoleBinding{}
	err := r.DeleteInNamespace(ctx, namespace, name, result)
	return result, err
}
//
//func (r *Resolver) RoleBindingSubscription(ctx context.Context, namespace string) (<-chan *gqlschema.RoleBindingEvent, error) {
//	channel := make(chan *gqlschema.RoleBindingEvent, 1)
//	filter := func(apiRule v1.RoleBinding) bool {
//		return apiRule.Namespace == namespace
//	}
//	unsubscribe, err := r.RoleBindingService().Subscribe(NewRoleBindingEventHandler(channel, filter))
//	if err != nil {
//		return nil, err
//	}
//
//	go func() {
//		defer close(channel)
//		defer unsubscribe()
//		<-ctx.Done()
//	}()
//
//	return channel, nil
//}
