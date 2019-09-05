package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlApiConverter -output=automock -outpkg=automock -case=underscore
type gqlApiConverter interface {
	ToGQL(in *v1alpha2.Api) *gqlschema.API
}

//go:generate mockery -name=extractor -output=automock -outpkg=automock -case=underscore
type extractor interface {
	Do(interface{}) (*v1alpha2.Api, error)
}

type Api struct {
	channel   chan<- gqlschema.ApiEvent
	filter    func(api *v1alpha2.Api) bool
	converter gqlApiConverter
	extractor extractor
}

func NewApi(channel chan<- gqlschema.ApiEvent, filter func(api *v1alpha2.Api) bool, converter gqlApiConverter, extractor extractor) *Api {
	return &Api{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor,
	}
}

func (l *Api) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Api) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Api) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Api) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	convertedApi, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Api", object))
		return
	}

	if convertedApi == nil {
		return
	}

	if l.filter(convertedApi) {
		l.notify(eventType, convertedApi)
	}
}

func (l *Api) notify(eventType gqlschema.SubscriptionEventType, api *v1alpha2.Api) {
	gqlApi := l.converter.ToGQL(api)
	if gqlApi == nil {
		return
	}

	event := gqlschema.ApiEvent{
		Type: eventType,
		API:  *gqlApi,
	}

	l.channel <- event
}
