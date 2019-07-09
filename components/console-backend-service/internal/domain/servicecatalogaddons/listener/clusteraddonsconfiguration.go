package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

//go:generate mockery -name=gqlClusterAddonsConfigurationConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterAddonsConfigurationConverter interface {
	ToGQL(item *v1alpha1.ClusterAddonsConfiguration) *gqlschema.AddonsConfiguration
}

type AddonsConfiguration struct {
	channel   chan<- gqlschema.AddonsConfigurationEvent
	filter    func(entity *v1alpha1.ClusterAddonsConfiguration) bool
	converter gqlClusterAddonsConfigurationConverter
}

func NewClusterAddonsConfiguration(channel chan<- gqlschema.AddonsConfigurationEvent, filter func(entity *v1alpha1.ClusterAddonsConfiguration) bool, converter gqlClusterAddonsConfigurationConverter) *AddonsConfiguration {
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
	entity, ok := object.(*v1alpha1.ClusterAddonsConfiguration)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *v1alpha1.ClusterAddonsConfiguration", object))
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *AddonsConfiguration) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.ClusterAddonsConfiguration) {
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
