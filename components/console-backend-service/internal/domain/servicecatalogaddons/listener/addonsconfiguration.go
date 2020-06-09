package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlAddonsConfigurationConverter -output=automock -outpkg=automock -case=underscore
type gqlAddonsConfigurationConverter interface {
	ToGQL(item *v1alpha1.AddonsConfiguration) *gqlschema.AddonsConfiguration
}

type AddonsConfiguration struct {
	channel   chan<- *gqlschema.AddonsConfigurationEvent
	filter    func(entity *v1alpha1.AddonsConfiguration) bool
	converter gqlAddonsConfigurationConverter
	extractor extractor.AddonsUnstructuredExtractor
}

func NewAddonsConfiguration(channel chan<- *gqlschema.AddonsConfigurationEvent, filter func(entity *v1alpha1.AddonsConfiguration) bool, converter gqlAddonsConfigurationConverter) *AddonsConfiguration {
	return &AddonsConfiguration{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.AddonsUnstructuredExtractor{},
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
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *v1alpha1.AddonsConfiguration", object))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *AddonsConfiguration) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.AddonsConfiguration) {
	gqlAddonsConfiguration := l.converter.ToGQL(entity)
	if gqlAddonsConfiguration == nil {
		return
	}

	event := &gqlschema.AddonsConfigurationEvent{
		Type:                eventType,
		AddonsConfiguration: gqlAddonsConfiguration,
	}

	l.channel <- event
}
