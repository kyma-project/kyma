package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlPodConverter -output=automock -outpkg=automock -case=underscore
type gqlPodConverter interface {
	ToGQL(in *v1.Pod) (*gqlschema.Pod, error)
}

type Pod struct {
	channel   chan<- *gqlschema.PodEvent
	filter    func(pod *v1.Pod) bool
	converter gqlPodConverter
}

func NewPod(channel chan<- *gqlschema.PodEvent, filter func(pod *v1.Pod) bool, converter gqlPodConverter) *Pod {
	return &Pod{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *Pod) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Pod) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Pod) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Pod) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	pod, ok := object.(*v1.Pod)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Pod", object))
		return
	}

	if l.filter(pod) {
		l.notify(eventType, pod)
	}
}

func (l *Pod) notify(eventType gqlschema.SubscriptionEventType, pod *v1.Pod) {
	gqlPod, err := l.converter.ToGQL(pod)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *Pod"))
		return
	}
	if gqlPod == nil {
		return
	}

	event := &gqlschema.PodEvent{
		Type: eventType,
		Pod:  gqlPod,
	}

	l.channel <- event
}
