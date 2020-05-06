package serverless

import (
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
)

type functionListener struct {
	channel   chan<- gqlschema.FunctionEvent
	filter    func(entity *v1alpha1.Function) bool
	converter gqlFunctionConverter
	extractor *functionUnstructuredExtractor
}

func newFunctionListener(channel chan<- gqlschema.FunctionEvent, filter func(entity *v1alpha1.Function) bool, converter gqlFunctionConverter) *functionListener {
	return &functionListener{
		channel:   channel,
		filter:    filter,
		converter: converter,
		extractor: newFunctionUnstructuredExtractor(),
	}
}

func (l *functionListener) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *functionListener) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *functionListener) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *functionListener) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, err := l.extractor.do(object)
	if err != nil {
		glog.Error(errors.Wrapf(err, "incorrect object type: %T, should be: *%s", object, pretty.FunctionType))
		return
	}
	if entity == nil {
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *functionListener) notify(eventType gqlschema.SubscriptionEventType, entity *v1alpha1.Function) {
	gqlFunction, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *%s", pretty.FunctionType))
		return
	}
	if gqlFunction == nil {
		return
	}

	event := gqlschema.FunctionEvent{
		Type:     eventType,
		Function: *gqlFunction,
	}

	l.channel <- event
}
