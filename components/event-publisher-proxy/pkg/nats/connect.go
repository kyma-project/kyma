package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// TODO: document me
type connectionData struct {
	url                  string
	retryOnFailedConnect bool
	maxReconnects        int
	reconnectWait        time.Duration
}

type BackendConnection struct {
	connectionData connectionData
	Connection     *nats.Conn
}

type BackendConnectionOpt func(*BackendConnection)

func WithBackendConnectionURL(url string) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.url = url
	}
}

func WithBackendConnectionRetryOnFailedConnect(retry bool) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.retryOnFailedConnect = retry
	}
}

func WithBackendConnectionMaxReconnects(reconnects int) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.maxReconnects = reconnects
	}
}

func WithBackendConnectionReconnectWait(wait time.Duration) BackendConnectionOpt {
	return func(bc *BackendConnection) {
		bc.connectionData.reconnectWait = wait
	}
}

// NewBackendConnection returns a new new Nats connection instance with the given BackendConnectionOpt
func NewBackendConnection(opts ...BackendConnectionOpt) *BackendConnection {

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

	connection, err := nats.Connect(bc.connectionData.url, nats.RetryOnFailedConnect(bc.connectionData.retryOnFailedConnect),
		nats.MaxReconnects(bc.connectionData.maxReconnects), nats.ReconnectWait(bc.connectionData.reconnectWait))
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
