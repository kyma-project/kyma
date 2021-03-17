package nats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"

	publishertesting "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestConnectToNats(t *testing.T) {
	natsServer := publishertesting.StartNatsServer()
	assert.NotNil(t, natsServer)
	assert.NotEmpty(t, natsServer.ClientURL())
	defer natsServer.Shutdown()

	connection, err := ConnectToNats(natsServer.ClientURL(), true, 1, 3)
	assert.Nil(t, err)
	assert.NotNil(t, connection)
	assert.Equal(t, connection.Status(), nats.CONNECTED)
	defer connection.Close()
}
