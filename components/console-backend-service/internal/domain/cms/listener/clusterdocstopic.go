package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterDocsTopicConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterDocsTopicConverter interface {
	ToGQL(in *v1alpha1.ClusterDocsTopic) (*gqlschema.ClusterDocsTopic, error)
}

type ClusterDocsTopic struct {
	channel   chan<- gqlschema.ClusterDocsTopicEvent
	filter    func(entity *v1alpha1.ClusterDocsTopic) bool
	converter gqlClusterDocsTopicConverter
	extractor extractor.ClusterDocsTopicUnstructuredExtractor
}

func NewClusterDocsTopic(channel chan<- gqlschema.ClusterDocsTopicEvent, filter func(entity *v1alpha1.ClusterDocsTopic) bool, converter gqlClusterDocsTopicConverter) *ClusterDocsTopic {
	return &ClusterDocsTopic{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.ClusterDocsTopicUnstructuredExtractor{},
	}
}

func (l *ClusterDocsTopic) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ClusterDocsTopic) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ClusterDocsTopic) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ClusterDocsTopic) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ClusterDocsTopic", object))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ClusterDocsTopic) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.ClusterDocsTopic) {
	gqlClusterDocsTopic, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ClusterDocsTopic"))
		return
	}
	if gqlClusterDocsTopic == nil {
		return
	}

	event := gqlschema.ClusterDocsTopicEvent{
		Type:             eventType,
		ClusterDocsTopic: *gqlClusterDocsTopic,
	}

	l.channel <- event
}
