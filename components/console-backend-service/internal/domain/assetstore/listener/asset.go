package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlAssetConverter -output=automock -outpkg=automock -case=underscore
type gqlAssetConverter interface {
	ToGQL(in *v1alpha2.Asset) (*gqlschema.Asset, error)
}

type Asset struct {
	channel   chan<- gqlschema.AssetEvent
	filter    func(entity *v1alpha2.Asset) bool
	converter gqlAssetConverter
	extractor extractor.AssetUnstructuredExtractor
}

func NewAsset(channel chan<- gqlschema.AssetEvent, filter func(entity *v1alpha2.Asset) bool, converter gqlAssetConverter) *Asset {
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
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Asset", object))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *Asset) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha2.Asset) {
	gqlAsset, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *Asset"))
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
