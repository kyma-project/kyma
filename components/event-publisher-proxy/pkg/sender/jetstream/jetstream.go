package jetstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
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
	opts       *options.Options
}

func (s *Sender) URL() string {
	return s.envCfg.URL
}

// NewSender returns a new NewSender instance with the given NATS connection.
func NewSender(ctx context.Context, connection *nats.Conn, envCfg *env.NATSConfig, opts *options.Options, logger *logger.Logger) *Sender {
	return &Sender{ctx: ctx, connection: connection, envCfg: envCfg, opts: opts, logger: logger}
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
	// if subscription CRD v1alpha2 is enabled then do not append prefix.
	if s.opts.EnableNewCRDVersion && !strings.HasPrefix(subject, s.envCfg.EventTypePrefix) {
		return subject
	}

	return fmt.Sprintf("%s.%s", env.JetStreamSubjectPrefix, subject)
}

func (s *Sender) namedLogger() *zap.SugaredLogger {
	return s.logger.WithContext().Named(jestreamHandlerName).With("backend", natsBackend, "jetstream enabled", true)
}
