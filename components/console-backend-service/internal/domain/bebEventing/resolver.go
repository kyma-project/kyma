package bebEventing

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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

//protocolSettings: ProtocolSettingsInput
//bebfilters.protocol = 'beb'

func (r *Resolver) CreateEventSubscription(ctx context.Context, namespace string,  name string, params gqlschema.EventSubscriptionSpecInput) (*v1alpha1.Subscription, error) {

	protocolSettings := &v1alpha1.ProtocolSettings{
		ContentMode:     "",
		ExemptHandshake: true,
		Qos:             "AT-LEAST-ONCE",
		WebhookAuth:     nil,
	}

	eventSource := v1alpha1.Filter{
		Type:     "exact",
		Property: "/default/sap.kyma/kh", //TODO
		Value:    "source",
	}

	filters := []

	bebFilters := &v1alpha1.BebFilters{
		Dialect: "beb,",
		Filters: filters,
	}

	spec := v1alpha1.SubscriptionSpec{
		Protocol:         "BEB",
		ProtocolSettings: protocolSettings,
		Sink:             fmt.Sprintf("http://%s.%s.svc.cluster.local", params.ServiceName, namespace),
		Filter:           bebFilters,
	}
	
	eventSubscription := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: subscriptionsGroupVersionResource.GroupVersion().String(),
			Kind:       subscriptionsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	result := &v1alpha1.Subscription{}
	err := r.Service().Create(eventSubscription, result)
	return result, err
}