package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type connectionData struct {
	// url is URL of the NATS server.
	url string
	// retryOnFailedConnect reconnects if the conneciton cannot be established.
	retryOnFailedConnect bool
	// maxReconnects is used when retryOnFailedConnect is true. It sets the number of reconnect attempts. Negative means try indefinitely.
	maxReconnects int
	// reconnectWait is used when retryOnFailedConnect is true. It sets the time to wait until a reconnect is established.
	reconnectWait time.Duration
}

type Connection struct {
	connectionData connectionData
	Connection     *nats.Conn
}

type Opt func(*Connection)

func WithRetryOnFailedConnect(retryOnFailedConnect bool) Opt {
	return func(bc *Connection) {
		bc.connectionData.retryOnFailedConnect = retryOnFailedConnect
	}
}

func WithMaxReconnects(maxReconnects int) Opt {
	return func(bc *Connection) {
		bc.connectionData.maxReconnects = maxReconnects
	}
}

func WithReconnectWait(reconnectWait time.Duration) Opt {
	return func(bc *Connection) {
		bc.connectionData.reconnectWait = reconnectWait
	}
}

// NewConnection returns a new Nats connection instance with the given BackendConnectionOpt
func NewConnection(url string, opts ...Opt) *Connection {

	connData := connectionData{
		url: url,
	}
	bc := &Connection{
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
func (bc *Connection) Connect() error {

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
