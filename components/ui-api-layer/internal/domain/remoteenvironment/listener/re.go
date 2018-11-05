package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

//go:generate mockery -name=reConverter -output=automock -outpkg=automock -case=underscore
type reConverter interface {
	ToGQL(in *api.RemoteEnvironment) gqlschema.RemoteEnvironment
}

type RemoteEnvironment struct {
	channel   chan<- gqlschema.RemoteEnvironmentEvent
	converter reConverter
}

func NewRemoteEnvironment(channel chan<- gqlschema.RemoteEnvironmentEvent, converter reConverter) *RemoteEnvironment {
	return &RemoteEnvironment{
		channel:   channel,
		converter: converter,
	}
}

func (l *RemoteEnvironment) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *RemoteEnvironment) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *RemoteEnvironment) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *RemoteEnvironment) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	remoteEnvironment, ok := object.(*api.RemoteEnvironment)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *RemoteEnvironment", object))
		return
	}

	l.notify(eventType, remoteEnvironment)
}

func (l *RemoteEnvironment) notify(eventType gqlschema.SubscriptionEventType, remoteEnvironment *api.RemoteEnvironment) {
	gqlRemoteEnvironment := l.converter.ToGQL(remoteEnvironment)

	event := gqlschema.RemoteEnvironmentEvent{
		Type:              eventType,
		RemoteEnvironment: gqlRemoteEnvironment,
	}

	l.channel <- event
}
