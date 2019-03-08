package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlKServiceConverter -output=automock -outpkg=automock -case=underscore
type gqlKServiceConverter interface {
	ToGQL(in *v1.Service) *gqlschema.KService
}

type KService struct {
	channel   chan<- gqlschema.ServiceEvent
	filter    func(pod *v1.Service) bool
	converter gqlKServiceConverter
}

func NewService(channel chan<- gqlschema.ServiceEvent,
	filter func(service *v1.Service) bool,
	converter gqlKServiceConverter) *KService {

	return &KService{channel, filter, converter}
}

func (l *KService) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *KService) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *KService) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *KService) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	service, ok := object.(*v1.Service)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Service", object))
		return
	}

	if l.filter(service) {
		l.notify(eventType, service)
	}
}

func (l *KService) notify(eventType gqlschema.SubscriptionEventType, service *v1.Service) {
	gqlKService := l.converter.ToGQL(service)
	if gqlKService == nil {
		return
	}

	event := gqlschema.ServiceEvent{
		Type:    eventType,
		Service: *gqlKService,
	}

	l.channel <- event
}
