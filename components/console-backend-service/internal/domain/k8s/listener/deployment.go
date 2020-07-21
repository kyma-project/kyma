package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "k8s.io/api/apps/v1"
)

//go:generate mockery -name=gqlDeploymentConverter -output=automock -outpkg=automock -case=underscore
type gqlDeploymentConverter interface {
	ToGQL(in *api.Deployment) *gqlschema.Deployment
}

type Deployment struct {
	channel   chan<- *gqlschema.DeploymentEvent
	filter    func(deployment *api.Deployment) bool
	converter gqlDeploymentConverter
}

func NewDeployment(channel chan<- *gqlschema.DeploymentEvent, filter func(deployment *api.Deployment) bool, converter gqlDeploymentConverter) *Deployment {
	return &Deployment{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *Deployment) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Deployment) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *Deployment) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Deployment) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	deployment, ok := object.(*api.Deployment)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Deployment", object))
		return
	}

	if l.filter(deployment) {
		l.notify(eventType, deployment)
	}
}

func (l *Deployment) notify(eventType gqlschema.SubscriptionEventType, deployment *api.Deployment) {
	gqlDeployment := l.converter.ToGQL(deployment)
	if gqlDeployment == nil {
		return
	}

	event := &gqlschema.DeploymentEvent{
		Type:       eventType,
		Deployment: gqlDeployment,
	}

	l.channel <- event
}
