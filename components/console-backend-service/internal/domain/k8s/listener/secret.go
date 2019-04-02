package listener

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

//go:generate mockery -name=gqlSecretConverter -output=automock -outpkg=automock -case=underscore
type gqlSecretConverter interface {
	ToGQL(in *v1.Secret) (*gqlschema.Secret, error)
}

type Secret struct {
	channel   chan<- gqlschema.SecretEvent
	filter    func(secret *v1.Secret) bool
	converter gqlSecretConverter
}

func NewSecret(channel chan<- gqlschema.SecretEvent, filter func(secret *v1.Secret) bool, converter gqlSecretConverter) *Secret {
	return &Secret{
		channel:   channel,
		filter:    filter,
		converter: converter,
	}
}

func (l *Secret) OnAdd(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeAdd, object)
}

func (l *Secret) OnUpdate(oldObject, newObject interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeUpdate, newObject)
}
func (l *Secret) OnDelete(object interface{}) {
	l.onEvent(gqlschema.SubscriptionEventTypeDelete, object)
}

func (l *Secret) onEvent(eventType gqlschema.SubscriptionEventType, object interface{}) {
	secret, ok := object.(*v1.Secret)
	if !ok {
		glog.Error(fmt.Errorf("incorrect object type: %T, should be: *Secret", object))
		return
	}

	if l.filter(secret) {
		l.notify(eventType, secret)
	}
}

func (l *Secret) notify(eventType gqlschema.SubscriptionEventType, secret *v1.Secret) {
	gqlSecret, err := l.converter.ToGQL(secret)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting *Secret"))
		return
	}
	if gqlSecret == nil {
		return
	}

	event := gqlschema.SecretEvent{
		Type:   eventType,
		Secret: *gqlSecret,
	}

	l.channel <- event
}
