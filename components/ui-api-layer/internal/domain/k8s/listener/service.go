package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlServiceConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceConverter interface {
	ToGQL(in *v1.Service) *gqlschema.Service
}

type Service struct {
	channel   chan<- gqlschema.ServiceEvent
	filter    func(pod *v1.Service) bool
	converter gqlServiceConverter
}

func NewService(channel chan<- gqlschema.ServiceEvent,
	filter func(service *v1.Service) bool,
	converter gqlServiceConverter) *Service {

	return &Service{channel, filter, converter}
}

func (l *Service) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Service) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Service) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Service) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	service, ok := object.(*v1.Service)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Service", object))
		return
	}

	if l.filter(service) {
		l.notify(eventType, service)
	}
}

func (l *Service) notify(eventType gqlschema.SubscriptionEventType, service *v1.Service) {
	gqlService := l.converter.ToGQL(service)
	if gqlService == nil {
		return
	}

	event := gqlschema.ServiceEvent{
		Type:    eventType,
		Service: *gqlService,
	}

	l.channel <- event
}
