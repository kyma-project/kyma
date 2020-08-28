package eventing

import (
	"context"
	"errors"
	"fmt"

	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type TriggerList []*v1alpha1.Trigger

func (l *TriggerList) Append() interface{} {
	e := &v1alpha1.Trigger{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) TriggersQuery(ctx context.Context, namespace string, subscriber *duckv1.Destination) ([]*v1alpha1.Trigger, error) {
	items := TriggerList{}
	var err error

	if subscriber == nil {
		err = r.Service().ListInNamespace(namespace, &items)
	} else if subscriber.Ref != nil {
		key := createTriggersSubscriberRefIndexKey(namespace, subscriber.Ref)
		err = r.Service().ListByIndex(triggersSubscriberRefIndex, key, &items)
	} else if subscriber.URI != nil {
		key := createTriggersSubscriberRefURIKey(namespace, subscriber.URI)
		err = r.Service().ListByIndex(triggersSubscriberURIIndex, key, &items)
	} else {
		return nil, errors.New("subscriber is not null but it is empty")
	}
	return items, err
}

func (r *Resolver) StatusField(ctx context.Context, obj *v1alpha1.Trigger) (*gqlschema.TriggerStatus, error) {
	gqlStatus := &gqlschema.TriggerStatus{
		Status: gqlschema.TriggerStatusTypeReady,
	}
	for _, condition := range obj.Status.Conditions {
		if condition.IsFalse() {
			gqlStatus.Reason = append(gqlStatus.Reason, condition.Message)
			gqlStatus.Status = gqlschema.TriggerStatusTypeFailed
		} else if condition.IsUnknown() {
			gqlStatus.Reason = append(gqlStatus.Reason, condition.Message)
			if gqlStatus.Status != gqlschema.TriggerStatusTypeFailed {
				gqlStatus.Status = gqlschema.TriggerStatusTypeUnknown
			}
		}
	}
	return gqlStatus, nil
}

func (r *Resolver) FilterField(ctx context.Context, obj *v1alpha1.TriggerSpec) (gqlschema.JSON, error) {
	attr := make(map[string]interface{})

	if obj.Filter == nil || obj.Filter.Attributes == nil {
		return attr, nil
	}
	for key, value := range *obj.Filter.Attributes {
		attr[key] = value
	}

	return attr, nil
}

func (r *Resolver) CreateTrigger(ctx context.Context, namespace string, in gqlschema.TriggerCreateInput, ownerRef []*v1.OwnerReference) (*v1alpha1.Trigger, error) {
	trigger := r.buildTrigger(namespace, in, ownerRef)

	result := &v1alpha1.Trigger{}
	err := r.Service().Create(trigger, result)
	return result, err
}

func (r *Resolver) CreateTriggers(ctx context.Context, namespace string, triggers []*gqlschema.TriggerCreateInput, ref []*v1.OwnerReference) ([]*v1alpha1.Trigger, error) {
	items := TriggerList{}
	for _, trigger := range triggers {
		result, err := r.CreateTrigger(ctx, namespace, *trigger, ref)
		if err != nil {
			return items, err
		}
		items = append(items, result)
	}
	return items, nil
}

func (r *Resolver) DeleteTrigger(ctx context.Context, namespace string, name string) (*v1alpha1.Trigger, error) {
	result := &v1alpha1.Trigger{}
	err := r.Service().DeleteInNamespace(namespace, name, result)
	return result, err
}

func (r *Resolver) DeleteTriggers(ctx context.Context, namespace string, names []string) ([]*v1alpha1.Trigger, error) {
	items := TriggerList{}
	for _, triggerName := range names {
		result, err := r.DeleteTrigger(ctx, namespace, triggerName)
		if err != nil {
			return items, err
		}
		items = append(items, result)
	}
	return items, nil
}

func (r *Resolver) TriggerEventSubscription(ctx context.Context, namespace string, subscriber *duckv1.Destination) (<-chan *gqlschema.TriggerEvent, error) {
	channel := make(chan *gqlschema.TriggerEvent, 1)
	filter := func(trigger v1alpha1.Trigger) bool {
		namespaceMatches := trigger.Namespace == namespace
		subscriberMatches := subscriber == nil || r.areSubscribersEqual(subscriber, trigger.Spec.Subscriber)
		return namespaceMatches && subscriberMatches
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

func (r *Resolver) solveFilters(json gqlschema.JSON) *v1alpha1.TriggerFilter {
	filters := make(v1alpha1.TriggerFilterAttributes)
	if json == nil {
		return nil
	}

	for key, value := range json {
		filters[key] = fmt.Sprint(value)
	}
	return &v1alpha1.TriggerFilter{
		Attributes: &filters,
	}
}

func (r *Resolver) buildTrigger(namespace string, in gqlschema.TriggerCreateInput, ownerRef []*v1.OwnerReference) *v1alpha1.Trigger {
	in = *r.checkTriggerName(&in)

	meta := v1alpha1.SchemeGroupVersion.WithKind("Trigger")
	trigger := &v1alpha1.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       meta.Kind,
			APIVersion: meta.GroupVersion().String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            *in.Name,
			Namespace:       namespace,
			OwnerReferences: []v1.OwnerReference{},
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: in.Broker,
			Filter: r.solveFilters(in.FilterAttributes),
			Subscriber: duckv1.Destination{
				Ref: in.Subscriber.Ref,
				URI: in.Subscriber.URI,
			},
		},
	}

	for _, ref := range ownerRef {
		trigger.OwnerReferences = append(trigger.OwnerReferences, *ref)
	}

	return trigger
}

func (r *Resolver) areSubscribersEqual(expected *duckv1.Destination, actual duckv1.Destination) bool {
	if expected.URI != nil {
		if actual.URI == nil || expected.URI.String() != actual.URI.String() {
			return false
		}
	}

	if expected.Ref != nil {
		if actual.Ref == nil {
			return false
		}
		return r.compareRefs(expected.Ref, actual.Ref)
	}
	return true
}

func (r *Resolver) compareRefs(r1 *duckv1.KReference, r2 *duckv1.KReference) bool {
	return r1.Name == r2.Name && r1.APIVersion == r2.APIVersion && r1.Kind == r2.Kind && r1.Namespace == r2.Namespace
}

func (r *Resolver) checkTriggerName(trigger *gqlschema.TriggerCreateInput) *gqlschema.TriggerCreateInput {
	if trigger.Name == nil || *trigger.Name == "" {
		name := r.generateName()
		trigger.Name = &name
	}
	return trigger
}
