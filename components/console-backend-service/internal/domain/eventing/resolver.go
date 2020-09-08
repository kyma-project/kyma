package eventing

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"knative.dev/pkg/apis"

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

func (r *Resolver) TriggersQuery(ctx context.Context, namespace string, serviceName string) ([]*v1alpha1.Trigger, error) {
	items := TriggerList{}

	err := r.Service().ListByIndex(triggerIndexKey, createTriggerRefIndexKey(namespace, serviceName), &items)
	err = r.Service().ListByIndex(triggerIndexKey, "commerce-mock.default.svc.cluster.local", &items)

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
	trigger, err := r.buildTrigger(namespace, in, ownerRef)

	if err != nil {
		return nil, err
	}

	result := &v1alpha1.Trigger{}
	err = r.Service().Create(trigger, result)
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

func (r *Resolver) TriggerEventSubscription(ctx context.Context, namespace, serviceName string) (<-chan *gqlschema.TriggerEvent, error) {
	channel := make(chan *gqlschema.TriggerEvent, 1)
	filter := func(trigger v1alpha1.Trigger) bool {
		namespaceMatches := trigger.Namespace == namespace
		serviceMatches := trigger.Spec.Subscriber.Ref.Name == serviceName
		return namespaceMatches && serviceMatches
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

func (r *Resolver) buildTrigger(namespace string, in gqlschema.TriggerCreateInput, ownerRef []*v1.OwnerReference) (*v1alpha1.Trigger, error) {
	in = *r.checkTriggerName(&in)

	meta := v1alpha1.SchemeGroupVersion.WithKind("Trigger")

	port := in.Subscriber.Port
	if port == nil {
		defaultPort := uint32(80)
		port = &defaultPort
	}
	path := in.Subscriber.Path
	if path == nil {
		defaultPath := "/"
		path = &defaultPath
	}

	uriString := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d%s", in.Subscriber.Ref.Name, namespace, *port, *path)
	uri, err := apis.ParseURL(uriString)
	if err != nil {
		return nil, errors.Wrap(err, "while creating trigger subscriber uri")
	}

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
				URI: uri,
			},
		},
	}

	for _, ref := range ownerRef {
		trigger.OwnerReferences = append(trigger.OwnerReferences, *ref)
	}

	return trigger, nil
}

func (r *Resolver) checkTriggerName(trigger *gqlschema.TriggerCreateInput) *gqlschema.TriggerCreateInput {
	if trigger.Name == nil || *trigger.Name == "" {
		name := r.generateName()
		trigger.Name = &name
	}
	return trigger
}
