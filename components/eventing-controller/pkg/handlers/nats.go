package handlers

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

// compile time check
var _ NatsInterface = &Nats{}

type NatsInterface interface {
	Initialize(cfg env.NatsConfig) error
}

type Nats struct {
	Connection *nats.Conn
	Log        logr.Logger
}

// Initialize creates a connection to Nats
func (n *Nats) Initialize(cfg env.NatsConfig) error {
	n.Log.Info("Initialize NATS connection")
	var err error
	if n.Connection == nil || n.Connection.Status() != nats.CONNECTED {
		n.Connection, err = nats.Connect(cfg.Url,
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(cfg.MaxReconnects),
			nats.ReconnectWait(cfg.ReconnectWait))
		if err != nil {
			return errors.Wrapf(err, "failed to connect to Nats")
		}
		if n.Connection.Status() != nats.CONNECTED {
			notConnectedErr := fmt.Errorf("not connected: status: %v", n.Connection.Status())
			return notConnectedErr
		}
	}
	n.Log.Info(fmt.Sprintf("Successfully connected to Nats: %v", n.Connection.Status()))
	return nil
}
