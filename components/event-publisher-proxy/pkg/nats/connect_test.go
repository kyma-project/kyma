//go:build unit
// +build unit

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
		name                      string
		givenRetryOnFailedConnect bool
		givenMaxReconnect         int
		givenReconnectWait        time.Duration
	}{
		{
			name:                      "do not retry failed connections",
			givenRetryOnFailedConnect: false,
			givenMaxReconnect:         0,
			givenReconnectWait:        time.Millisecond,
		},
		{
			name:                      "keep retrying failed connections",
			givenRetryOnFailedConnect: true,
			givenMaxReconnect:         -1,
			givenReconnectWait:        time.Millisecond,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			// test in both default and jetstream NATS modes
			for _, serverMode := range publishertesting.NATSServerModes {
				t.Run(serverMode.Name, func(t *testing.T) {
					tc := tc
					// given
					natsServer := publishertesting.StartNATSServer(serverMode.JetStreamEnabled)
					assert.NotNil(t, natsServer)
					defer natsServer.Shutdown()

					clientURL := natsServer.ClientURL()
					assert.NotEmpty(t, clientURL)

					// when
					connection, err := pkgnats.Connect(clientURL,
						pkgnats.WithRetryOnFailedConnect(tc.givenRetryOnFailedConnect),
						pkgnats.WithMaxReconnects(tc.givenMaxReconnect),
						pkgnats.WithReconnectWait(tc.givenReconnectWait),
					)
					assert.Nil(t, err)
					assert.NotNil(t, connection)
					defer func() { connection.Close() }()

					// then
					assert.Equal(t, connection.Status(), nats.CONNECTED)
					assert.Equal(t, clientURL, connection.Opts.Servers[0])
					assert.Equal(t, tc.givenRetryOnFailedConnect, connection.Opts.RetryOnFailedConnect)
					assert.Equal(t, tc.givenMaxReconnect, connection.Opts.MaxReconnect)
					assert.Equal(t, tc.givenReconnectWait, connection.Opts.ReconnectWait)
				})
			}
		})
	}
}
