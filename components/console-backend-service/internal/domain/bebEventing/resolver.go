package bebEventing

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

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

func (r *Resolver) EventSubscriptionsQuery(ctx context.Context, ownerName string, namespace string) ([]*v1alpha1.Subscription, error) {
	items := EventSubscriptionList{}
	err := r.Service().ListByIndex(eventSubscriptionServiceIndex, eventSubscriptionServiceIndexKey(namespace, ownerName), &items)
	return items, err
}

func (r *Resolver) CreateEventSubscription(ctx context.Context, namespace string, name string, params gqlschema.EventSubscriptionSpecInput) (*v1alpha1.Subscription, error) {
	sourceName, err := r.getBEBSourceName()
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching sourceName")
	}
	spec := r.createSpec(params, namespace, sourceName)
	eventSubscription := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: subscriptionsGroupVersionResource.GroupVersion().String(),
			Kind:       subscriptionsKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				*params.OwnerRef,
			},
		},
		Spec: spec,
	}

	result := &v1alpha1.Subscription{}
	err = r.Service().Create(eventSubscription, result)
	return result, err
}

func (r *Resolver) UpdateEventSubscription(ctx context.Context, namespace string, name string, params gqlschema.EventSubscriptionSpecInput) (*v1alpha1.Subscription, error) {
	if len(params.Filters) == 0 {
		return r.DeleteEventSubscription(ctx, namespace, name)
	}

	result := &v1alpha1.Subscription{}
	sourceName, err := r.getBEBSourceName()
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching sourceName")
	}
	err = r.Service().UpdateInNamespace(name, namespace, result, func() error {
		result.Spec = r.createSpec(params, namespace, sourceName)
		return nil
	})
	return result, err
}

func (r *Resolver) DeleteEventSubscription(ctx context.Context, namespace string, name string) (*v1alpha1.Subscription, error) {
	result := &v1alpha1.Subscription{}
	err := r.Service().DeleteInNamespace(namespace, name, result)
	return result, err
}

func (r *Resolver) SubscribeEventSubscription(ctx context.Context, ownerName, namespace string) (<-chan *gqlschema.SubscriptionEvent, error) {
	channel := make(chan *gqlschema.SubscriptionEvent, 1)
	filter := func(subscription v1alpha1.Subscription) bool {
		namespaceMatches := subscription.ObjectMeta.Namespace == namespace
		ownerNameMatches := len(subscription.ObjectMeta.OwnerReferences) != 0 && subscription.ObjectMeta.OwnerReferences[0].Name == ownerName
		return namespaceMatches && ownerNameMatches
	}

	unsubscribe, err := r.Service().Subscribe(NewEventHandler(channel, filter))
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

func (r *Resolver) getBEBSourceName() (string, error) {
	secrets, err := r.client.CoreV1().Secrets("kyma-system").Get(context.Background(), "eventing", metav1.GetOptions{})

	if err != nil {
		return "", err
	}

	sourceName := secrets.Data["authentication.bebNamespace"]
	return string(sourceName), nil
}

func (r *Resolver) createSpec(params gqlschema.EventSubscriptionSpecInput, namespace string, sourceName string) v1alpha1.SubscriptionSpec {
	protocolSettings := &v1alpha1.ProtocolSettings{
		ContentMode:     "",
		ExemptHandshake: true,
		Qos:             "AT-LEAST-ONCE",
		WebhookAuth:     nil,
	}

	bebFilters := &v1alpha1.BebFilters{
		Dialect: "beb,",
		Filters: r.createBebFilters(params.Filters, sourceName),
	}

	spec := v1alpha1.SubscriptionSpec{
		Protocol:         "BEB",
		ProtocolSettings: protocolSettings,
		Sink:             fmt.Sprintf("http://%s.%s.svc.cluster.local", params.OwnerRef.Name, namespace),
		Filter:           bebFilters,
	}
	return spec
}

func (r *Resolver) createBebFilters(filters []*gqlschema.FiltersInput, sourceName string) []*v1alpha1.BebFilter {

	eventSource := v1alpha1.Filter{
		Type:     "exact",
		Property: "source",
		Value:    sourceName,
	}

	var bebFilters []*v1alpha1.BebFilter

	for _, filter := range filters {
		eventType := v1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    fmt.Sprintf("sap.kyma.custom.%s.%s.%s", filter.ApplicationName, filter.EventName, filter.Version),
		}
		bebFilters = append(bebFilters, &v1alpha1.BebFilter{
			EventSource: &eventSource,
			EventType:   &eventType,
		})
	}

	return bebFilters
}
