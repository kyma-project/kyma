package listener

import (
	"fmt"

	"github.com/golang/glog"
	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterServiceBrokerConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterServiceBrokerConverter interface {
	ToGQL(in *api.ClusterServiceBroker) (*gqlschema.ClusterServiceBroker, error)
}

type ClusterServiceBroker struct {
	channel   chan<- *gqlschema.ClusterServiceBrokerEvent
	filter    func(entity *api.ClusterServiceBroker) bool
	converter gqlClusterServiceBrokerConverter
}

func NewClusterServiceBroker(channel chan<- *gqlschema.ClusterServiceBrokerEvent, filter func(entity *api.ClusterServiceBroker) bool, converter gqlClusterServiceBrokerConverter) *ClusterServiceBroker {
	return &ClusterServiceBroker{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *ClusterServiceBroker) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *ClusterServiceBroker) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}

func (l *ClusterServiceBroker) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *ClusterServiceBroker) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	entity, ok := object.(*api.ClusterServiceBroker)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *ClusterServiceBroker", object))
		return
	}

	if l.filter(entity) {
		l.notify(eventType, entity)
	}
}

func (l *ClusterServiceBroker) notify(eventType gqlschema.SubscriptionEventType, entity *api.ClusterServiceBroker) {
	gqlClusterServiceBroker, err := l.converter.ToGQL(entity)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *ClusterServiceBroker"))
		return
	}
	if gqlClusterServiceBroker == nil {
		return
	}

	event := &gqlschema.ClusterServiceBrokerEvent{
		Type:                 eventType,
		ClusterServiceBroker: gqlClusterServiceBroker,
	}

	l.channel <- event
}
