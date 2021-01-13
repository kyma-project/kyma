package sender

import (
	"context"
	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/sirupsen/logrus"
	"net/http"
)

// compile time check
var _ GenericSender = &NatsMessageSender{}

type GenericSender interface {
	Send(context.Context, *cev2event.Event) (int, error)
}

// HttpMessageSender is responsible for sending messages over HTTP.
type NatsMessageSender struct {
	Ctx context.Context
	Target string
	Logger *logrus.Logger
}

// NewNatsMessageSender returns a new NewNatsMessageSender instance with the given target and client.
func NewNatsMessageSender(ctx context.Context, target string, logger *logrus.Logger) *NatsMessageSender {
	return &NatsMessageSender{Ctx: ctx, Target: target, Logger: logger}
}

// Send dispatches the given Cloud Event to NATS and returns the response details and dispatch time.
func (h *NatsMessageSender) Send(ctx context.Context, event *cev2event.Event) (int, error) {
	h.Logger.Infof("Sending event to NATS, id:[%s]", event.ID())
	// The same Nats subject used by Nats subscription
	subject := event.Type()

	p, err := cenats.NewSender(h.Target, subject, cenats.NatsOptions())
	if err != nil {
		h.Logger.Errorf("Failed to create nats protocol, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	c, err := cev2.NewClient(p)
	if err != nil {
		h.Logger.Errorf("Failed to create client, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	err = c.Send(ctxWithCancel, *event);
	if cev2.IsUndelivered(err) {
		h.Logger.Printf("failed to send: %s", err.Error())
		return http.StatusBadGateway, err
	}

	h.Logger.Printf("sent id:[%s], accepted: %t", event.ID(), cev2.IsACK(err))
	return http.StatusNoContent, nil
}