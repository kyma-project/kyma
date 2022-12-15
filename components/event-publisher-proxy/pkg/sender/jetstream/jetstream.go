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

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/builder"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/eventmesh"

	"github.com/nats-io/nats.go"
)

const (
	JSStoreFailedCode     = 10077
	natsBackend           = "nats"
	jestreamHandlerName   = "jetstream-handler"
	noSpaceLeftErrMessage = "no space left on device"
)

// compile time check
var _ sender.GenericSender = &Sender{}
var _ health.Checker = &Sender{}

var (
	ErrNotConnected       = errors.New("no connection to NATS JetStream server")
	ErrCannotSendToStream = errors.New("cannot send to stream")
)

// Sender is responsible for sending messages over HTTP.
type Sender struct {
	ctx        context.Context
	logger     *logger.Logger
	connection *nats.Conn
	envCfg     *env.NATSConfig
}

func (s *Sender) URL() string {
	return s.envCfg.URL
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
func (s *Sender) Send(_ context.Context, event *builder.Event) (sender.PublishResult, error) {
	if s.ConnectionStatus() != nats.CONNECTED {
		return nil, ErrNotConnected
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
		if errors.Is(err, nats.ErrNoStreamResponse) {
			return nil, fmt.Errorf("%w : %v", sender.ErrBackendTargetNotFound, fmt.Errorf("%w, %v", ErrCannotSendToStream, err))
		}

		var apiErr nats.JetStreamError
		if ok := errors.As(err, &apiErr); ok {
			if apiErr.APIError().ErrorCode == JSStoreFailedCode {
				return nil, fmt.Errorf("%w: %v", sender.ErrInsufficientStorage, err)
			}
		}
		if strings.Contains(err.Error(), noSpaceLeftErrMessage) {
			return nil, fmt.Errorf("%w: %v", sender.ErrInsufficientStorage, err)
		}
		return nil, fmt.Errorf("%w : %v", sender.ErrInternalBackendError, fmt.Errorf("%w, %v", ErrCannotSendToStream, err))
	}
	return eventmesh.HTTPPublishResult{Status: http.StatusNoContent}, nil
}

// eventToNATSMsg translates cloud event into the NATS Msg.
func (s *Sender) eventToNATSMsg(event *builder.Event) (*nats.Msg, error) {
	header := make(nats.Header)
	header.Set(internal.HeaderContentType, event.CloudEvent().DataContentType())
	header.Set(internal.CeSpecVersionHeader, event.CloudEvent().SpecVersion())
	header.Set(internal.CeTypeHeader, event.Type())
	header.Set(internal.CeSourceHeader, event.CloudEvent().Source())
	header.Set(internal.CeIDHeader, event.CloudEvent().ID())

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
