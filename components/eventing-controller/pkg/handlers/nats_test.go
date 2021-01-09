package handlers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	pkgerrors "github.com/pkg/errors"

	"github.com/avast/retry-go"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
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
			expectedCloudEvent: NewCloudEvent("foo-data", "id", "foosource", eventTime, "fooeventtype", t),
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
			expectedCloudEvent: NewCloudEvent("\\\"foo-data\\\"", "id", "foosource", eventTime, "fooeventtype", t),
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

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}

func NewCloudEvent(data, id, source, eventTime, eventType string, t *testing.T) cev2event.Event {
	timeInRFC3339, err := time.Parse(time.RFC3339, eventTime)
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}
	dataContentType := "application/json"
	return cev2event.Event{
		Context: &cev2event.EventContextV1{
			Type: eventType,
			Source: types.URIRef{
				URL: url.URL{
					Path: source,
				},
			},
			ID:              id,
			DataContentType: &dataContentType,
			Time:            &types.Timestamp{Time: timeInRFC3339},
		},
		DataEncoded: []byte(data),
	}
}

func TestSubscription(t *testing.T) {
	natsPort := 5222
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/subscribe", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
	natsServer := reconcilertesting.RunNatsServerOnPort(natsPort)
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
	subscriber := NewSubscriber(fmt.Sprintf(":%d", subscriberPort), t)
	subscriber.Start()
	defer subscriber.Shutdown()

	// Check subscriber is running or not by checking the store
	err = CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Errorf("subscriber did not receive the event: %v", err)
	}

	// Create a subscription
	sub := reconcilertesting.NewSubscription("sub", "foo", reconcilertesting.WithFilterForNats)
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
	err = CheckEvent(data, subscriberCheckURL)
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

	// Check for the event that it did not reach subscriber so the old data should be still there in the subscriber
	err = CheckEvent(newData, subscriberCheckURL)
	if err == nil {
		t.Error("subscription still exists in Nats")
	}
}

type Subscriber struct {
	addr   string
	t      *testing.T
	server *http.Server
}

func NewSubscriber(addr string, t *testing.T) *Subscriber {
	return &Subscriber{
		addr: addr,
		t:    t,
	}
}

func (s Subscriber) Start() {
	store := ""
	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.t.Fatalf("failed to read data: %v", err)
		}
		store = string(data)
	})
	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(store))
		if err != nil {
			s.t.Fatalf("failed to write to the response: %v", err)
		}
	})

	server := http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		s.t.Fatal(server.ListenAndServe())
	}()
	s.server = &server
}

func (s Subscriber) Shutdown() {
	err := s.server.Close()
	if err != nil {
		s.t.Errorf("failed to shutdown Subscriber: %v", err)
	}
}

func SendEvent(natsClient *Nats, data string) error {
	eventTime := time.Now().Format(time.RFC3339)
	eventType := "kyma.ev2.poc.event1.v1"
	sampleEvent := NewNatsMessagePayload(data, "id", "foosource", eventTime, eventType)
	return natsClient.Connection.Publish(eventType, []byte(sampleEvent))
}

func CheckEvent(expectedData, subscriberCheckURL string) error {
	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(subscriberCheckURL)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			if string(body) != expectedData {
				return fmt.Errorf("subscriber did not get the event with data: \"%s\" yet...waiting", expectedData)
			}
			return nil
		},
		retry.Delay(2*time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
	)
	if err != nil {
		return pkgerrors.Wrapf(err, "failed to check the event after retries")
	}

	return nil
}
