package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type connectionData struct {
	// url is URL of the NATS server.
	url string
	// retryOnFailedConnect reconnects if the connection cannot be established.
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
	return func(c *Connection) {
		c.connectionData.retryOnFailedConnect = retryOnFailedConnect
	}
}

func WithMaxReconnects(maxReconnects int) Opt {
	return func(c *Connection) {
		c.connectionData.maxReconnects = maxReconnects
	}
}

func WithReconnectWait(reconnectWait time.Duration) Opt {
	return func(c *Connection) {
		c.connectionData.reconnectWait = reconnectWait
	}
}

// NewConnection returns a new Nats connection instance with the given BackendConnectionOpt
func NewConnection(url string, opts ...Opt) *Connection {

	connData := connectionData{
		url: url,
	}
	c := &Connection{
		connectionData: connData,
	}

	// apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect returns a NATS connection that is ready for use, or an error if connection to the NATS server failed.
// It uses the nats.Connect function which is thread-safe.
func (c *Connection) Connect() error {

	connection, err := nats.Connect(c.connectionData.url, nats.RetryOnFailedConnect(c.connectionData.retryOnFailedConnect),
		nats.MaxReconnects(c.connectionData.maxReconnects), nats.ReconnectWait(c.connectionData.reconnectWait))
	if err != nil {
		return err
	}
	if status := connection.Status(); status != nats.CONNECTED {
		return fmt.Errorf("connection status not connected: %v", status)
	}
	// OK
	c.Connection = connection
	return nil
}
