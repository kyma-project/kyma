//go:build integration
// +build integration

package jetstreamv2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/testing"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_ConnectionBuilder_Build tests that the connection object is created and correctly configured.
func Test_ConnectionBuilder_Build(t *testing.T) {
	// SuT: ConnectionBuilder
	// UoW: Build()
	// test kind: state verification

	// given: a NATS server.
	natsServer := startManagedNATSServer(t)

	config := backendnats.Config{URL: natsServer.ClientURL()}
	cb := jetstreamv2.NewConnectionBuilder(config)

	// when: connection establishment should work.
	connection, err := cb.Build()

	// then: connection is configured correctly.
	require.NoError(t, err)
	require.NotNil(t, connection)

	// get the actual connection for state testing
	natsConn, ok := connection.(*nats.Conn)
	require.True(t, ok)

	// ensure the options are set
	assert.Equal(t, natsConn.Opts.Name, "Kyma Controller")
	assert.Equal(t, natsConn.Opts.MaxReconnect, config.MaxReconnects)
	assert.Equal(t, natsConn.Opts.ReconnectWait, config.ReconnectWait)
	assert.Equal(t, natsConn.Opts.RetryOnFailedConnect, true)
}

func Test_ConnectionBuilder_Build_ForErrConnect(t *testing.T) {
	// SuT: ConnectionBuilder
	// UoW: Build()
	// test kind: value
	t.Parallel()

	// given: a free port on localhost without a NATS server running.
	url := fixtureUnusedLocalhostURL(t)

	config := backendnats.Config{URL: url}
	cb := jetstreamv2.NewConnectionBuilder(config)

	// when: connection establishment should fail as
	// there is no NATS server to connect to on this URL.
	_, err := cb.Build()

	// then
	assert.ErrorIs(t, err, jetstreamv2.ErrConnect)
}

// Test_ConnectionBuilder_IsConnected ensures that the IsConnected method always returns the correct value.
func Test_ConnectionBuilder_IsConnected(t *testing.T) {
	// SuT: ConnectionBuilder
	// UoW: Build()
	// test kind: state verification

	// given
	natsServer := startManagedNATSServer(t)

	config := backendnats.Config{URL: natsServer.ClientURL()}
	cb := jetstreamv2.NewConnectionBuilder(config)

	// when: a NATS server is online
	connection, err := cb.Build()

	// then
	assert.True(t, connection.IsConnected())
	assert.NoError(t, err)
	assert.NotNil(t, connection)

	// when: NATS server is offline
	natsServer.Shutdown()
	// then
	require.Eventually(t, func() bool {
		return !connection.IsConnected()
	}, 60*time.Second, 10*time.Millisecond)
}

// startManagedNATSServer starts a NATS server and shuts the server down as soon as the test is
// completed (also when it failed!).
func startManagedNATSServer(t *testing.T) *server.Server {
	natsServer, _, err := natstesting.StartNATSServer(evtesting.WithJetStreamEnabled())
	require.NoError(t, err)
	t.Cleanup(func() {
		natsServer.Shutdown()
	})
	return natsServer
}

// fixtureUnusedLocalhostUrl provides a localhost URL with an unused port.
func fixtureUnusedLocalhostURL(t *testing.T) string {
	port, err := testingv2.GetFreePort()
	require.NoError(t, err)
	url := fmt.Sprintf("http://localhost:%d", port)
	return url
}
