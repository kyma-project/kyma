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

// NewBackendConnection returns a new NewNatsConnection instance with the given NATS connection data.
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
func (bc *BackendConnection) Connect() (err error) {
	bc.Connection, err = nats.Connect(bc.connectionData.url,
		nats.RetryOnFailedConnect(bc.connectionData.retry),
		nats.MaxReconnects(bc.connectionData.reconnects),
		nats.ReconnectWait(bc.connectionData.wait),
	)
	if err != nil {
		return err
	}

	if status := bc.Connection.Status(); status != nats.CONNECTED {
		return fmt.Errorf("cannot connect to NATS server, status: %v", status)
	}

	return nil
}
