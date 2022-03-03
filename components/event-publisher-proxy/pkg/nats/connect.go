package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Connect returns a NATS connection that is ready for use, or an error if connection to the NATS server failed.
// It uses the nats.Connect function which is thread-safe.
func Connect(url string, retry bool, reconnects int, wait time.Duration) (*nats.Conn, error) {
	connection, err := nats.Connect(url,
		nats.RetryOnFailedConnect(retry),
		nats.MaxReconnects(reconnects),
		nats.ReconnectWait(wait),
	)
	if err != nil {
		return nil, err
	}

	if status := connection.Status(); status != nats.CONNECTED {
		return nil, fmt.Errorf("NATS connection not connected with status:%v", status)
	}

	return connection, err
}
