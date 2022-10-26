package testing

import (
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

const (
	StreamName = "kyma"
)

func StartNATSServer() *server.Server {
	opts := test.DefaultTestOptions
	opts.Port = server.RANDOM_PORT
	opts.JetStream = true
	opts.Host = "localhost"

	log, _ := logger.New("json", "info")
	log.WithContext().Info("Starting test NATS Server in JetStream mode")
	return test.RunServer(&opts)
}

func ConnectToNATSServer(url string) (*nats.Conn, error) {
	return pkgnats.Connect(url,
		pkgnats.WithRetryOnFailedConnect(true),
		pkgnats.WithMaxReconnects(3),
		pkgnats.WithReconnectWait(time.Second),
	)
}
