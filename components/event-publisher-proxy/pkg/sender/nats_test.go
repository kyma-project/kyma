// Tests in this file are integration tests.
// They do use a real NATS server using github.com/nats-io/nats-server/v2/server.
// Messages are sent using NatsMessageSender interface.
package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

// TestEnvironment contains the necessary entities to perform NATS integration tests
type TestEnvironment struct {
	context context.Context
	logger  *logrus.Logger

	// natsServer is a real NATS server for integration testing.
	natsServer *server.Server
	// natsSendConnection is a connection to the NATS server which shall be used to send events.
	natsSendConnection *pkgnats.Connection
	// natsRecvConnection is a connection to the NATS server which shall be used to receive events.
	natsRecvConnection *pkgnats.Connection
	// natsMessageSender is a sender for publishing events to the NATS server.
	natsMessageSender *NatsMessageSender
}

// setupTestEnvironment creates the test environment for the integration tests in this file.
// It performs the follow steps:
// - create logger
// - create and start NATS server
// - establish two connections to the NATS server: one for sending and one for receiving events
// - create a sender to publish messsages to NATS
func setupTestEnvironment(t *testing.T, connectionOpts ...pkgnats.Opt) TestEnvironment {

	// Create logger
	logger := logrus.New()

	// Start Nats server
	natsServer := testingutils.StartNatsServer()
	assert.NotNil(t, natsServer)
	t.Cleanup(func() {
		natsServer.Shutdown()
	})

	allConnectionOpts := []pkgnats.Opt{
		pkgnats.WithMaxReconnects(1),
		pkgnats.WithRetryOnFailedConnect(true),
		pkgnats.WithReconnectWait(time.Second),
	}
	allConnectionOpts = append(allConnectionOpts, connectionOpts...)

	// setup natsSendConnection
	natsSendConnection := pkgnats.NewConnection(
		natsServer.ClientURL(),
		allConnectionOpts...,
	)
	err := natsSendConnection.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, natsSendConnection.Connection)

	// setup natsSendConnection
	natsRecvConnection := pkgnats.NewConnection(
		natsServer.ClientURL(),
		allConnectionOpts...,
	)
	err = natsRecvConnection.Connect()
	assert.NotNil(t, natsRecvConnection)
	assert.Nil(t, err)

	// create message natsMessageSender
	ctx := context.Background()
	natsMessageSender := NewNatsMessageSender(ctx, natsSendConnection, logger)
	logger.Info("TestNatsSender started")

	return TestEnvironment{
		context:            ctx,
		natsServer:         natsServer,
		natsSendConnection: natsSendConnection,
		natsRecvConnection: natsRecvConnection,
		natsMessageSender:  natsMessageSender,
		logger:             logger,
	}
}

// TestSendCloudEventsToNats tests that sending cloud events to a NATS server works.
func TestSendCloudEventsToNats(t *testing.T) {
	testCases := []struct {
		name                              string
		givenRetries                      int
		wantHTTPStatusCode                int
		wantClosedConnectionBeforeSending bool
	}{
		{
			name:                              "sending event to NATS works",
			givenRetries:                      1,
			wantHTTPStatusCode:                http.StatusNoContent,
			wantClosedConnectionBeforeSending: false,
		},
		{
			name:               "sending event to NATS works given a closed connection",
			givenRetries:       10,
			wantHTTPStatusCode: http.StatusNoContent,
			// Close connection before sending so we can check the reconnect behaviour of the NATS connection.
			wantClosedConnectionBeforeSending: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testEnv := setupTestEnvironment(t, pkgnats.WithMaxReconnects(tc.givenRetries))

			// subscribe to subject
			subject := fmt.Sprintf(`"%s"`, testingutils.CloudEventData)
			// NOTE: we are using the testEnv.natsRecvConnection instead of testEnv.natsSendConnection because the latter will get reconnected based on wantClosedConnectionBeforeSending, this will fail the test when trying to receive a message.
			done := subscribeToSubject(t, testEnv.natsRecvConnection, subject)

			// create cloudevent with default data (testing.CloudEventData)
			ce := cloudevents.NewEvent()
			ce.SetType(testingutils.CloudEventType)
			err := json.Unmarshal([]byte(testingutils.StructuredCloudEventPayloadWithCleanEventType), &ce)
			assert.Nil(t, err)

			if tc.wantClosedConnectionBeforeSending {
				// close connection
				testEnv.natsSendConnection.Connection.Close()
				// ensure connection is closed
				// this is important because we want to test that the connection gets re-established as soon as we send an event
				assert.True(t, testEnv.natsSendConnection.Connection.IsClosed())
			}
			sendEventAndAssertHTTPStatusCode(testEnv.context, t, testEnv.natsMessageSender, &ce, tc.wantHTTPStatusCode)

			// wait for subscriber to receive the messages
			err = testingutils.WaitForChannelOrTimeout(done, time.Second*3)
			assert.NoError(t, err, "Subscriber did not receive the message")
		})
	}
}

// sendEventAndAssertHTTPStatusCode sends the event to NATS and asserts that the expectedStatus is returned from NATS
func sendEventAndAssertHTTPStatusCode(ctx context.Context, t *testing.T, sender *NatsMessageSender, event *event.Event, expectedStatus int) {
	status, err := sender.Send(ctx, event)
	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, status)
}

// subscribeToSubject subscribes to the NATS subject using the connection.
// It then ensures that the data is received on the subject and validated using the natsMessageDataValidator.
func subscribeToSubject(t *testing.T, connection *pkgnats.Connection, subject string) chan bool {
	done := make(chan bool, 1)
	natsMessageDataValidator := testingutils.ValidateNatsMessageDataOrFail(t, subject, done)
	testingutils.SubscribeToEventOrFail(t, connection.Connection, testingutils.CloudEventType, natsMessageDataValidator)
	return done
}
