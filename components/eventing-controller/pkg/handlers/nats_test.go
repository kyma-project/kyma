package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/nats-io/nats.go"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func TestConvertMsgToCE(t *testing.T) {
	eventTime := time.Now().Format(time.RFC3339)
	testCases := []struct {
		name               string
		natsMsg            nats.Msg
		expectedCloudEvent cev2event.Event
		expectedErr        error
	}{
		{
			name: "a valid Cloud Event NatsMessage",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("foo-data", "id", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: eventingtesting.NewCloudEvent("\"foo-data\"", "id", "foosource", eventTime, "fooeventtype", t),
			expectedErr:        nil,
		}, {
			name: "an invalid Cloud Event NatsMessage with empty id",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("foo-data", "", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: cev2event.New(cev2event.CloudEventsVersionV1),
			expectedErr:        errors.New("id: MUST be a non-empty string\n"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotCE, err := convertMsgToCE(&tc.natsMsg)
			if err != nil && tc.expectedErr == nil {
				t.Fatalf("Should not give error, got: %v", err)
			}
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Received nil error, expected: %v got: %v", tc.expectedErr, err)
				}
				if tc.expectedErr.Error() != err.Error() {
					t.Fatalf("Received wrong error, expected: %v got: %v", tc.expectedErr, err)
				}
				return
			}
			if gotCE == nil {
				t.Fatalf("Test failed, got nil cloudevent")
			}
			if !(gotCE.Subject() == tc.expectedCloudEvent.Subject()) ||
				!(gotCE.ID() == tc.expectedCloudEvent.ID()) ||
				!(gotCE.DataContentType() == tc.expectedCloudEvent.DataContentType()) ||
				!(gotCE.Source() == tc.expectedCloudEvent.Source()) ||
				!(gotCE.Time().String() == tc.expectedCloudEvent.Time().String()) ||
				!(string(gotCE.Data()) == string(tc.expectedCloudEvent.Data())) {
				t.Errorf("received wrong cloudevent, expected: %v got: %v", tc.expectedCloudEvent, gotCE)
			}
		})
	}
}

func TestSubscription(t *testing.T) {
	natsPort := 5222
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	natsURL := natsServer.ClientURL()
	natsClient := Nats{
		subscriptions: make(map[string]*nats.Subscription),
		config: env.NatsConfig{
			Url:           natsURL,
			MaxReconnects: 2,
			ReconnectWait: time.Second,
		},
		log: ctrl.Log.WithName("reconciler").WithName("Subscription"),
	}
	err := natsClient.Initialize(env.Config{})
	if err != nil {
		t.Fatalf("failed to connect to Nats Server: %v", err)
	}

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()

	// Shutting down subscriber
	defer subscriber.Shutdown()

	// Check subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, ctrl.Log.WithName("cleaner"))

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsClient.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("failed to Sync subscription: %v", err)
	}

	data := "sampledata"
	// Send an event
	err = SendEventToNats(&natsClient, data)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	// Check for the event
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Delete subscription
	err = natsClient.DeleteSubscription(sub)
	if err != nil {
		t.Fatalf("failed to delete subscription: %v", err)
	}

	newData := "datawhichdoesnotexist"
	// Send an event
	err = SendEventToNats(&natsClient, newData)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Check for the event that it did not reach subscriber
	// Store should never return newdata hence CheckEvent should fail to match newdata
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	err = subscriber.CheckEvent(notExpectedNewDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "failed to check the event after retries") {
		t.Fatalf("failed to CheckEvent: %v", err)
	}
	// newdata was received by the subscriber meaning the subscription was not deleted
	if err == nil {
		t.Fatal("subscription still exists in Nats")
	}
}
// TODO should be here and not in "handlers/utils.go"
/*
func SendEvent(natsClient *Nats, data string) error {
	// assumption: the event-type used for publishing is already cleaned from none-alphanumeric characters
	// because the publisher-application should have cleaned it already before publishing
	eventType := eventingtesting.EventType
	eventTime := time.Now().Format(time.RFC3339)
	sampleEvent := NewNatsMessagePayload(data, "id", eventingtesting.EventSource, eventTime, eventType)
	return natsClient.connection.Publish(eventType, []byte(sampleEvent))
}

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}
*/