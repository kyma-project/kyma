package nats_test

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"

	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	publishertesting "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestConnect(t *testing.T) {
	testCases := []struct {
		name                     string
		wantRetryOnFailedConnect bool
		wantMaxReconnect         int
		wantReconnectWait        time.Duration
	}{
		{
			name:                     "do not retry failed connections",
			wantRetryOnFailedConnect: false,
			wantMaxReconnect:         0,
			wantReconnectWait:        time.Millisecond,
		},
		{
			name:                     "keep retrying failed connections",
			wantRetryOnFailedConnect: true,
			wantMaxReconnect:         -1,
			wantReconnectWait:        time.Millisecond,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			natsServer := publishertesting.StartNatsServer()
			assert.NotNil(t, natsServer)
			defer natsServer.Shutdown()

			clientURL := natsServer.ClientURL()
			assert.NotEmpty(t, clientURL)

			connection, err := pkgnats.Connect(clientURL, tc.wantRetryOnFailedConnect, tc.wantMaxReconnect, tc.wantReconnectWait)
			assert.Nil(t, err)
			assert.NotNil(t, connection)

			assert.Equal(t, connection.Status(), nats.CONNECTED)
			assert.Equal(t, clientURL, connection.Opts.Servers[0])
			assert.Equal(t, tc.wantRetryOnFailedConnect, connection.Opts.RetryOnFailedConnect)
			assert.Equal(t, tc.wantMaxReconnect, connection.Opts.MaxReconnect)
			assert.Equal(t, tc.wantReconnectWait, connection.Opts.ReconnectWait)
		})
	}
}
