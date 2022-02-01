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

	bc := pkgnats.NewBackendConnection2(
		pkgnats.WithBackendConnectionURL(test.natsURL),
		pkgnats.WithBackendConnectionRetries(true),
		pkgnats.WithBackendConnectionReconnects(1),
		pkgnats.WithBackendConnectionWait(3),
	)
	err := bc.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, bc.Connection)
	assert.Equal(t, bc.Connection.Status(), nats.CONNECTED)

	bc.Connection.Close()

	err = bc.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, bc.Connection)
	assert.Equal(t, bc.Connection.Status(), nats.CONNECTED)

	defer bc.Connection.Close()
}
