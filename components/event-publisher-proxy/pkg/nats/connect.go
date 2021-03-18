package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// ConnectToNats returns a nats connection that is ready for use, or error if connection to the nats server failed
func ConnectToNats(url string, retry bool, reconnects int, wait time.Duration) (*nats.Conn, error) {
	connection, err := nats.Connect(url, nats.RetryOnFailedConnect(retry), nats.MaxReconnects(reconnects), nats.ReconnectWait(wait))
	if err != nil {
		return nil, err
	}
	if status := connection.Status(); status != nats.CONNECTED {
		return nil, fmt.Errorf("connection status not connected: %v", status)
	}
	return connection, nil
}
