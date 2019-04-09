package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlDocsTopicConverter -output=automock -outpkg=automock -case=underscore
type gqlDocsTopicConverter interface {
	ToGQL(in *v1alpha1.DocsTopic) (*gqlschema.DocsTopic, error)
}

type DocsTopic struct {
	channel   chan<- gqlschema.DocsTopicEvent
	filter    func(entity *v1alpha1.DocsTopic) bool
	converter gqlDocsTopicConverter
	extractor extractor.DocsTopicUnstructuredExtractor
}

func NewDocsTopic(channel chan<- gqlschema.DocsTopicEvent, filter func(entity *v1alpha1.DocsTopic) bool, converter gqlDocsTopicConverter) *DocsTopic {
	return &DocsTopic{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.DocsTopicUnstructuredExtractor{},
	}
}

func (l *DocsTopic) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *DocsTopic) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *DocsTopic) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *DocsTopic) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *DocsTopic", object))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *DocsTopic) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.DocsTopic) {
	gqlDocsTopic, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *DocsTopic"))
		return
	}
	if gqlDocsTopic == nil {
		return
	}

	event := gqlschema.DocsTopicEvent{
		Type:      eventType,
		DocsTopic: *gqlDocsTopic,
	}

	l.channel <- event
}
