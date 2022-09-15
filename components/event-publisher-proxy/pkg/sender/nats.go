package sender

import (
	"context"
	"errors"
	"net/http"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
)

// compile time check
var _ GenericSender = &NATSMessageSender{}

type GenericSender interface {
	Send(context.Context, *event.Event) (int, error)
	ConnectionStatus() nats.Status
	URL() string
}

// NATSMessageSender is responsible for sending messages over HTTP.
type NATSMessageSender struct {
	ctx        context.Context
	logger     *logger.Logger
	connection *nats.Conn
}

// NewNATSMessageSender returns a new NewNATSMessageSender instance with the given nats connection.
func NewNATSMessageSender(ctx context.Context, connection *nats.Conn, logger *logger.Logger) *NATSMessageSender {
	return &NATSMessageSender{ctx: ctx, connection: connection, logger: logger}
}

// URL returns the URL of the Sender's connection.
func (s *NATSMessageSender) URL() string {
	return s.connection.ConnectedUrl()
}

// ConnectionStatus returns nats.Status for the NATS connection used by the NATSMessageSender.
func (s *NATSMessageSender) ConnectionStatus() nats.Status {
	return s.connection.Status()
}

// Send dispatches the event.Event to NATS server.
// If the NATS connection is not open, it returns an error.
func (s *NATSMessageSender) Send(ctx context.Context, event *event.Event) (int, error) {
	if s.ConnectionStatus() != nats.CONNECTED {
		return http.StatusBadGateway, errors.New("connection status: no connection to NATS server")
	}

	sender, err := cenats.NewSenderFromConn(s.connection, event.Type())
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
