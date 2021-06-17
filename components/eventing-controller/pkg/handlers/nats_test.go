package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
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
	err = SendEventToNATS(&natsClient, data)
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
	err = SendEventToNATS(&natsClient, newData)
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

func TestIsValidSubscription(t *testing.T) {
	g := NewWithT(t)

	natsPort := 5223
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	//defer natsServer.Shutdown()

	// Create Nats client
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

	// get filter
	filter := sub.Spec.Filter.Filters[0]
	subject, err := getSubject(filter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// get internal key
	key := getKey(sub, subject)
	g.Expect(key).To(Not(BeEmpty()))
	natsSub := natsClient.subscriptions[key]
	g.Expect(natsSub).To(Not(BeNil()))

	// the associated Nats subscription should be valid
	g.Expect(natsSub.IsValid()).To(BeTrue())

	// check the mapping of Kyma subscription and Nats subscription
	nsn := getKymaSubscriptionNamespacedName(key, natsSub)
	g.Expect(nsn.Namespace).To(BeIdenticalTo(sub.Namespace))
	g.Expect(nsn.Name).To(BeIdenticalTo(sub.Name))

	// check that no invalid subscriptions exist
	inactiveNsn := natsClient.GetInvalidSubscriptions()
	g.Expect(len(*inactiveNsn)).To(BeZero())

	natsClient.connection.Close()

	// the associated Nats subscription should be valid
	g.Expect(natsSub.IsValid()).To(BeFalse())

	// the associated Nats subscription should not be valid anymore
	err = checkIsNotValid(natsSub, t)
	g.Expect(err).To(BeNil())

	// shutdown NATS
	natsServer.Shutdown()

	// the associated Nats subscription should not be valid anymore
	err = checkIsNotValid(natsSub, t)
	g.Expect(err).To(BeNil())

	// check that only one invalid subscription exist
	inactiveNsn = natsClient.GetInvalidSubscriptions()
	g.Expect(len(*inactiveNsn)).To(BeIdenticalTo(1))

	// Restart Nats server
	natsServer = eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	// TODO: the controller should react and inject Kyma subscriptions into NATS for all "invalid" subscriptions
	// the associated Nats subscription should be valid again
	/*
	err = checkIsValid(natsSub, t)
	g.Expect(err).To(BeNil())
	*/


}

func checkIsValid(sub *nats.Subscription, t *testing.T) error {
	return checkValidity(sub, true, t)
}

func checkIsNotValid(sub *nats.Subscription, t *testing.T) error {
	return checkValidity(sub, false, t)
}

func checkValidity(sub *nats.Subscription, toCheckIsValid bool, t *testing.T) error {
	maxAttempts := uint(5)
	delay := time.Second
	err := retry.Do(
		func() error {
			if toCheckIsValid == sub.IsValid() {
				return nil
			}
			if toCheckIsValid {
				return errors.New("still invalid Nats subscription, expected valid")
			}
			return errors.New("still valid Nats subscription, expected invalid")
		},
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(maxAttempts),
		retry.OnRetry(func(n uint, err error) { t.Logf("[%v] try failed: %s", n, err) }),
	)
	return err
}