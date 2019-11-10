package listener

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"

	"github.com/golang/glog"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlAssetGroupConverter -output=automock -outpkg=automock -case=underscore
type gqlAssetGroupConverter interface {
	ToGQL(in *v1beta1.AssetGroup) (*gqlschema.AssetGroup, error)
}

type AssetGroup struct {
	channel   chan<- gqlschema.AssetGroupEvent
	filter    func(entity *v1beta1.AssetGroup) bool
	converter gqlAssetGroupConverter
	extractor extractor.AssetGroupUnstructuredExtractor
}

func NewAssetGroup(channel chan<- gqlschema.AssetGroupEvent, filter func(entity *v1beta1.AssetGroup) bool, converter gqlAssetGroupConverter) *AssetGroup {
	return &AssetGroup{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.AssetGroupUnstructuredExtractor{},
	}
}

func (l *AssetGroup) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *AssetGroup) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *AssetGroup) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *AssetGroup) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(errors.Wrapf(err, "incorrect object type: %T, should be: *%s", object, pretty.AssetGroupType))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *AssetGroup) notify(eventType gqlschema.SubscriptionEventType, entity *v1beta1.AssetGroup) {
	gqlAssetGroup, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *%s", pretty.AssetGroupType))
		return
	}
	if gqlAssetGroup == nil {
		return
	}

	event := gqlschema.AssetGroupEvent{
		Type:             eventType,
		AssetGroup: 	  *gqlAssetGroup,
	}

	l.channel <- event
}
