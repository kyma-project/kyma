package natsv2

import (
	"github.com/nats-io/nats.go"
)

type ConnClosedHandler func(conn *nats.Conn)

type Subscriber interface {
	SubscriptionSubject() string
	ConsumerInfo() (*nats.ConsumerInfo, error)
	IsValid() bool
	Unsubscribe() error
	SetPendingLimits(int, int) error
	PendingLimits() (int, int, error)
}

type Subscription struct {
	*nats.Subscription
}

func (js Subscription) SubscriptionSubject() string {
	return js.Subject
}
