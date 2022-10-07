package jetstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/beb"

	"github.com/nats-io/nats.go"
)

const (
	natsBackend           = "nats"
	jestreamHandlerName   = "jetstream-handler"
	noSpaceLeftErrMessage = "no space left on device"
)

// compile time check
var _ sender.GenericSender = &Sender{}
var _ health.Checker = &Sender{}

var (
	ErrNotConnected = errors.New("connection status: no connection to NATS JetStream server")
)

// Sender is responsible for sending messages over HTTP.
type Sender struct {
	ctx        context.Context
	logger     *logger.Logger
	connection *nats.Conn
	envCfg     *env.NATSConfig
}

// NewSender returns a new NewSender instance with the given NATS connection.
func NewSender(ctx context.Context, connection *nats.Conn, envCfg *env.NATSConfig, logger *logger.Logger) *Sender {
	return &Sender{ctx: ctx, connection: connection, envCfg: envCfg, logger: logger}
}

// ConnectionStatus returns nats.Status for the NATS connection used by the Sender.
func (s *Sender) ConnectionStatus() nats.Status {
	return s.connection.Status()
}

// Send dispatches the event to the NATS backend in JetStream mode.
// If the NATS connection is not open, it returns an error.
func (s *Sender) Send(_ context.Context, event *event.Event) (sender.PublishResult, error) {
	if s.ConnectionStatus() != nats.CONNECTED {
		return nil, ErrNotConnected
	}
	// ensure the stream exists
	streamExists, err := s.streamExists(s.connection)
	if err != nil {
		return nil, err
	}
	if !streamExists {
		return nil, nats.ErrStreamNotFound
	}

	jsCtx, jsError := s.connection.JetStream()
	if jsError != nil {
		return nil, jsError
	}
	msg, err := s.eventToNATSMsg(event)
	if err != nil {
		return nil, err
	}

	// send the event
	_, err = jsCtx.PublishMsg(msg)
	if err != nil {
		s.namedLogger().Errorw("Cannot send event to backend", "error", err)
		if strings.Contains(err.Error(), noSpaceLeftErrMessage) {
			return nil, err
		}
		return nil, err
	}
	return beb.HTTPPublishResult{Status: http.StatusNoContent}, nil
}

// streamExists checks if the stream with the expected name exists.
func (s *Sender) streamExists(connection *nats.Conn) (bool, error) {
	jsCtx, err := connection.JetStream()
	if err != nil {
		return false, err
	}
	if info, err := jsCtx.StreamInfo(s.envCfg.JSStreamName); err == nil {
		s.namedLogger().Infof("Stream %s exists, using it for publishing", info.Config.Name)
		return true, nil
	} else if !errors.Is(err, nats.ErrStreamNotFound) {
		s.namedLogger().Debug("The connection to NATS server is not established!")
		return false, err
	}
	return false, nats.ErrStreamNotFound
}

// eventToNATSMsg translates cloud event into the NATS Msg.
func (s *Sender) eventToNATSMsg(event *event.Event) (*nats.Msg, error) {
	header := make(nats.Header)
	header.Set(internal.HeaderContentType, event.DataContentType())
	header.Set(internal.CeSpecVersionHeader, event.SpecVersion())
	header.Set(internal.CeTypeHeader, event.Type())
	header.Set(internal.CeSourceHeader, event.Source())
	header.Set(internal.CeIDHeader, event.ID())

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return &nats.Msg{
		Subject: s.getJsSubjectToPublish(event.Type()),
		Header:  header,
		Data:    eventJSON,
	}, err
}

// getJsSubjectToPublish appends stream name to subject if needed.
func (s *Sender) getJsSubjectToPublish(subject string) string {
	return fmt.Sprintf("%s.%s", env.JetStreamSubjectPrefix, subject)
}

func (s *Sender) namedLogger() *zap.SugaredLogger {
	return s.logger.WithContext().Named(jestreamHandlerName).With("backend", natsBackend, "jetstream enabled", true)
}
