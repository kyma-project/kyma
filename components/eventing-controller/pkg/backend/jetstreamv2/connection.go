package jetstreamv2

import (
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"
	"github.com/nats-io/nats.go"
)

type Builder interface {
	Build() (ConnectionInterface, error)
}

// ConnectionBuilder helps in establishing a connection to NATS.
type ConnectionBuilder struct {
	config backendnats.Config
}

func NewConnectionBuilder(config backendnats.Config) Builder {
	return ConnectionBuilder{config: config}
}

// Build connects to NATS and returns the connection. If an error occurs, ErrConnect is returned.
func (b ConnectionBuilder) Build() (ConnectionInterface, error) {
	config := b.config
	jsOptions := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.Name("Kyma Controller"),
	}
	conn, err := nats.Connect(config.URL, jsOptions...)
	if err != nil || !conn.IsConnected() {
		return nil, pkgerrors.MakeError(ErrConnect, err)
	}

	return conn, nil
}

// ConnectionInterface is a contract for a NATS connection object.
type ConnectionInterface interface {
	IsConnected() bool
	SetClosedHandler(cb nats.ConnHandler)
	SetReconnectHandler(rcb nats.ConnHandler)
	JetStream(opts ...nats.JSOpt) (nats.JetStreamContext, error)
}
