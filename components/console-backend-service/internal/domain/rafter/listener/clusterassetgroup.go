package listener

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterAssetGroupConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterAssetGroupConverter interface {
	ToGQL(in *v1beta1.ClusterAssetGroup) (*gqlschema.ClusterAssetGroup, error)
}

type ClusterAssetGroup struct {
	channel   chan<- gqlschema.ClusterAssetGroupEvent
	filter    func(entity *v1beta1.ClusterAssetGroup) bool
	converter gqlClusterAssetGroupConverter
	extractor extractor.ClusterAssetGroupUnstructuredExtractor
}

func NewClusterAssetGroup(channel chan<- gqlschema.ClusterAssetGroupEvent, filter func(entity *v1beta1.ClusterAssetGroup) bool, converter gqlClusterAssetGroupConverter) *ClusterAssetGroup {
	return &ClusterAssetGroup{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.ClusterAssetGroupUnstructuredExtractor{},
	}
}

func (l *ClusterAssetGroup) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ClusterAssetGroup) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ClusterAssetGroup) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ClusterAssetGroup) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(errors.Wrapf(err, "incorrect object type: %T, should be: *%s", object, pretty.ClusterAssetGroupType))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ClusterAssetGroup) notify(eventType gqlschema.SubscriptionEventType, entity *v1beta1.ClusterAssetGroup) {
	gqlClusterAssetGroup, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *%s", pretty.ClusterAssetGroupType))
		return
	}
	if gqlClusterAssetGroup == nil {
		return
	}

	event := gqlschema.ClusterAssetGroupEvent{
		Type:              eventType,
		ClusterAssetGroup: *gqlClusterAssetGroup,
	}

	l.channel <- event
}
