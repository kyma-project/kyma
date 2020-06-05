package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlConfigMapConverter -output=automock -outpkg=automock -case=underscore
type gqlConfigMapConverter interface {
	ToGQL(in *v1.ConfigMap) (*gqlschema.ConfigMap, error)
}

type ConfigMap struct {
	channel   chan<- gqlschema.ConfigMapEvent
	filter    func(configMap *v1.ConfigMap) bool
	converter gqlConfigMapConverter
}

func NewConfigMap(channel chan<- gqlschema.ConfigMapEvent, filter func(configMap *v1.ConfigMap) bool, converter gqlConfigMapConverter) *ConfigMap {
	return &ConfigMap{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *ConfigMap) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ConfigMap) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ConfigMap) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ConfigMap) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	configMap, ok := object.(*v1.ConfigMap)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ConfigMap", object))
		return
	}

	if l.filter(configMap) {
		l.notify(eventType, configMap)
	}
}

func (l *ConfigMap) notify(eventType gqlschema.SubscriptionEventType, configMap *v1.ConfigMap) {
	gqlConfigMap, err := l.converter.ToGQL(configMap)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ConfigMap"))
		return
	}
	if gqlConfigMap == nil {
		return
	}

	event := gqlschema.ConfigMapEvent{
		Type:      eventType,
		ConfigMap: *gqlConfigMap,
	}

	l.channel <- event
}
