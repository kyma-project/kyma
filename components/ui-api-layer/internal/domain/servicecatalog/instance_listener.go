package servicecatalog

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

// TODO: Move to subpackage

type instanceListener struct {
	channel           chan<- gqlschema.ServiceInstanceEvent
	filter            func(object interface{}) bool
	instanceConverter gqlInstanceConverter
}

func newInstanceListener(channel chan<- gqlschema.ServiceInstanceEvent, filter func(object interface{}) bool, instanceConverter gqlInstanceConverter) *instanceListener {
	return &instanceListener{
		channel:           channel,
		filter:            filter,
		instanceConverter: instanceConverter,
	}
}

func (il *instanceListener) OnAdd(object interface{}) {
	if il.filter(object) {
		il.notify(gqlschema.ServiceInstanceEventTypeAdd, object)
	}
}

func (il *instanceListener) OnUpdate(oldObject, newObject interface{}) {
	if il.filter(newObject) {
		il.notify(gqlschema.ServiceInstanceEventTypeUpdate, newObject)
	}
}

func (il *instanceListener) OnDelete(object interface{}) {
	if il.filter(object) {
		il.notify(gqlschema.ServiceInstanceEventTypeDelete, object)
	}
}

func (il *instanceListener) notify(eventType gqlschema.ServiceInstanceEventType, object interface{}) {
	instance, ok := object.(*v1beta1.ServiceInstance)
	if !ok {
		glog.Error(fmt.Errorf("Incorrect object type: %T, should be: *ServiceInstance", object))
		return
	}

	gqlInstance := il.instanceConverter.ToGQL(instance)
	event := gqlschema.ServiceInstanceEvent{
		Type:     eventType,
		Instance: *gqlInstance,
	}

	il.channel <- event
}
