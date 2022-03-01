package sender

import (
	"context"
	"errors"
	"net/http"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
)

// compile time check
var _ GenericSender = &NatsMessageSender{}

type GenericSender interface {
	Send(context.Context, *event.Event) (int, error)
}

// NatsMessageSender is responsible for sending messages over HTTP.
type NatsMessageSender struct {
	ctx               context.Context
	logger            *logrus.Logger
	backendConnection *pkgnats.BackendConnection
}

// NewNatsMessageSender returns a new NewNatsMessageSender instance with the given nats connection.
func NewNatsMessageSender(ctx context.Context, bc *pkgnats.BackendConnection, logger *logrus.Logger) *NatsMessageSender {
	return &NatsMessageSender{ctx: ctx, backendConnection: bc, logger: logger}
}

// ConnectionStatus returns nats.Status for the NATS connection used by the NatsMessageSender.
func (s *NatsMessageSender) ConnectionStatus() nats.Status {
	return s.backendConnection.Connection.Status()
}

// Send dispatches the event.Event to NATS server.
// If the NATS connection is not open, it returns an error.
func (s *NatsMessageSender) Send(ctx context.Context, event *event.Event) (int, error) {
	if s.ConnectionStatus() != nats.CONNECTED {
		return http.StatusBadGateway, errors.New("connection status: no connection to NATS server")
	}

	sender, err := cenats.NewSenderFromConn(s.backendConnection.Connection, event.Type())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	client, err := cev2.NewClient(sender)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := client.Send(ctx, *event); cev2.IsUndelivered(err) {
		return http.StatusBadGateway, err
	}

	return http.StatusNoContent, nil
}
