package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlServiceBrokerConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceBrokerConverter interface {
	ToGQL(in *api.ServiceBroker) (*gqlschema.ServiceBroker, error)
}

type ServiceBroker struct {
	channel   chan<- *gqlschema.ServiceBrokerEvent
	filter    func(entity *api.ServiceBroker) bool
	converter gqlServiceBrokerConverter
}

func NewServiceBroker(channel chan<- *gqlschema.ServiceBrokerEvent, filter func(entity *api.ServiceBroker) bool, converter gqlServiceBrokerConverter) *ServiceBroker {
	return &ServiceBroker{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *ServiceBroker) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ServiceBroker) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ServiceBroker) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ServiceBroker) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, ok := object.(*api.ServiceBroker)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ServiceBroker", object))
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ServiceBroker) notify(eventType gqlschema.SubscriptionEventType, entity *api.ServiceBroker) {
	gqlServiceBroker, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ServiceBroker"))
		return
	}
	if gqlServiceBroker == nil {
		return
	}

	event := &gqlschema.ServiceBrokerEvent{
		Type:          eventType,
		ServiceBroker: gqlServiceBroker,
	}

	l.channel <- event
}
