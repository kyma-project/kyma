package nats

import (
	"testing"

	publishertesting "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestConnectToNats(t *testing.T) {
	natsServer := publishertesting.StartNatsServer()
	assert.NotNil(t, natsServer)
	assert.NotEmpty(t, natsServer.ClientURL())
	defer natsServer.Shutdown()

	c := NewConnection(
		natsServer.ClientURL(),
		WithRetryOnFailedConnect(true),
		WithMaxReconnects(1),
		WithReconnectWait(3),
	)
	err := c.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, c.Connection)
	assert.Equal(t, c.Connection.Status(), nats.CONNECTED)

	c.Connection.Close()

	err = c.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, c.Connection)
	assert.Equal(t, c.Connection.Status(), nats.CONNECTED)

	defer c.Connection.Close()
}
