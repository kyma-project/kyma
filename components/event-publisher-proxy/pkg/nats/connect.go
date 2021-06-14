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
	bc := &BackendConnection{}
	bc.connectionData = connectionData{url: url, retry: retry, reconnects: reconnects, wait: wait}
	return bc
}

// Connect returns a nats connection that is ready for use, or error if connection to the nats server failed
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

func (bc *BackendConnection) Reconnect() error {
	return bc.Connect()
}
