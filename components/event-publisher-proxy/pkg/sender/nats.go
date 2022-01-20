package sender

import (
	"context"
	"net/http"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	"github.com/sirupsen/logrus"
)

// compile time check
var _ GenericSender = &NatsMessageSender{}

type GenericSender interface {
	Send(context.Context, *cev2event.Event) (int, error)
}

// NatsMessageSender is responsible for sending messages over HTTP.
type NatsMessageSender struct {
	ctx               context.Context
	logger            *logrus.Logger
	backendConnection *pkgnats.BackendConnection
}

// NewNatsMessageSender returns a new NewNatsMessageSender instance with the given NATS connection.
func NewNatsMessageSender(ctx context.Context, bc *pkgnats.BackendConnection, logger *logrus.Logger) *NatsMessageSender {
	return &NatsMessageSender{ctx: ctx, backendConnection: bc, logger: logger}
}

// Send dispatches the given Cloud Event to NATS and returns the response details and dispatch time.
func (h *NatsMessageSender) Send(ctx context.Context, event *cev2event.Event) (int, error) {
	if h.backendConnection.Connection.IsClosed() {
		h.logger.Info("Reconnect to NATS server")
		if err := h.backendConnection.Connect(); err != nil {
			h.logger.Errorf("Failed to reconnect to NATS server, %s", err.Error())
			return http.StatusInternalServerError, err
		}
	}

	sender, err := cenats.NewSenderFromConn(h.backendConnection.Connection, event.Type())
	if err != nil {
		h.logger.Errorf("Failed to create NATS sender, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	client, err := cev2.NewClient(sender)
	if err != nil {
		h.logger.Errorf("Failed to create cloudevents client, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	err = client.Send(ctx, *event)
	if cev2.IsUndelivered(err) {
		h.logger.Errorf("Failed to send cloudevent, %s", err.Error())
		return http.StatusBadGateway, err
	}

	return http.StatusNoContent, nil
}
