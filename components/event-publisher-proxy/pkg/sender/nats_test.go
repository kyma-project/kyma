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
	connection, err := pkgnats.ConnectToNats(natsServer.ClientURL(), true, 1, time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, connection)

	// create message sender
	ctx := context.Background()
	sender := NewNatsMessageSender(ctx, connection, logger)

	// subscribe to subject
	done := make(chan bool, 1)
	validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), done)
	testingutils.SubscribeToEventOrFail(t, connection, testingutils.CloudEventType, validator)

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
