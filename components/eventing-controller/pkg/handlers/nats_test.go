package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go/v3"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/nats-io/nats.go"
	. "github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
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
			expectedErr:        errors.New("id: MUST be a non-empty string\n"), //nolint:golint
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
		URL:           natsServer.ClientURL(),
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
	_, err = natsClient.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}
	g.Expect(sub.Status.Config).NotTo(BeNil()) // It should apply the defaults
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(defaultMaxInflight))

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

func TestMultipleSubscriptionsToSameEvent(t *testing.T) {
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
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 1
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, defaultLogger)

	if err := natsClient.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to Nats server failed: %v", err)
	}

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()

	// Shutting down subscriber
	defer subscriber.Shutdown()

	// Check if the subscriber is running or not by checking its store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// Create 3 subscriptions having the same sink and the same event type
	var subs [3]*eventingv1alpha1.Subscription
	for i := 0; i < len(subs); i++ {
		subs[i] = eventingtesting.NewSubscription(fmt.Sprintf("sub-%d", i), "foo", eventingtesting.WithNotCleanEventTypeFilter)
		subs[i].Spec.Sink = subscriberReceiveURL
		if _, err = natsClient.SyncSubscription(subs[i], cleaner); err != nil {
			t.Fatalf("sync subscription %s failed: %v", subs[i].Name, err)
		}
		g.Expect(subs[i].Status.Config).NotTo(BeNil()) // It should apply the defaults
		g.Expect(subs[i].Status.Config.MaxInFlightMessages).To(Equal(defaultMaxInflight))
	}

	// Send only one event. It should be multiplexed to 3 by NATS, cause 3 subscriptions exist
	data := "sampledata"
	err = SendEventToNATS(natsClient, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	// Check for the 3 events that should be received by the subscriber
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	for i := 0; i < len(subs); i++ {
		if err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL); err != nil {
			t.Fatalf("subscriber did not receive the event: %v", err)
		}
	}

	// Delete all 3 subscription
	for i := 0; i < len(subs); i++ {
		if err = natsClient.DeleteSubscription(subs[i]); err != nil {
			t.Fatalf("delete subscription %s failed: %v", subs[i].Name, err)
		}
	}

	// Check if all subscriptions are deleted in NATS
	// Send an event again which should not be delivered to subscriber
	newData := "test-data"
	err = SendEventToNATS(natsClient, newData)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	// Check for the event that did not reach the subscriber
	// Store should never return newdata hence CheckEvent should fail to match newdata
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	err = subscriber.CheckEvent(notExpectedNewDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("check event failed: %v", err)
	}
	// newdata was received by the subscriber meaning the subscription was not deleted
	if err == nil {
		t.Fatal("subscriptions still exists in NATS")
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
		URL:           natsServer.ClientURL(),
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
	filter := &eventingv1alpha1.BEBFilter{
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
	sub.Spec.Filter = &eventingv1alpha1.BEBFilters{
		Filters: []*eventingv1alpha1.BEBFilter{filter, filter},
	}
	sub.Spec.Sink = subscriberReceiveURL
	idFunc := func(et string) (string, error) { return et, nil }
	if _, err := natsClient.SyncSubscription(sub, eventtype.CleanerFunc(idFunc)); err != nil {
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

func TestSubscriptionWithMaxInFlightChange(t *testing.T) {
	g := NewWithT(t)

	natsPort := 5223
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)

	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	filter := sub.Spec.Filter.Filters[0]
	subject, err := getCleanSubject(filter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	g.Expect(sub.Status.Config).NotTo(BeNil())
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(defaultSubsConfig.MaxInFlightMessages))

	// get internal key
	var key string
	var natsSub *nats.Subscription
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		key = createKey(sub, subject, i)
		natsSub = natsBackend.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
	}
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))

	// check that no invalid subscriptions exist
	invalidNsn := natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeZero())

	sub.Spec.Config = &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 7}
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	g.Expect(sub.Status.Config).NotTo(BeNil())
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(sub.Spec.Config.MaxInFlightMessages))
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		key = createKey(sub, subject, i)
		natsSub = natsBackend.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
	}
	g.Expect(len(natsBackend.subscriptions)).To(Equal(sub.Spec.Config.MaxInFlightMessages))
	// check that no invalid subscriptions exist
	invalidNsn = natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeZero())
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
		URL:           natsServer.ClientURL(),
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
	_, err = natsClient.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get filter
	filter := sub.Spec.Filter.Filters[0]
	subject, err := getCleanSubject(filter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	g.Expect(sub.Status.Config).NotTo(BeNil())
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(defaultSubsConfig.MaxInFlightMessages))

	// get internal key
	var key string
	var natsSub *nats.Subscription
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
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
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(sub.Status.Config.MaxInFlightMessages))

	// restart NATS server
	natsServer = eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	// check that only one invalid subscription still exist, the controller is not running...
	invalidNsn = natsClient.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(sub.Status.Config.MaxInFlightMessages))

}

