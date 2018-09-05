package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=GQLBindingUsageConverter -output=automock -outpkg=automock -case=underscore
type GQLBindingUsageConverter interface {
	ToGQL(in *api.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error)
}

type BindingUsage struct {
	channel   chan<- gqlschema.ServiceBindingUsageEvent
	filter    func(bindingUsage *api.ServiceBindingUsage) bool
	converter GQLBindingUsageConverter
}

func NewBindingUsage(channel chan<- gqlschema.ServiceBindingUsageEvent, filter func(bindingUsage *api.ServiceBindingUsage) bool, converter GQLBindingUsageConverter) *BindingUsage {
	return &BindingUsage{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *BindingUsage) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *BindingUsage) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *BindingUsage) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *BindingUsage) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	bindingUsage, ok := object.(*api.ServiceBindingUsage)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ServiceBindingUsage", object))
		return
	}

	if l.filter(bindingUsage) {
		l.notify(eventType, bindingUsage)
	}
}

func (l *BindingUsage) notify(eventType gqlschema.SubscriptionEventType, bindingUsage *api.ServiceBindingUsage) {
	gqlBindingUsage, err := l.converter.ToGQL(bindingUsage)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ServiceBindingUsage"))
		return
	}
	if gqlBindingUsage == nil {
		return
	}

	event := gqlschema.ServiceBindingUsageEvent{
		Type:         eventType,
		BindingUsage: *gqlBindingUsage,
	}

	l.channel <- event
}
