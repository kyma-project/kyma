package trigger

import (
	"errors"
	"fmt"
	"knative.dev/pkg/apis"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

//go:generate mockery -name=GQLConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=GQLConverter -case=underscore -output disabled -outpkg disabled
type GQLConverter interface {
	ToGQL(in *v1alpha1.Trigger) (*gqlschema.Trigger, error)
	ToGQLs(in []*v1alpha1.Trigger) ([]gqlschema.Trigger, error)
	ToTrigger(in gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) (*v1alpha1.Trigger, error)
	ToTriggers(in []gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) ([]*v1alpha1.Trigger, error)
}

type triggerConverter struct {
}

func NewTriggerConverter() GQLConverter {
	return &triggerConverter{}
}

func (c *triggerConverter) ToGQL(in *v1alpha1.Trigger) (*gqlschema.Trigger, error) {
	if in == nil {
		return nil, errors.New("input trigger cannot be nil")
	}

	attributes := solveAttributes(in.Spec.Filter)
	status := calculateStatus(in.Status)
	dest, err := solveDestination(in.Spec.Subscriber)
	if err != nil {
		return nil, err
	}

	return &gqlschema.Trigger{
		Name:             in.Name,
		Namespace:        in.Namespace,
		Broker:           in.Spec.Broker,
		FilterAttributes: attributes,
		Subscriber:       dest,
		Status:           status,
	}, nil
}

func (c *triggerConverter) ToGQLs(in []*v1alpha1.Trigger) ([]gqlschema.Trigger, error) {
	if in == nil {
		return nil, errors.New("input triggers cannot be nil")
	}

	triggers := []gqlschema.Trigger{}
	for _, trigger := range in {
		item, err := c.ToGQL(trigger)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, *item)
	}
	return triggers, nil
}

func (c *triggerConverter) ToTrigger(in gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) (*v1alpha1.Trigger, error) {
	meta := v1alpha1.SchemeGroupVersion.WithKind("Trigger")
	trigger := &v1alpha1.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       meta.Kind,
			APIVersion: meta.GroupVersion().String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            in.Name,
			Namespace:       in.Namespace,
			OwnerReferences: []v1.OwnerReference{},
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: in.Broker,
			Filter: solveFilters(in.FilterAttributes),
		},
		Status: v1alpha1.TriggerStatus{},
	}

	for _, ref := range ownerRef {
		trigger.OwnerReferences = append(trigger.OwnerReferences, v1.OwnerReference{
			APIVersion:         ref.APIVersion,
			Kind:               ref.Kind,
			Name:               ref.Name,
			Controller:         &ref.Controller,
			BlockOwnerDeletion: &ref.BlockOwnerDeletion,
		})
	}

	subscriber, err := solveSubscriberInput(in.Subscriber)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("while resolving subscriber for trigger `%s`", trigger.Name))
	}
	trigger.Spec.Subscriber = subscriber

	return trigger, nil
}

func (c *triggerConverter) ToTriggers(in []gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) ([]*v1alpha1.Trigger, error) {
	var triggers []*v1alpha1.Trigger
	for _, trigger := range in {
		item, err := c.ToTrigger(trigger, ownerRef)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, item)
	}
	return triggers, nil
}

func solveFilters(json gqlschema.JSON) *v1alpha1.TriggerFilter {
	filters := make(v1alpha1.TriggerFilterAttributes)
	for key, value := range json {
		filters[key] = fmt.Sprint(value)
	}
	return &v1alpha1.TriggerFilter{
		Attributes: &filters,
	}
}

func calculateStatus(status v1alpha1.TriggerStatus) gqlschema.TriggerStatus {
	gqlStatus := gqlschema.TriggerStatus{
		Status: gqlschema.TriggerStatusTypeReady,
	}
	for _, condition := range status.Conditions {
		if condition.IsFalse() {
			gqlStatus.Reason = append(gqlStatus.Reason, condition.GetReason())
			gqlStatus.Status = gqlschema.TriggerStatusTypeFailed
		}
	}
	return gqlStatus
}

func solveDestination(dest duckv1.Destination) (gqlschema.Subscriber, error) {
	if dest.URI != nil {
		uri := dest.URI.Path
		return gqlschema.Subscriber{URI: &uri}, nil
	} else if dest.Ref != nil {
		return gqlschema.Subscriber{Ref: &gqlschema.SubscriberRef{
			APIVersion: dest.Ref.APIVersion,
			Kind:       dest.Ref.Kind,
			Name:       dest.Ref.Name,
			Namespace:  dest.Ref.Namespace,
		}}, nil
	}

	return gqlschema.Subscriber{}, errors.New("no data inside `destination` structure")
}

func solveSubscriberInput(ref gqlschema.SubscriberInput) (duckv1.Destination, error) {
	if ref.URI != nil {
		url, err := apis.ParseURL(*ref.URI)
		if err != nil {
			return duckv1.Destination{}, err
		}
		return duckv1.Destination{URI: url}, nil
	} else if ref.Ref != nil {
		return duckv1.Destination{
			Ref: &duckv1.KReference{
				APIVersion: ref.Ref.APIVersion,
				Kind:       ref.Ref.Kind,
				Name:       ref.Ref.Name,
				Namespace:  ref.Ref.Namespace,
			},
		}, nil
	}

	return duckv1.Destination{}, errors.New("no data inside `subscriberInput` structure")
}

func solveAttributes(filter *v1alpha1.TriggerFilter) map[string]interface{} {
	attr := make(map[string]interface{})
	if filter != nil && filter.Attributes != nil {
		for key, value := range *filter.Attributes {
			attr[key] = value
		}
	}

	return attr
}
