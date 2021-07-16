package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	. "github.com/onsi/gomega"
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
	g := NewWithT(t)
	natsPort := 5222
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	natsConfig := env.NatsConfig{
		Url:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 9
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, defaultLogger)

	if err := natsClient.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to Nats server failed: %v", err)
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
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, appliedSubConfig, err := natsClient.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}
	g.Expect(appliedSubConfig).NotTo(BeNil())
	g.Expect(appliedSubConfig.MaxInFlightMessages).To(Equal(defaultMaxInflight))

	data := "sampledata"
	// Send an event
	err = SendEventToNATS(natsClient, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
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
		t.Fatalf("delete subscription failed: %v", err)
	}

	newData := "test-data"
	// Send an event
	err = SendEventToNATS(natsClient, newData)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	// Check for the event that it did not reach subscriber
	// Store should never return newdata hence CheckEvent should fail to match newdata
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	err = subscriber.CheckEvent(notExpectedNewDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("check event failed: %v", err)
	}
	// newdata was received by the subscriber meaning the subscription was not deleted
	if err == nil {
		t.Fatal("subscription still exists in Nats")
	}
}

func TestSubscriptionWithDuplicateFilters(t *testing.T) {
	g := NewWithT(t)

	natsPort := 5223
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	natsConfig := env.NatsConfig{
		Url:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}, defaultLogger)

	if err := natsClient.Initialize(env.Config{}); err != nil {
		t.Fatalf("start NATS eventing backend failed: %s", err.Error())
	}

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown()

	// Check subscriber is running or not by checking the store
	if err := subscriber.CheckEvent("", subscriberCheckURL); err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	sub := eventingtesting.NewSubscription("sub", "foo")
	filter := &eventingv1alpha1.BebFilter{
		EventSource: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    "",
		},
		EventType: &eventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    eventingtesting.OrderCreatedEventType,
		},
	}
	sub.Spec.Filter = &eventingv1alpha1.BebFilters{
		Filters: []*eventingv1alpha1.BebFilter{filter, filter},
	}
	sub.Spec.Sink = subscriberReceiveURL
	idFunc := func(et string) (string, error) { return et, nil }
	if _, _, err := natsClient.SyncSubscription(sub, eventtype.CleanerFunc(idFunc)); err != nil {
		t.Fatalf("sync subscription failed: %s", err.Error())
	}

	data := "sampledata"
	// Send an event
	if err := SendEventToNATS(natsClient, data); err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	if err := subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL); err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// There should be no more!
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	g.Expect(err).Should(HaveOccurred())
}

func TestIsValidSubscription(t *testing.T) {
	g := NewWithT(t)

	natsPort := 5223
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// Create NATS client
	natsConfig := env.NatsConfig{
		Url:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	natsClient := NewNats(natsConfig, defaultSubsConfig, defaultLogger)

	if err := natsClient.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, appliedConfig, err := natsClient.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get filter
	filter := sub.Spec.Filter.Filters[0]
	subject, err := createSubject(filter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	g.Expect(appliedConfig).NotTo(BeNil())
	g.Expect(appliedConfig.MaxInFlightMessages).To(Equal(defaultSubsConfig.MaxInFlightMessages))

	// get internal key
	var key string
	var natsSub *nats.Subscription
	for i := 0; i < appliedConfig.MaxInFlightMessages; i++ {
		key = createKey(sub, subject, i)
		g.Expect(key).To(Not(BeEmpty()))
		natsSub = natsClient.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
	}
	// check the mapping of Kyma subscription and Nats subscription
	nsn := createKymaSubscriptionNamespacedName(key, natsSub)
	g.Expect(nsn.Namespace).To(BeIdenticalTo(sub.Namespace))
	g.Expect(nsn.Name).To(BeIdenticalTo(sub.Name))

	// the associated Nats subscription should be valid
	g.Expect(natsSub.IsValid()).To(BeTrue())

	// check that no invalid subscriptions exist
	invalidNsn := natsClient.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeZero())

	natsClient.connection.Close()

	// the associated NATS subscription should not be valid anymore
	err = checkIsNotValid(natsSub, t)
	g.Expect(err).To(BeNil())

	// shutdown NATS
	natsServer.Shutdown()

	// the associated NATS subscription should not be valid anymore
	err = checkIsNotValid(natsSub, t)
	g.Expect(err).To(BeNil())

	// check that only one invalid subscription exist
	invalidNsn = natsClient.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(appliedConfig.MaxInFlightMessages))

	// restart NATS server
	natsServer = eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	// check that only one invalid subscription still exist, the controller is not running...
	invalidNsn = natsClient.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(appliedConfig.MaxInFlightMessages))

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
