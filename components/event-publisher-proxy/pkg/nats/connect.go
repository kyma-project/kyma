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

func WithBackendConnectionURL(url string) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.url = url
	}
}

func WithBackendConnectionRetry(retry bool) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.retry = retry
	}
}

func WithBackendConnectionReconnects(reconnects int) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.reconnects = reconnects
	}
}

func WithBackendConnectionWait(wait time.Duration) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.wait = wait
	}
}

// NewBackendConnection returns a new new Nats connection instance with the given BackendConnectionOpt
func NewBackendConnection(opts ...*BackendConnectionOpt) *BackendConnection {
	connData := connectionData{}
	bc := &BackendConnection{
		connectionData: connData,
	}

	// apply options
	for _, opt := range opts {
		opt(bc)
	}

	return bc
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
