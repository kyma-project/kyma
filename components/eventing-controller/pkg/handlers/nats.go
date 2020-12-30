package handlers

import (
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/nats-io/nats.go"
)

// compile time check
var _ NatsInterface = &Nats{}

type NatsInterface interface {
	Initialize(cfg env.NatsConfig)
}

type Nats struct {
	Connection *nats.Conn
	Log        logr.Logger
}

type NatsResponse struct {
	StatusCode int
	Error      error
}

func (n *Nats) Initialize(cfg env.NatsConfig) {
	n.Log.Info("Initialize NATS connection")
	if n.Connection == nil {
		var err error
		n.Connection, err = nats.Connect(cfg.Url)
		if err != nil {
			n.Log.Error(err, "Can't connect to NATS Server")
		}
	}
}
