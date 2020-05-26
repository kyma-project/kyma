package ui

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/pkg/errors"
)

type BackendModuleListener struct {
	channel   chan<- *model.BackendModuleEvent
	filter    func(pod *v1alpha1.BackendModule) bool
	converter BackendModuleConverter
}

func NewBackendModuleListener(channel chan<- *model.BackendModuleEvent, filter func(pod *v1alpha1.BackendModule) bool) *BackendModuleListener {
	return &BackendModuleListener{
		channel:   channel,
		filter:    filter,
		converter: BackendModuleConverter{},
	}
}

func (l *BackendModuleListener) OnAdd(object interface{}) {
	l.onEvent(model.EventTypeAdd, object)
}

func (l *BackendModuleListener) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(model.EventTypeUpdate, newObject)
}

func (l *BackendModuleListener) OnDelete(object interface{}) {
	l.onEvent(model.EventTypeDelete, object)
}

func (l *BackendModuleListener) onEvent(eventType model.EventType, object interface{}) {
	u, ok := object.(*unstructured.Unstructured)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *unstructured.Unstructured", object))
		return
	}

	result := &v1alpha1.BackendModule{}
	err := resource.FromUnstructured(u, result)
	if err != nil {
		glog.Error(fmt.Errorf("cannot convert to unstructured: %v", object))
		return
	}

	if l.filter(result) {
		l.notify(eventType, result)
	}
}

func (l *BackendModuleListener) notify(eventType model.EventType, obj *v1alpha1.BackendModule) {
	gqlPod, err := l.converter.ToGQL(obj)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *BackendModule"))
		return
	}
	if gqlPod == nil {
		return
	}

	event := &model.BackendModuleEvent{
		Type:     &eventType,
		Resource: gqlPod,
	}

	l.channel <- event
}
