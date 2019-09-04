package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlNamespaceConverter -output=automock -outpkg=automock -case=underscore
type gqlNamespaceConverter interface {
	ToGQL(in *v1.Namespace) (*gqlschema.Namespace, error)
}

type Namespace struct {
	channel   chan<- gqlschema.NamespaceEvent
	filter    func(namespace *v1.Namespace) bool
	converter gqlNamespaceConverter
}

func NewNamespace(channel chan<- gqlschema.NamespaceEvent, filter func(namespace *v1.Namespace) bool, converter gqlNamespaceConverter) *Namespace {
	return &Namespace{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *Namespace) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Namespace) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Namespace) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Namespace) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	namespace, ok := object.(*v1.Namespace)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Namespace", object))
		return
	}

	if l.filter(namespace) {
		l.notify(eventType, namespace)
	}
}

func (l *Namespace) notify(eventType gqlschema.SubscriptionEventType, namespace *v1.Namespace) {
	gqlNamespace, err := l.converter.ToGQL(namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *Namespace"))
		return
	}
	if gqlNamespace == nil {
		return
	}

	event := gqlschema.NamespaceEvent{
		Type: eventType,
		Namespace:  *gqlNamespace,
	}

	l.channel <- event
}
