package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

//go:generate mockery -name=appConverter -output=automock -outpkg=automock -case=underscore
type appConverter interface {
	ToGQL(in *api.Application) gqlschema.Application
}

type Application struct {
	channel   chan<- gqlschema.ApplicationEvent
	converter appConverter
}

func NewApplication(channel chan<- gqlschema.ApplicationEvent, converter appConverter) *Application {
	return &Application{
		channel:   channel,
		converter: converter,
	}
}

func (l *Application) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Application) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Application) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Application) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	app, ok := object.(*api.Application)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Application", object))
		return
	}

	l.notify(eventType, app)
}

func (l *Application) notify(eventType gqlschema.SubscriptionEventType, application *api.Application) {
	gqlApplication := l.converter.ToGQL(application)

	event := gqlschema.ApplicationEvent{
		Type:        eventType,
		Application: gqlApplication,
	}

	l.channel <- event
}
