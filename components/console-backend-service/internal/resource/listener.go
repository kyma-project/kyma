package resource

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/golang/glog"
)

type EventHandlerProvider func() EventHandler

type EventHandler interface {
	K8sResource() interface{}
	ShouldNotify() bool
	Notify(gqlschema.SubscriptionEventType)
}

type Listener struct {
	handler EventHandlerProvider
}

func NewListener(handler EventHandlerProvider) *Listener {
	return &Listener{
		handler: handler,
	}
}

func (l *Listener) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Listener) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Listener) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Listener) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	u, ok := object.(*unstructured.Unstructured)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *unstructured.Unstructured", object))
		return
	}

	eventHandler := l.handler()
	err := FromUnstructured(u, eventHandler.K8sResource())
	if err != nil {
		glog.Error(fmt.Errorf("cannot convert from unstructured: %v", object))
		return
	}

	if eventHandler.ShouldNotify() {
		eventHandler.Notify(eventType)
	}
}
