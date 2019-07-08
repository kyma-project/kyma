package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlConfigMapAddonsConfigurationConverter -output=automock -outpkg=automock -case=underscore
type gqlConfigMapAddonsConfigurationConverter interface {
	ConfigMapToGQL(item *v1.ConfigMap) *gqlschema.AddonsConfiguration
}

type ConfigMapAddonsConfiguration struct {
	channel   chan<- gqlschema.AddonsConfigurationEvent
	filter    func(entity *v1.ConfigMap) bool
	converter gqlConfigMapAddonsConfigurationConverter
}

func NewConfigMapAddonsConfiguration(channel chan<- gqlschema.AddonsConfigurationEvent, filter func(entity *v1.ConfigMap) bool, converter gqlConfigMapAddonsConfigurationConverter) *ConfigMapAddonsConfiguration {
	return &ConfigMapAddonsConfiguration{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *ConfigMapAddonsConfiguration) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ConfigMapAddonsConfiguration) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ConfigMapAddonsConfiguration) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ConfigMapAddonsConfiguration) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, ok := object.(*v1.ConfigMap)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ConfigMap", object))
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ConfigMapAddonsConfiguration) notify(eventType gqlschema.SubscriptionEventType, entity *v1.ConfigMap) {
	gqlAddonsConfiguration := l.converter.ConfigMapToGQL(entity)
	if gqlAddonsConfiguration == nil {
		return
	}

	event := gqlschema.AddonsConfigurationEvent{
		Type:                eventType,
		AddonsConfiguration: *gqlAddonsConfiguration,
	}

	l.channel <- event
}
