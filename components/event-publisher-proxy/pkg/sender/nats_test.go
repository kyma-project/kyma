package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestSendCloudEvent(t *testing.T) {
	logger := logrus.New()
	logger.Info("TestNatsSender started")

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

	// subscribe to subject
	done := make(chan bool, 1)
	validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), done)
	testingutils.SubscribeToEventOrFail(t, bc.Connection, testingutils.CloudEventType, validator)

	// create cloudevent
	ce := testingutils.StructuredCloudEventPayloadWithCleanEventType
	event := cloudevents.NewEvent()
	event.SetType(testingutils.CloudEventType)
	err = json.Unmarshal([]byte(ce), &event)
	assert.Nil(t, err)

	// send cloudevent
	status, err := sender.Send(ctx, &event)
	assert.Nil(t, err)
	assert.Equal(t, status, http.StatusNoContent)

	// wait for subscriber to receive the messages
	if err := testingutils.WaitForChannelOrTimeout(done, time.Second*3); err != nil {
		t.Fatalf("Subscriber did not receive the message with error: %v", err)
	}
}

func TestSendCloudEventWithReconnect(t *testing.T) {
	logger := logrus.New()
	logger.Info("TestNatsSender started")

	// Start Nats server
	natsServer := testingutils.StartNatsServer()
	assert.NotNil(t, natsServer)
	defer natsServer.Shutdown()

	// connect to nats
	bc := pkgnats.NewBackendConnection(natsServer.ClientURL(), true, 10, time.Second)
	err := bc.Connect()
	assert.Nil(t, err)
	assert.NotNil(t, bc.Connection)

	// create message sender
	ctx := context.Background()
	sender := NewNatsMessageSender(ctx, bc, logger)

	// subscribe to subject
	done := make(chan bool, 1)
	validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), done)
	testingutils.SubscribeToEventOrFail(t, bc.Connection, testingutils.CloudEventType, validator)

	// create cloudevent
	ce := testingutils.StructuredCloudEventPayloadWithCleanEventType
	event := cloudevents.NewEvent()
	event.SetType(testingutils.CloudEventType)
	err = json.Unmarshal([]byte(ce), &event)
	assert.Nil(t, err)

	// send cloudevent
	status, err := sender.Send(ctx, &event)
	assert.Nil(t, err)
	assert.Equal(t, status, http.StatusNoContent)

	// wait for subscriber to receive the messages
	if err := testingutils.WaitForChannelOrTimeout(done, time.Second*3); err != nil {
		t.Fatalf("Subscriber did not receive the message with error: %v", err)
	}

	// close connection
	bc.Connection.Close()

	// send the cloudevent again, the message is not delivered cause the connection was closed
	status, err = sender.Send(ctx, &event)
	assert.NotNil(t, err)
	assert.Equal(t, status, http.StatusBadGateway)

	// send the cloudevent again, the reconnection should work
	status, err = sender.Send(ctx, &event)
	assert.Nil(t, err)
	assert.Equal(t, status, http.StatusNoContent)
}
