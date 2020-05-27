package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=appConverter -output=automock -outpkg=automock -case=underscore
type appConverter interface {
	ToGQL(in *api.Application) *gqlschema.Application
}

//go:generate mockery -name=extractor -output=automock -outpkg=automock -case=underscore
type extractor interface {
	Do(interface{}) (*api.Application, error)
}

type Application struct {
	channel   chan<- *gqlschema.ApplicationEvent
	converter appConverter
	extractor extractor
}

func NewApplication(channel chan<- *gqlschema.ApplicationEvent, converter appConverter, extractor extractor) *Application {
	return &Application{
		channel:   channel,
		converter: converter,
		extractor: extractor,
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
	convertedApp, err := l.extractor.Do(object)
	if err != nil {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Application", object))
		return
	}

	if convertedApp == nil {
		return
	}

	l.notify(eventType, convertedApp)
}

func (l *Application) notify(eventType gqlschema.SubscriptionEventType, application *api.Application) {
	gqlApplication := l.converter.ToGQL(application)

	event := &gqlschema.ApplicationEvent{
		Type:        eventType,
		Application: gqlApplication,
	}

	l.channel <- event
}
