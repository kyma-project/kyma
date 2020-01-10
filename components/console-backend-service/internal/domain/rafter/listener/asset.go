package listener

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlAssetConverter -output=automock -outpkg=automock -case=underscore
type gqlAssetConverter interface {
	ToGQL(in *v1beta1.Asset) (*gqlschema.Asset, error)
}

type Asset struct {
	channel   chan<- gqlschema.AssetEvent
	filter    func(entity *v1beta1.Asset) bool
	converter gqlAssetConverter
	extractor extractor.AssetUnstructuredExtractor
}

func NewAsset(channel chan<- gqlschema.AssetEvent, filter func(entity *v1beta1.Asset) bool, converter gqlAssetConverter) *Asset {
	return &Asset{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor.AssetUnstructuredExtractor{},
	}
}

func (l *Asset) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Asset) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Asset) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Asset) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(errors.Wrapf(err, "incorrect object type: %T, should be: *%s", object, pretty.AssetType))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *Asset) notify(eventType gqlschema.SubscriptionEventType, entity *v1beta1.Asset) {
	gqlAsset, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *%s", pretty.AssetType))
		return
	}
	if gqlAsset == nil {
		return
	}

	event := gqlschema.AssetEvent{
		Type:  eventType,
		Asset: *gqlAsset,
	}

	l.channel <- event
}
