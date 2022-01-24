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

// Tests in this file are integration tests.
// They do use a real NATS server using `github.com/nats-io/nats-server/v2/server`.
// Messages are sent using `NatsMessageSender` interface.

// TestEnvironment contains the necessary entities to perform NATS integration tests
type TestEnvironment struct {
	context context.Context
	// a NATS server
	natsServer *server.Server
	// a connection to the NATS server
	backendConnection *pkgnats.BackendConnection
	// a sender for publishing events to the NATS server
	natsMessageSender *NatsMessageSender
	// a logger
	logger *logrus.Logger
}

// setupTestEnvironment creates the test environment for the integration tests in this file
// It performs the follow steps:
// - create logger
// - create and start NATS server
// - establish a connection to the NATS server `TestEnvironment.backendConnection`
// - create a sender to publish messsages to NATS
// TODO: what if I want configure something else than the nats server ?
// NOTE: if you need any of these objects to be customized, add it to the method signature and overwrite the default instance
func setupTestEnvironment(t *testing.T, connectionOpts ...pkgnats.BackendConnectionOption) TestEnvironment {

	// Create logger
	logger := logrus.New()

	// Start Nats server
	natsServer := testingutils.StartNatsServer()
	assert.NotNil(t, natsServer)

	// TODO: move backend connection outside ?
	// connect to nats
	bc := pkgnats.NewBackendConnection(natsServer.ClientURL(), true, 1, time.Second)
	for _, opt := range connectionOpts {
		opt(bc)
	}
	err := bc.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, bc.Connection)

	// create message sender
	ctx := context.Background()
	sender := NewNatsMessageSender(ctx, bc, logger)
	logger.Info("TestNatsSender started")

	return TestEnvironment{
		context:           ctx,
		natsServer:        natsServer,
		backendConnection: bc,
		natsMessageSender: sender,
		logger:            logger,
	}
}

func TestSendCloudEventsToNats(t *testing.T) {
	testCases := []struct {
		name         string
		givenRetries int
	}{
		{name: "1 retry", givenRetries: 1},
		{name: "10 retries", givenRetries: 10},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testEnv := setupTestEnvironment(t, pkgnats.WithBackendConnectionRetries(tc.givenRetries))
			defer testEnv.natsServer.Shutdown()

			// subscribe to subject
			subject := fmt.Sprintf(`"%s"`, testingutils.CloudEventData)
			done := subscribeToSubject(t, testEnv.backendConnection, subject)

			// create cloudevent with default data (testing.CloudEventData)
			ce := cloudevents.NewEvent()
			ce.SetType(testingutils.CloudEventType)
			err := json.Unmarshal([]byte(testingutils.StructuredCloudEventPayloadWithCleanEventType), &ce)
			assert.Nil(t, err)

			sendEventAndAssertStatus(testEnv.context, t, testEnv.natsMessageSender, &ce, http.StatusNoContent)

			// TODO: rewrite using assertion
			// wait for subscriber to receive the messages
			if err := testingutils.WaitForChannelOrTimeout(done, time.Second*3); err != nil {
				t.Fatalf("Subscriber did not receive the message with error: %v", err)
			}

			// close connection
			testEnv.backendConnection.Connection.Close()
			assert.True(t, testEnv.backendConnection.Connection.IsClosed())
			sendEventAndAssertStatus(testEnv.context, t, testEnv.natsMessageSender, &ce, http.StatusNoContent)
		})
	}
}

func sendEventAndAssertStatus(ctx context.Context, t *testing.T, sender *NatsMessageSender, event *event.Event, expectedStatus int) {
	status, err := sender.Send(ctx, event)
	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, status)
}

// subscribeToSubject subscribes to the given subject using the given connection.
// It then ensures that the data received on the subject
func subscribeToSubject(t *testing.T, connection *pkgnats.BackendConnection, subject string) chan bool {
	done := make(chan bool, 1)
	natsMessageDataValidator := testingutils.ValidateNatsMessageDataOrFail(t, subject, done)
	// TODO: how does validation of message data work ?
	testingutils.SubscribeToEventOrFail(t, connection.Connection, testingutils.CloudEventType, natsMessageDataValidator)
	return done
}
