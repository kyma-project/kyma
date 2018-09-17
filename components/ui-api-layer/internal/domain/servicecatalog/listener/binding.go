package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

//go:generate mockery -name=gqlBindingConverter -output=automock -outpkg=automock -case=underscore
type gqlBindingConverter interface {
	ToGQL(in *api.ServiceBinding) *gqlschema.ServiceBinding
}

type Binding struct {
	channel   chan<- gqlschema.ServiceBindingEvent
	filter    func(bindingUsage *api.ServiceBinding) bool
	converter gqlBindingConverter
}

func NewBinding(channel chan<- gqlschema.ServiceBindingEvent, filter func(binding *api.ServiceBinding) bool, converter gqlBindingConverter) *Binding {
	return &Binding{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *Binding) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Binding) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Binding) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Binding) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	binding, ok := object.(*api.ServiceBinding)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ServiceBinding", object))
		return
	}

	if l.filter(binding) {
		l.notify(eventType, binding)
	}
}

func (l *Binding) notify(eventType gqlschema.SubscriptionEventType, binding *api.ServiceBinding) {
	gqlBinding := l.converter.ToGQL(binding)
	if gqlBinding == nil {
		return
	}

	event := gqlschema.ServiceBindingEvent{
		Type:    eventType,
		Binding: *gqlBinding,
	}

	l.channel <- event
}
