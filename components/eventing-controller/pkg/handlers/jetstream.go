package handlers

import (
	"fmt"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var _ MessagingBackend = &JetStream{}

const (
	jsHandlerName = "js-handler"
)

type JetStream struct {
	config env.NatsConfig
	conn   *nats.Conn
	jsCtx  nats.JetStreamContext
	// connClosedHandler gets called by the NATS server when conn is closed and retry attempts are exhausted.
	connClosedHandler ConnClosedHandler
	logger            *logger.Logger
}

func NewJetStream(config env.NatsConfig, closeHandler ConnClosedHandler, logger *logger.Logger) *JetStream {
	return &JetStream{
		config:            config,
		connClosedHandler: closeHandler,
		logger:            logger,
	}
}

func (js *JetStream) Initialize(_ env.Config) error {
	if err := js.validateConfig(); err != nil {
		return err
	}
	if err := js.initNATSConn(); err != nil {
		return err
	}
	if err := js.initJSContext(); err != nil {
		return err
	}
	return js.ensureStreamExists()
}

func (js *JetStream) SyncSubscription(_ *eventingv1alpha1.Subscription, _ ...interface{}) (bool, error) {
	panic("implement me")
}

func (js *JetStream) DeleteSubscription(_ *eventingv1alpha1.Subscription) error {
	panic("implement me")
}

func (js *JetStream) validateConfig() error {
	if js.config.JSStreamName == "" {
		return errors.New("Stream name cannot be empty")
	}
	if _, err := toJetStreamStorageType(js.config.JSStreamStorageType); err != nil {
		return err
	}
	return nil
}

func (js *JetStream) initNATSConn() error {
	if js.conn == nil || js.conn.Status() != nats.CONNECTED {
		jsOptions := []nats.Option{
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(js.config.MaxReconnects),
			nats.ReconnectWait(js.config.ReconnectWait),
		}
		conn, err := nats.Connect(js.config.URL, jsOptions...)
		if err != nil || !conn.IsConnected() {
			return errors.Wrapf(err, "failed to connect to NATS JetStream")
		}
		js.conn = conn
		if js.connClosedHandler != nil {
			js.conn.SetClosedHandler(nats.ConnHandler(js.connClosedHandler))
		}
	}
	return nil
}

func (js *JetStream) initJSContext() error {
	jsCtx, err := js.conn.JetStream()
	if err != nil {
		return errors.Wrapf(err, "failed to create the JetStream context")
	}
	js.jsCtx = jsCtx
	return nil
}

func (js *JetStream) ensureStreamExists() error {
	if info, err := js.jsCtx.StreamInfo(js.config.JSStreamName); err == nil {
		// TODO: in case the stream exists, should we make sure all of its configs matches the stream config in the chart?
		js.namedLogger().Infow("reusing existing Stream", "streamName", info.Config.Name)
		return nil
	} else if err != nats.ErrStreamNotFound {
		return err
	}
	storage, err := toJetStreamStorageType(js.config.JSStreamStorageType)
	if err != nil {
		return err
	}
	js.namedLogger().Infow("Stream not found, creating a new Stream",
		"streamName", js.config.JSStreamName, "streamStorageType", storage.String())
	_, err = js.jsCtx.AddStream(&nats.StreamConfig{
		Name:    js.config.JSStreamName,
		Storage: storage,
		// Since one stream is used to store events of all types, the stream has to match all event types, and therefore
		// we use the wildcard char >. However, to avoid matching internal JetStream and non-Kyma-related subjects, we
		// use the stream name as a prefix. This prefix is handled only on the JetStream level (i.e. JetStream handler
		// and EPP) and should not be exposed in the Kyma subscription. Any Kyma event type gets appended with the
		// configured stream name.
		Subjects: []string{fmt.Sprintf("%s.>", js.config.JSStreamName)},
	})
	return err
}

func (js *JetStream) namedLogger() *zap.SugaredLogger {
	return js.logger.WithContext().Named(jsHandlerName)
}
