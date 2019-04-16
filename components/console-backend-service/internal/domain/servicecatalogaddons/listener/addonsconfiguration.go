package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlAddonsConfigurationConverter -output=automock -outpkg=automock -case=underscore
type gqlAddonsConfigurationConverter interface {
	ToGQL(item *v1.ConfigMap) *gqlschema.AddonsConfiguration
}

type AddonsConfiguration struct {
	channel   chan<- gqlschema.AddonsConfigurationEvent
	filter    func(entity *v1.ConfigMap) bool
	converter gqlAddonsConfigurationConverter
}

func NewAddonsConfiguration(channel chan<- gqlschema.AddonsConfigurationEvent, filter func(entity *v1.ConfigMap) bool, converter gqlAddonsConfigurationConverter) *AddonsConfiguration {
	return &AddonsConfiguration{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *AddonsConfiguration) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *AddonsConfiguration) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *AddonsConfiguration) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *AddonsConfiguration) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, ok := object.(*v1.ConfigMap)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ConfigMap", object))
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *AddonsConfiguration) notify(eventType gqlschema.SubscriptionEventType, entity *v1.ConfigMap) {
	gqlAddonsConfiguration := l.converter.ToGQL(entity)
	if gqlAddonsConfiguration == nil {
		return
	}

	event := gqlschema.AddonsConfigurationEvent{
		Type:                eventType,
		AddonsConfiguration: *gqlAddonsConfiguration,
	}

	l.channel <- event
}
