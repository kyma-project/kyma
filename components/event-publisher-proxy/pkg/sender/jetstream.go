package sender

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
	"github.com/nats-io/nats.go"
)

const (
	natsBackend           = "nats"
	jetStreamHandlerName  = "jetstream-handler"
	noSpaceLeftErrMessage = "no space left on device"
)

// Compile time check of interface implementation.
var _ GenericSender = &JetStreamMessageSender{}

// JetStreamMessageSender is responsible for sending messages over HTTP.
type JetStreamMessageSender struct {
	ctx        context.Context
	logger     *logger.Logger
	connection *nats.Conn
	envCfg     *env.NatsConfig
}

// NewJetStreamMessageSender returns a new NewJetStreamMessageSender instance with the given NATS connection.
func NewJetStreamMessageSender(ctx context.Context, connection *nats.Conn, envCfg *env.NatsConfig, logger *logger.Logger) *JetStreamMessageSender {
	return &JetStreamMessageSender{ctx: ctx, connection: connection, envCfg: envCfg, logger: logger}
}

// URL returns the URL of the Sender's connection.
func (s *JetStreamMessageSender) URL() string {
	return s.connection.ConnectedUrl()
}

// ConnectionStatus returns nats.Status for the NATS connection used by the JetStreamMessageSender.
func (s *JetStreamMessageSender) ConnectionStatus() nats.Status {
	return s.connection.Status()
}

// Send dispatches the event to the NATS backend in JetStream mode.
// If the NATS connection is not open, it returns an error.
func (s *JetStreamMessageSender) Send(_ context.Context, event *event.Event) (int, error) {
	if s.ConnectionStatus() != nats.CONNECTED {
		return http.StatusBadGateway, errors.New("connection status: no connection to NATS Jetstream server")
	}
	// ensure the stream exists
	streamExists, err := s.streamExists(s.connection)
	if err != nil && err != nats.ErrStreamNotFound {
		return http.StatusInternalServerError, err
	}
	if !streamExists {
		return http.StatusNotFound, nats.ErrStreamNotFound
	}

	jsCtx, jsError := s.connection.JetStream()
	if jsError != nil {
		return http.StatusInternalServerError, jsError
	}
	msg, err := s.eventToNatsMsg(event)
	if err != nil {
		return http.StatusUnprocessableEntity, err
	}

	// send the event
	_, err = jsCtx.PublishMsg(msg)
	if err != nil {
		s.namedLogger().Errorw("Cannot send event to backend", "error", err)
		if strings.Contains(err.Error(), noSpaceLeftErrMessage) {
			return http.StatusInsufficientStorage, err
		}
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

// streamExists checks if a stream with the expected name exists.
func (s *JetStreamMessageSender) streamExists(connection *nats.Conn) (bool, error) {
	jsCtx, err := connection.JetStream()
	if err != nil {
		return false, err
	}
	if info, err := jsCtx.StreamInfo(s.envCfg.JSStreamName); err == nil {
		s.namedLogger().Infof("Stream %s exists, using it for publishing", info.Config.Name)
		return true, nil
	} else if err != nats.ErrStreamNotFound {
		s.namedLogger().Debug("The connection to NATS server is not established!")
		return false, err
	}
	return false, nats.ErrStreamNotFound
}

// eventToNatsMsg translates cloud event into the NATS Msg.
func (s *JetStreamMessageSender) eventToNatsMsg(event *event.Event) (*nats.Msg, error) {
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
func (s *JetStreamMessageSender) getJsSubjectToPublish(subject string) string {
	return fmt.Sprintf("%s.%s", env.JetstreamSubjectPrefix, subject)
}

func (s *JetStreamMessageSender) namedLogger() *zap.SugaredLogger {
	return s.logger.WithContext().Named(jetStreamHandlerName).With("backend", natsBackend, "jetstream enabled", true)
}
