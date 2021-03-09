package sender

import (
	"bytes"
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	"fmt"
)

// compile time check
var _ GenericSender = &NatsMessageSender{}

type GenericSender interface {
	Send(context.Context, *cev2event.Event) (int, error)
}

// HttpMessageSender is responsible for sending messages over HTTP.
type NatsMessageSender struct {
	ctx        context.Context
	logger     *logrus.Logger
	connection *nats.Conn
}

// NewNatsMessageSender returns a new NewNatsMessageSender instance with the given nats connection.
func NewNatsMessageSender(ctx context.Context, connection *nats.Conn, logger *logrus.Logger) *NatsMessageSender {
	return &NatsMessageSender{ctx: ctx, connection: connection, logger: logger}
}

// Send dispatches the given Cloud Event to NATS and returns the response details and dispatch time.
func (h *NatsMessageSender) Send(ctx context.Context, event *cev2event.Event) (int, error) {

	b := new(bytes.Buffer)
	for key, value := range event.Extensions() {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}

	log.Infof("Sending event to NATS, id:[%s], context[%s], extensions[%s]", event.ID(), event.Context.String(), b.String())

	// The same Nats subject used by Nats subscription
	subject := event.Type()

	sender, err := cenats.NewSenderFromConn(h.connection, subject)
	if err != nil {
		h.logger.Errorf("Failed to create nats protocol, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	client, err := cev2.NewClient(sender)
	if err != nil {
		h.logger.Errorf("Failed to create client, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	err = client.Send(ctx, *event)
	if cev2.IsUndelivered(err) {
		h.logger.Errorf("Failed to send: %s", err.Error())
		return http.StatusBadGateway, err
	}

	h.logger.Infof("sent id:[%s], accepted: %t", event.ID(), cev2.IsACK(err))
	return http.StatusNoContent, nil
}
