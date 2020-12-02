package bebEventing

import (
	"context"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventSubscriptionList []*v1alpha1.Subscription

func (l *EventSubscriptionList) Append() interface{} {
	e := &v1alpha1.Subscription{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) EventSubscriptionQuery(ctx context.Context, namespace string, name string) (*v1alpha1.Subscription, error) {
	var result *v1alpha1.Subscription
	err := r.Service().GetInNamespace(name, namespace, &result)
	return result, err
}

func (r *Resolver) EventSubscriptionsQuery(ctx context.Context, namespace string) ([]*v1alpha1.Subscription, error) {
	items := EventSubscriptionList{}
	err := r.Service().ListInNamespace(namespace, &items)
	return items, err
}

func (r *Resolver) CreateEventSubscription(ctx context.Context, namespace string,  name string, params v1alpha1.SubscriptionSpec) (*v1alpha1.Subscription, error) {
	eventSubscription := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: subscriptionsGroupVersionResource.GroupVersion().String(),
			Kind:       subscriptionsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: params,
	}
	result := &v1alpha1.Subscription{}
	err := r.Service().Create(eventSubscription, result)
	return result, err
}