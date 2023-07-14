package jetstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/common"
)

const (
	JSStoreFailedCode     = 10077
	natsBackend           = "nats"
	handlerName           = "jetstream-handler"
	noSpaceLeftErrMessage = "no space left on device"
)

// compile time check.
var _ sender.GenericSender = &Sender{}
var _ health.Checker = &Sender{}

//nolint:lll // reads better this way
var (
	ErrNotConnected        = common.BackendPublishError{HTTPCode: http.StatusBadGateway, Info: "no connection to NATS JetStream server"}
	ErrCannotSendToStream  = common.BackendPublishError{HTTPCode: http.StatusGatewayTimeout, Info: "cannot send to stream"}
	ErrNoSpaceLeftOnDevice = common.BackendPublishError{HTTPCode: http.StatusInsufficientStorage, Info: "insufficient resources on target stream"}
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

// ConnectionStatus returns nats.code for the NATS connection used by the Sender.
func (s *Sender) ConnectionStatus() nats.Status {
	return s.connection.Status()
}

// Send dispatches the event to the NATS backend in JetStream mode.
// If the NATS connection is not open, it returns an error.
func (s *Sender) Send(_ context.Context, event *event.Event) sender.PublishError {
	if s.ConnectionStatus() != nats.CONNECTED {
		return ErrNotConnected
	}

	jsCtx, err := s.connection.JetStream()
	if err != nil {
		s.namedLogger().Error("error", err)
		return common.ErrClientNoConnection
	}

	msg, err := s.eventToNATSMsg(event)
	if err != nil {
		s.namedLogger().Error("error", err)
		e := common.ErrClientConversionFailed
		e.Wrap(err)
		return e
	}

	// send the event
	_, err = jsCtx.PublishMsg(msg)
	if err != nil {
		s.namedLogger().Errorw("Cannot send event to backend", "error", err)
		return natsErrorToPublishError(err)
	}
	return nil
}

func natsErrorToPublishError(err error) sender.PublishError {
	if errors.Is(err, nats.ErrNoStreamResponse) {
		return ErrCannotSendToStream
	}

	if strings.Contains(err.Error(), noSpaceLeftErrMessage) {
		return ErrNoSpaceLeftOnDevice
	}

	var apiErr nats.JetStreamError
	e := common.BackendPublishError{HTTPCode: http.StatusInternalServerError}
	if errors.As(err, &apiErr) {
		if apiErr.APIError().ErrorCode == JSStoreFailedCode {
			return ErrNoSpaceLeftOnDevice
		}
		e.HTTPCode = apiErr.APIError().Code
		e.Info = apiErr.APIError().Description
		e.Wrap(err)
		return e
	}
	return common.ErrInternalBackendError
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
	// do not append prefix, if event type prefix is not present.
	if !strings.HasPrefix(subject, s.envCfg.EventTypePrefix) {
		return subject
	}

	// append prefix, for v1alpha1 subscriptions.
	return fmt.Sprintf("%s.%s", env.JetStreamSubjectPrefix, subject)
}

func (s *Sender) namedLogger() *zap.SugaredLogger {
	return s.logger.WithContext().Named(handlerName).With("backend", natsBackend, "jetstream enabled", true)
}
