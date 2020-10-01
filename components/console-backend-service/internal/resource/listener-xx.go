package resource

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type EventHandlerProviderXX func() EventHandlerXX

type EventHandlerXX interface {
	K8sResource() runtime.Object
	ShouldNotify() bool
	Notify(gqlschema.SubscriptionEventType)
}

type ListenerXX struct {
	handler EventHandlerProviderXX
}

func NewListenerXX(handler EventHandlerProviderXX) *ListenerXX {
	return &ListenerXX{
		handler: handler,
	}
}

func (l *ListenerXX) OnAdd(object runtime.Object) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ListenerXX) OnUpdate(oldObject, newObject runtime.Object) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ListenerXX) OnDelete(object runtime.Object) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ListenerXX) onEvent(eventType gqlschema.SubscriptionEventType, object runtime.Object) {
	eventHandler := l.handler()
	panic("TODO")
	//eventHandler.res = object

	if eventHandler.ShouldNotify() {
		eventHandler.Notify(eventType)
	}
}
