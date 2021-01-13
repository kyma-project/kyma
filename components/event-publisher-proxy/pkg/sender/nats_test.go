package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-logr/logr"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	//ctrl "sigs.k8s.io/controller-runtime"
	//"github.com/go-logr/logr"
	"testing"
	"time"
)

func TestNatsSender(t *testing.T) {
	logger := logrus.New()
	logger.Info("TestNatsSender started")

	// test environment
	natsPort := 5222
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://localhost:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://localhost:%d/check", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	assert.NotNil(t, natsServer)
	defer natsServer.Shutdown()

	natsURL := natsServer.ClientURL()
	natsClient := handlers.Nats{
		Subscriptions: make(map[string]*nats.Subscription),
		Config: env.NatsConfig{
			Url:           natsURL,
			MaxReconnects: 2,
			ReconnectWait: time.Second,
		},
		Log: logr.Discard(),
	}
	err := natsClient.Initialize()
	if err != nil {
		t.Fatalf("failed to connect to Nats Server: %v", err)
	}

	// create a Nats sender
	natsUrl := natsServer.ClientURL()
	assert.NotEmpty(t, natsUrl)
	ctx := context.Background()
	sender := NewNatsMessageSender(ctx, natsUrl, logger)

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown()

	// Check subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Errorf("subscriber did not receive the event: %v", err)
	}

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithFilterForNats)
	sub.Spec.Sink = subscriberReceiveURL
	err = natsClient.SyncSubscription(sub)
	if err != nil {
		t.Fatalf("failed to Sync subscription: %v", err)
	}

	// create a CE event
	ce := testingutils.StructuredCloudEventPayload
	event := cloudevents.NewEvent()
	err = json.Unmarshal([]byte(ce), &event)
	assert.Nil(t, err)
	// set ce event type the same as the event type from subscription's filter. The Nats subject is defined by ce.Type
	event.SetType("kyma.ev2.poc.event1.v1")

	// send the event
	c, err := sender.Send(ctx, &event)
	assert.Nil(t, err)
	assert.Equal(t, c, http.StatusNoContent)

	// Check for the event
	expectedDataInStore := fmt.Sprintf("\"%s\"", "{\\\"foo\\\":\\\"bar\\\"}")
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Errorf("subscriber did not receive the event: %v", err)
	}
}
