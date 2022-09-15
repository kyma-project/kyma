package testing

import (
	"github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
)

type NatsServerOpt func(opts *server.Options)

func WithJetStreamEnabled() NatsServerOpt {
	return func(opts *server.Options) {
		opts.JetStream = true
	}
}

func WithPort(port int) NatsServerOpt {
	return func(opts *server.Options) {
		opts.Port = port
	}
}

// RunNatsServerOnPort will run a server with the given server options.
func RunNatsServerOnPort(opts ...NatsServerOpt) *server.Server {
	serverOpts := &natstestserver.DefaultTestOptions
	for _, opt := range opts {
		opt(serverOpts)
	}
	return natstestserver.RunServer(serverOpts)
}

// ShutDownNATSServer shuts down test NATS server and waits until shutdown is complete.
func ShutDownNATSServer(natsServer *server.Server) {
	natsServer.Shutdown()
	natsServer.WaitForShutdown()
}
