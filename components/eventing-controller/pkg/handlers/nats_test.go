package handlers

import (
	"errors"
	"fmt"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
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
			name: "data without quotes",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("foo-data", "id", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: eventingtesting.NewCloudEvent("foo-data", "id", "foosource", eventTime, "fooeventtype", t),
			expectedErr:        nil,
		},
		{
			name: "data with quotes",
			natsMsg: nats.Msg{
				Subject: "fooeventtype",
				Reply:   "",
				Header:  nil,
				Data:    []byte(NewNatsMessagePayload("\\\"foo-data\\\"", "id", "foosource", eventTime, "fooeventtype")),
				Sub:     nil,
			},
			expectedCloudEvent: eventingtesting.NewCloudEvent("\\\"foo-data\\\"", "id", "foosource", eventTime, "fooeventtype", t),
			expectedErr:        nil,
		},
		{
			name: "natsMessage which is an invalid Cloud Event with empty id",
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
				t.Errorf("should not give error, got: %v", err)
				return
			}
			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("received nil error, expected: %v got: %v", tc.expectedErr, err)
					return
				}
				if tc.expectedErr.Error() != err.Error() {
					t.Errorf("received wrong error, expected: %v got: %v", tc.expectedErr, err)
				}
				return
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
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/subscribe", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	natsURL := natsServer.ClientURL()
	natsClient := Nats{
		Subscriptions: make(map[string]*nats.Subscription),
		Config: env.NatsConfig{
			Url:           natsURL,
			MaxReconnects: 2,
			ReconnectWait: time.Second,
		},
		Log: ctrl.Log.WithName("reconciler").WithName("Subscription"),
	}
	err := natsClient.Initialize()
	if err != nil {
		t.Fatalf("failed to connect to Nats Server: %v", err)
	}

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort), t)
	subscriber.Start()

	// Shutting down subscriber
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

	data := "sampledata"
	// Send an event
	err = SendEvent(&natsClient, data)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Check for the event
	err = subscriber.CheckEvent(data, subscriberCheckURL)
	if err != nil {
		t.Errorf("subscriber did not receive the event: %v", err)
	}

	// Delete subscription
	err = natsClient.DeleteSubscription(sub)
	if err != nil {
		t.Errorf("failed to delete subscription: %v", err)
	}

	newData := "newdata"
	// Send an event
	err = SendEvent(&natsClient, newData)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Check for the event that it did not reach subscriber
	err = subscriber.CheckEvent(newData, subscriberCheckURL)
	if err == nil {
		t.Error("subscription still exists in Nats")
	}
}

func SendEvent(natsClient *Nats, data string) error {
	eventTime := time.Now().Format(time.RFC3339)
	eventType := "kyma.ev2.poc.event1.v1"
	sampleEvent := NewNatsMessagePayload(data, "id", "foosource", eventTime, eventType)
	return natsClient.Connection.Publish(eventType, []byte(sampleEvent))
}

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}