func TestSubscriptionUsingCESDK(t *testing.T) {
	g := NewWithT(t)
	natsPort := 5222
	subscriberPort := 8080
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 1
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, defaultLogger)

	err = natsClient.Initialize(env.Config{})
	g.Expect(err).To(BeNil())

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()

	// Shutting down subscriber
	defer subscriber.Shutdown()

	// Check subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	g.Expect(err).To(BeNil())

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsClient.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())
	g.Expect(sub.Status.Config).NotTo(BeNil()) // It should apply the defaults
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(defaultMaxInflight))

	subject := eventingtesting.CloudEventType

	// Send a binary CE
	err = SendBinaryCloudEventToNATS(natsClient, subject)
	g.Expect(err).To(BeNil())
	// Check for the event
	err = subscriber.CheckEvent(eventingtesting.CloudEventData, subscriberCheckURL)
	g.Expect(err).To(BeNil())

	//  Send a structured CE
	err = SendStructuredCloudEventToNATS(natsClient, subject)
	g.Expect(err).To(BeNil())
	// Check for the event
	err = subscriber.CheckEvent("\""+eventingtesting.EventData+"\"", subscriberCheckURL)
	g.Expect(err).To(BeNil())

	// Delete subscription
	err = natsClient.DeleteSubscription(sub)
	g.Expect(err).To(BeNil())
}

func TestRetryUsingCESDK(t *testing.T) {
	g := NewWithT(t)
	natsPort := 5222
	subscriberPort := 8080
	subscriberCheckDataURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)
	subscriberCheckRetriesURL := fmt.Sprintf("http://127.0.0.1:%d/check_retries", subscriberPort)
	subscriberServerErrorURL := fmt.Sprintf("http://127.0.0.1:%d/return500", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer natsServer.Shutdown()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	maxRetries := 3
	defaultSubscriptionConfig := env.DefaultSubscriptionConfig{
		MaxInFlightMessages:   1,
		DispatcherRetryPeriod: time.Second,
		DispatcherMaxRetries:  maxRetries,
	}
	natsClient := NewNats(natsConfig, defaultSubscriptionConfig, defaultLogger)

	err = natsClient.Initialize(env.Config{})
	g.Expect(err).To(BeNil())

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()

	// Shutting down subscriber
	defer subscriber.Shutdown()

	// Check subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckDataURL)
	g.Expect(err).To(BeNil())

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationName, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithEventTypeFilter)
	sub.Spec.Sink = subscriberServerErrorURL
	_, err = natsClient.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())
	g.Expect(sub.Status.Config).NotTo(BeNil()) // It should apply the defaults

	subject := eventingtesting.CloudEventType

	//  Send a structured CE
	err = SendStructuredCloudEventToNATS(natsClient, subject)
	g.Expect(err).To(BeNil())

	// Check that the retries are done and that the sent data was correctly received
	err = subscriber.CheckRetries(maxRetries, "\""+eventingtesting.EventData+"\"", subscriberCheckDataURL, subscriberCheckRetriesURL)
	g.Expect(err).To(BeNil())

	// Delete subscription
	err = natsClient.DeleteSubscription(sub)
	g.Expect(err).To(BeNil())
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
