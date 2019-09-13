package listener

import (
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlBindingUsageConverter -output=automock -outpkg=automock -case=underscore
type gqlBindingUsageConverter interface {
	ToGQL(in *api.ServiceBindingUsage) (*gqlschema.ServiceBindingUsage, error)
}

type BindingUsage struct {
	channel   chan<- gqlschema.ServiceBindingUsageEvent
	filter    func(bindingUsage *api.ServiceBindingUsage) bool
	converter gqlBindingUsageConverter
	extractor extractor.BindingUsageUnstructuredExtractor
}

func NewBindingUsage(channel chan<- gqlschema.ServiceBindingUsageEvent, filter func(bindingUsage *api.ServiceBindingUsage) bool, converter gqlBindingUsageConverter) *BindingUsage {
	return &BindingUsage{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.BindingUsageUnstructuredExtractor{},
	}
}

func (l *BindingUsage) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *BindingUsage) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *BindingUsage) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *BindingUsage) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	bindingUsage, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(errors.New("cannot extract *ServiceBindingUsage"))
		return
	}
	if bindingUsage == nil {
		return
	}

	if l.filter(bindingUsage) {
		l.notify(eventType, bindingUsage)
	}
}

func (l *BindingUsage) notify(eventType gqlschema.SubscriptionEventType, bindingUsage *api.ServiceBindingUsage) {
	gqlBindingUsage, err := l.converter.ToGQL(bindingUsage)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ServiceBindingUsage"))
		return
	}
	if gqlBindingUsage == nil {
		return
	}

	event := gqlschema.ServiceBindingUsageEvent{
		Type:                eventType,
		ServiceBindingUsage: *gqlBindingUsage,
	}

	l.channel <- event
}
