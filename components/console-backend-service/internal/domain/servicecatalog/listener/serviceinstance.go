package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlInstanceConverter -output=automock -outpkg=automock -case=underscore
type gqlInstanceConverter interface {
	ToGQL(in *api.ServiceInstance) (*gqlschema.ServiceInstance, error)
}

type Instance struct {
	channel   chan<- gqlschema.ServiceInstanceEvent
	filter    func(instance *api.ServiceInstance) bool
	converter gqlInstanceConverter
}

func NewInstance(channel chan<- gqlschema.ServiceInstanceEvent, filter func(instance *api.ServiceInstance) bool, converter gqlInstanceConverter) *Instance {
	return &Instance{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *Instance) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Instance) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Instance) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Instance) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	instance, ok := object.(*api.ServiceInstance)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ServiceInstance", object))
		return
	}

	if l.filter(instance) {
		l.notify(eventType, instance)
	}
}

func (l *Instance) notify(eventType gqlschema.SubscriptionEventType, instance *api.ServiceInstance) {
	gqlInstance, err := l.converter.ToGQL(instance)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ServiceInstance"))
		return
	}
	if gqlInstance == nil {
		return
	}

	event := gqlschema.ServiceInstanceEvent{
		Type:            eventType,
		ServiceInstance: *gqlInstance,
	}

	l.channel <- event
}
