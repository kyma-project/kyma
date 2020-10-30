package roles

import (
	"context"
	"sort"

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
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items, err
}

func (r *Resolver) CreateRoleBinding(ctx context.Context, namespace string, name string, params gqlschema.RoleBindingInput) (*v1.RoleBinding, error) {
	convertedSubjects := make([]v1.Subject, len(params.Subjects))

	for i, sub := range params.Subjects {
		convertedSubjects[i] = v1.Subject{
			Kind:     string(sub.Kind),
			APIGroup: roleBindingGroupVersionResource.Group,
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

	result := &v1.RoleBinding{}
	err := r.RoleBindingService().Create(roleBinding, result)
	return result, err
}

func (r *Resolver) DeleteRoleBinding(ctx context.Context, namespace string, name string) (*v1.RoleBinding, error) {
	result := &v1.RoleBinding{}
	err := r.RoleBindingService().DeleteInNamespace(namespace, name, result)
	return result, err
}

func (r *Resolver) RoleBindingSubscription(ctx context.Context, namespace string) (<-chan *gqlschema.RoleBindingEvent, error) {
	channel := make(chan *gqlschema.RoleBindingEvent, 1)
	filter := func(apiRule v1.RoleBinding) bool {
		return apiRule.Namespace == namespace
	}
	unsubscribe, err := r.RoleBindingService().Subscribe(NewRoleBindingEventHandler(channel, filter))
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
