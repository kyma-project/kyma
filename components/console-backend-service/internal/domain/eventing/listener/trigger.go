package listener

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/golang/glog"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToGQL(in *v1alpha1.Trigger) (*gqlschema.Trigger, error)
}

type Trigger struct {
	channel   chan<- gqlschema.TriggerEvent
	filter    func(entity *v1alpha1.Trigger) bool
	converter Converter
	extractor extractor.TriggerUnstructuredExtractor
}

func NewTrigger(extractor extractor.TriggerUnstructuredExtractor, channel chan<- gqlschema.TriggerEvent, filter func(entity *v1alpha1.Trigger) bool, converter Converter) *Trigger {
	return &Trigger{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor,
	}
}

func (t *Trigger) OnAdd(object interface{}) {
	t.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (t *Trigger) OnUpdate(oldObject, newObject interface{}) {
	t.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (t *Trigger) OnDelete(object interface{}) {
	t.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (t *Trigger) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := t.extractor.Do(object)
	if err != nil {
		glog.Error(errors.Wrapf(err, "incorrect object type: %T, should be: *%s", object, pretty.TriggerType))
		return
	}
	if entity == nil {
		return
	}

	if t.filter(entity) {
		t.notify(eventType, entity)
	}
}

func (t *Trigger) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.Trigger) {
	gqlTrigger, err := t.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *%s", pretty.TriggerType))
		return
	}
	if gqlTrigger == nil {
		return
	}

	event := gqlschema.TriggerEvent{
		Type:    eventType,
		Trigger: *gqlTrigger,
	}

	t.channel <- event
}
