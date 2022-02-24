package sender

import (
	"context"
	"github.com/nats-io/nats.go"
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

// NewNatsMessageSender returns a new NewNatsMessageSender instance with the given nats connection.
func NewNatsMessageSender(ctx context.Context, bc *pkgnats.BackendConnection, logger *logrus.Logger) *NatsMessageSender {
	return &NatsMessageSender{ctx: ctx, backendConnection: bc, logger: logger}
}

func (s *NatsMessageSender) ConnectionStatus() nats.Status {
	return s.backendConnection.Connection.Status()
}

// Send dispatches the given event to NATS and returns the response details and dispatch time.
func (s *NatsMessageSender) Send(ctx context.Context, event *cev2event.Event) (int, error) {
	sender, err := cenats.NewSenderFromConn(s.backendConnection.Connection, event.Type())
	if err != nil {
		s.logger.Errorf("Failed to create NATS sender, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	client, err := cev2.NewClient(sender)
	if err != nil {
		s.logger.Errorf("Failed to create cloudevents client, %s", err.Error())
		return http.StatusInternalServerError, err
	}

	if err := client.Send(ctx, *event); cev2.IsUndelivered(err) {
		s.logger.Errorf("Failed to send cloudevent, %s", err.Error())
		return http.StatusBadGateway, err
	}

	return http.StatusNoContent, nil
}
