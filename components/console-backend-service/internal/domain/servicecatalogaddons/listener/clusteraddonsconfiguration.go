package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlClusterAddonsConfigurationConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterAddonsConfigurationConverter interface {
	ToGQL(item *v1alpha1.ClusterAddonsConfiguration) *gqlschema.AddonsConfiguration
}

type ClusterAddonsConfiguration struct {
	channel   chan<- *gqlschema.ClusterAddonsConfigurationEvent
	filter    func(entity *v1alpha1.ClusterAddonsConfiguration) bool
	converter gqlClusterAddonsConfigurationConverter
	extractor extractor.ClusterAddonsUnstructuredExtractor
}

func NewClusterAddonsConfiguration(channel chan<- *gqlschema.ClusterAddonsConfigurationEvent, filter func(entity *v1alpha1.ClusterAddonsConfiguration) bool, converter gqlClusterAddonsConfigurationConverter) *ClusterAddonsConfiguration {
	return &ClusterAddonsConfiguration{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.ClusterAddonsUnstructuredExtractor{},
	}
}

func (l *ClusterAddonsConfiguration) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ClusterAddonsConfiguration) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ClusterAddonsConfiguration) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ClusterAddonsConfiguration) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *v1alpha1.ClusterAddonsConfiguration", object))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ClusterAddonsConfiguration) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.ClusterAddonsConfiguration) {
	gqlAddonsConfiguration := l.converter.ToGQL(entity)
	if gqlAddonsConfiguration == nil {
		return
	}

	event := &gqlschema.ClusterAddonsConfigurationEvent{
		Type:                eventType,
		AddonsConfiguration: gqlAddonsConfiguration,
	}

	l.channel <- event
}
