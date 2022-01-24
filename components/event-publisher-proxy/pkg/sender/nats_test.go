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
	// TODO:
	backendConnection *pkgnats.BackendConnection
	// a sender for publishing events to the NATS server
	natsMessageSender *NatsMessageSender
	// a logger
	logger *logrus.Logger
}

// TestSendCloudEvent ensures that
func TestSendCloudEvent(t *testing.T) {
	testEnv := setupTestEnvironment(t)

	// subscribe to subject
	done := make(chan bool, 1)
	validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), done)
	testingutils.SubscribeToEventOrFail(t, testEnv.backendConnection.Connection, testingutils.CloudEventType, validator)

	// create cloudevent
	ce := cloudevents.NewEvent()
	ce.SetType(testingutils.CloudEventType)
	err := json.Unmarshal([]byte(testingutils.StructuredCloudEventPayloadWithCleanEventType), &ce)
	assert.Nil(t, err)

	sendEventAndAssertStatus(testEnv.context, t, testEnv.natsMessageSender, &ce, http.StatusNoContent)

	// wait for subscriber to receive the messages
	if err := testingutils.WaitForChannelOrTimeout(done, time.Second*3); err != nil {
		t.Fatalf("Subscriber did not receive the message with error: %v", err)
	}
}

func TestSendCloudEventWithReconnect(t *testing.T) {
	testEnv := setupTestEnvironment(t)

	// subscribe to subject
	done := make(chan bool, 1)
	validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), done)
	testingutils.SubscribeToEventOrFail(t, testEnv.backendConnection.Connection, testingutils.CloudEventType, validator)

	// create cloudevent
	ce := cloudevents.NewEvent()
	ce.SetType(testingutils.CloudEventType)
	err := json.Unmarshal([]byte(testingutils.StructuredCloudEventPayloadWithCleanEventType), &ce)
	assert.Nil(t, err)

	sendEventAndAssertStatus(testEnv.context, t, testEnv.natsMessageSender, &ce, http.StatusNoContent)

	// wait for subscriber to receive the messages
	if err := testingutils.WaitForChannelOrTimeout(done, time.Second*3); err != nil {
		t.Fatalf("Subscriber did not receive the message with error: %v", err)
	}

	// close connection
	testEnv.backendConnection.Connection.Close()
	assert.True(t, testEnv.backendConnection.Connection.IsClosed())
	sendEventAndAssertStatus(testEnv.context, t, testEnv.natsMessageSender, &ce, http.StatusNoContent)
}

func sendEventAndAssertStatus(ctx context.Context, t *testing.T, sender *NatsMessageSender, event *event.Event, expectedStatus int) {
	status, err := sender.Send(ctx, event)
	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, status)
}

// setupTestEnvironment creates the test environment for the integration tests in this file
// It performs the follow steps:
// - create logger
// - create and start NATS server
// - establish a connection to the NATS server `TestEnvironment.backendConnection`
// - create a sender to publish messsages to NATS
func setupTestEnvironment(t *testing.T) TestEnvironment {
	// Create logger
	logger := logrus.New()

	// Start Nats server
	natsServer := testingutils.StartNatsServer()
	assert.NotNil(t, natsServer)
	defer natsServer.Shutdown()

	// connect to nats
	bc := pkgnats.NewBackendConnection(natsServer.ClientURL(), true, 1, time.Second)
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
