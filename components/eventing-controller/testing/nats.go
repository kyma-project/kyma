package testing

import (
	"github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
)

// RunNatsServerOnPort will run a server on the given port.
func RunNatsServerOnPort(port int) *server.Server {
	opts := natstestserver.DefaultTestOptions
	opts.Port = port
	return natstestserver.RunServer(&opts)
}
