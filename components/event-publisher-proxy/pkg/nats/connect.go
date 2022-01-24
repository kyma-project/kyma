package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type connectionData struct {
	url        string
	retry      bool
	reconnects int
	wait       time.Duration
}

type BackendConnection struct {
	connectionData connectionData
	Connection     *nats.Conn
}

type BackendConnectionOpt func(*BackendConnection)

func WithBackendConnectionRetries(n int) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.reconnects = n
	}
}

// NewBackendConnection returns a new NewNatsConnection instance with the given NATS connection data.
// TODO: make use of option pattern here
func NewBackendConnection(url string, retry bool, reconnects int, wait time.Duration) *BackendConnection {
	return &BackendConnection{
		connectionData: connectionData{
			url:        url,
			retry:      retry,
			reconnects: reconnects,
			wait:       wait,
		},
	}
}

// Connect returns a NATS connection that is ready for use, or an error if connection to the NATS server failed.
// It uses the nats.Connect function which is thread-safe.
func (bc *BackendConnection) Connect() error {
	connection, err := nats.Connect(bc.connectionData.url, nats.RetryOnFailedConnect(bc.connectionData.retry),
		nats.MaxReconnects(bc.connectionData.reconnects), nats.ReconnectWait(bc.connectionData.wait))
	if err != nil {
		return err
	}
	if status := connection.Status(); status != nats.CONNECTED {
		return fmt.Errorf("connection status not connected: %v", status)
	}
	// OK
	bc.Connection = connection
	return nil
}
