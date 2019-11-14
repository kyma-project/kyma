package listener

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterAssetConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterAssetConverter interface {
	ToGQL(in *v1beta1.ClusterAsset) (*gqlschema.RafterClusterAsset, error)
}

type ClusterAsset struct {
	channel   chan<- gqlschema.RafterClusterAssetEvent
	filter    func(entity *v1beta1.ClusterAsset) bool
	converter gqlClusterAssetConverter
	extractor extractor.ClusterAssetUnstructuredExtractor
}

func NewClusterAsset(channel chan<- gqlschema.RafterClusterAssetEvent, filter func(entity *v1beta1.ClusterAsset) bool, converter gqlClusterAssetConverter) *ClusterAsset {
	return &ClusterAsset{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.ClusterAssetUnstructuredExtractor{},
	}
}

func (l *ClusterAsset) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ClusterAsset) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ClusterAsset) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ClusterAsset) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(errors.Wrapf(err, "incorrect object type: %T, should be: *%s", object, pretty.ClusterAssetType))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ClusterAsset) notify(eventType gqlschema.SubscriptionEventType, entity *v1beta1.ClusterAsset) {
	gqlClusterAsset, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *%s", pretty.ClusterAssetType))
		return
	}
	if gqlClusterAsset == nil {
		return
	}

	event := gqlschema.RafterClusterAssetEvent{
		Type:         eventType,
		ClusterAsset: *gqlClusterAsset,
	}

	l.channel <- event
}
