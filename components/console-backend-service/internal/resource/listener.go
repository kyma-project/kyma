package resource

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/golang/glog"
)

type EventHandlerProvider func() EventHandler

type EventType string
const (
	EventTypeAdd = "ADD"
	EventTypeUpdate = "UPDATE"
	EventTypeDelete = "DELETE"
)

type EventHandler interface {
	K8sResource() interface{}
	ShouldNotify() bool
	Notify(EventType)
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
	l.onEvent(EventTypeAdd, object)
}

func (l *Listener) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(EventTypeUpdate, newObject)
}

func (l *Listener) OnDelete(object interface{}) {
	l.onEvent(EventTypeDelete, object)
}

func (l *Listener) onEvent(eventType EventType, object interface{}) {
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