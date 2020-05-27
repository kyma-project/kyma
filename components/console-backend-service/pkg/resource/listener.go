package resource

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
)

type Handler func() EventHandler

type EventHandler interface {
	K8sResource() interface{}
	ShouldNotify() bool
	Notify(model.EventType)
}

type Listener struct {
	handler Handler
}

func NewListener(handler Handler) *Listener {
	return &Listener{
		handler: handler,
	}
}

func (l *Listener) OnAdd(object interface{}) {
	l.onEvent(model.EventTypeAdd, object)
}

func (l *Listener) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(model.EventTypeUpdate, newObject)
}

func (l *Listener) OnDelete(object interface{}) {
	l.onEvent(model.EventTypeDelete, object)
}

func (l *Listener) onEvent(eventType model.EventType, object interface{}) {
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
