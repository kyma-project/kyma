package listener

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlApiRuleConverter -output=automock -outpkg=automock -case=underscore
type gqlApiRuleConverter interface {
	ToGQL(in *v1alpha1.APIRule) (*gqlschema.APIRule, error)
}

//go:generate mockery -name=extractor -output=automock -outpkg=automock -case=underscore
type extractor interface {
	Do(interface{}) (*v1alpha1.APIRule, error)
}

type ApiRuleListener struct {
	channel   chan<- gqlschema.ApiRuleEvent
	filter    func(api *v1alpha1.APIRule) bool
	converter gqlApiRuleConverter
	extractor extractor
}

func NewApiRule(channel chan<- gqlschema.ApiRuleEvent, filter func(api *v1alpha1.APIRule) bool, converter gqlApiRuleConverter, extractor extractor) *ApiRuleListener {
	return &ApiRuleListener{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: extractor,
	}
}

func (l *ApiRuleListener) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ApiRuleListener) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ApiRuleListener) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ApiRuleListener) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	convertedApiRule, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ApiRule", object))
		return
	}

	if convertedApiRule == nil {
		return
	}

	if l.filter(convertedApiRule) {
		l.notify(eventType, convertedApiRule)
	}
}

func (l *ApiRuleListener) notify(eventType gqlschema.SubscriptionEventType, apiRule *v1alpha1.APIRule) {
	gqlApiRule, err := l.converter.ToGQL(apiRule)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *APIRule"))
		return
	}

	if gqlApiRule == nil {
		return
	}

	event := gqlschema.ApiRuleEvent{
		Type:    eventType,
		APIRule: *gqlApiRule,
	}

	l.channel <- event
}
