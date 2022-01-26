package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"

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

var (
	nextPort = &portGenerator{port: 5223}
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
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
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
	defaultMaxInflight := 9
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, nil, defaultLogger)

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

// TestNatsSubAfterSync_NoChange tests the SyncSubscription method
// when there is no change in the subscription then the method should
// not re-create NATS subjects on nats-server
func TestNatsSubAfterSync_NoChange(t *testing.T) {
	g := NewWithT(t)

	//// ######  Setup test assets ######
	// setup logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// create subscribers servers for testing
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown() // defer the shutdown of subscriber

	// check if the subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer) // defer the shutdown of nats-server

	// Create NATS backend handler instance
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	//// ###### Test logic ######
	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// set metadata on nats subscription
		if err := natsSub.SetPendingLimits(msgLimit, bytesLimit); err != nil {
			t.Fatalf("set pending limits for nats subscription failed: %v", err)
		}
	}

	// Now, sync the subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are now same then it means that nats subscription
		// were not re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Equal(msgLimit))
		g.Expect(subBytesLimit).To(Equal(msgLimit))
	}

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}
}

// TestNatsSubAfterSync_SinkChange tests the SyncSubscription method
// when only the sink is changed in subscription, then it should not re-create
// NATS subjects on nats-server
func TestNatsSubAfterSync_SinkChange(t *testing.T) {
	g := NewWithT(t)

	//// ######  Setup test assets ######
	// Setup logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// create two subscribers, as we need two sinks for this test
	natsPort := nextPort.get()
	subscriber1Port := nextPort.get()
	subscriber2Port := nextPort.get()

	subscriber1ReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriber1Port)
	subscriber2ReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriber2Port)
	subscriber1CheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriber1Port)
	subscriber2CheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriber2Port)

	// create new subscribers
	subscriber1 := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriber1Port))
	subscriber2 := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriber2Port))
	subscriber1.Start()
	subscriber2.Start()

	// shutting down subscribers
	defer subscriber1.Shutdown()
	defer subscriber2.Shutdown()

	// check if subscribers are running or not by checking the stores
	err = subscriber1.CheckEvent("", subscriber1CheckURL)
	if err != nil {
		t.Fatalf("subscriber 1 did not receive the event: %v", err)
	}
	err = subscriber2.CheckEvent("", subscriber2CheckURL)
	if err != nil {
		t.Fatalf("subscriber 2 did not receive the event: %v", err)
	}

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer)

	// Create NATS backend handler instance
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// ##### Test logic ######

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriber1ReceiveURL
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	err = subscriber1.CheckEvent(expectedDataInStore, subscriber1CheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// set metadata on NATS subscriptions
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// set metadata on nats subscription
		if err := natsSub.SetPendingLimits(msgLimit, bytesLimit); err != nil {
			t.Fatalf("set pending limits for nats subscription failed: %v", err)
		}
	}

	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2ReceiveURL
	// Sync subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are now same then it means that nats subscription
		// were not re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Equal(msgLimit))
		g.Expect(subBytesLimit).To(Equal(msgLimit))
	}

	// Test if the subscription is working for new sink only
	// Send an event
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	// Old sink should not have received the event
	err = subscriber1.CheckEvent(expectedDataInStore, subscriber1CheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("subscriber 1 check event failed: %v", err)
	}

	// New sink should have received the event
	err = subscriber2.CheckEvent(expectedDataInStore, subscriber2CheckURL)
	if err != nil {
		t.Fatalf("subscriber 2 check event failed: %v", err)
	}
}

// TestNatsSubAfterSync_FiltersChange tests the SyncSubscription method
// when the filters are changed in subscription
func TestNatsSubAfterSync_FiltersChange(t *testing.T) {
	g := NewWithT(t)

	//// ######  Setup test assets ######
	// setup logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// create subscribers servers for testing
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown() // defer the shutdown of subscriber

	// check if the subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer) // defer the shutdown of nats-server

	// Create NATS backend handler instance
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	//// ###### Test logic ######
	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for key := range natsBackend.subscriptions {
		// set metadata on nats subscription
		if err := natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit); err != nil {
			t.Fatalf("set pending limits for nats subscription failed: %v", err)
		}
	}

	// Now, change the filter in subscription
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", eventingtesting.OrderCreatedEventTypeNotClean)
	// Sync the subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get new cleaned subject
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(newSubject).To(Not(BeEmpty()))

	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, newSubject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are NOT same then it means that nats subscriptions
		// were re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Not(Equal(msgLimit)))
		g.Expect(subBytesLimit).To(Not(Equal(msgLimit)))
	}

	// Test if subscription is working for new subject only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)

	// Send an event on old subject
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	// The sink should not receive any event for old subject
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("check event failed: %v", err)
	}

	// Now, send an event on new subject
	err = SendEventToNATSOnEventType(natsBackend, newSubject, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	// The sink should receive the event for new subject
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("check event failed: %v", err)
	}
}

// TestNatsSubAfterSync_FilterAdded tests the SyncSubscription method
// when a new filter is added in subscription
func TestNatsSubAfterSync_FilterAdded(t *testing.T) {
	g := NewWithT(t)

	//// ######  Setup test assets ######
	// setup logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// create subscribers servers for testing
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown() // defer the shutdown of subscriber

	// check if the subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer) // defer the shutdown of nats-server

	// Create NATS backend handler instance
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	//// ###### Test logic ######
	// Create a subscription with single filter
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get cleaned subject
	firstSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(firstSubject).To(Not(BeEmpty()))

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for key := range natsBackend.subscriptions {
		// set metadata on nats subscription
		if err := natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit); err != nil {
			t.Fatalf("set pending limits for nats subscription failed: %v", err)
		}
	}

	// Now, add a new filter to subscription
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", eventingtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)

	// get new cleaned subject
	secondSubject, err := getCleanSubject(newFilter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(secondSubject).To(Not(BeEmpty()))

	// Sync the subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	expectedTotalNatsSubs := 2 * defaultSubsConfig.MaxInFlightMessages
	g.Expect(natsBackend.subscriptions).To(HaveLen(expectedTotalNatsSubs))

	// Verify that the nats subscriptions for first subject was not re-created
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, firstSubject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are same then it means that nats subscriptions
		// were not re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Equal(msgLimit))
		g.Expect(subBytesLimit).To(Equal(msgLimit))
	}

	// Test if subscription is working for both subjects
	// Send an event on first subject
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	// The sink should receive event for first subject
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("check event failed for first subject: %v", err)
	}

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATSOnEventType(natsBackend, secondSubject, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	// The sink should receive the event for second subject
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("check event failed for second subject: %v", err)
	}
}

// TestNatsSubAfterSync_FilterRemoved tests the SyncSubscription method
// when a filter is removed from subscription
func TestNatsSubAfterSync_FilterRemoved(t *testing.T) {
	g := NewWithT(t)

	//// ######  Setup test assets ######
	// setup logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// create subscribers servers for testing
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown() // defer the shutdown of subscriber

	// check if the subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer) // defer the shutdown of nats-server

	// Create NATS backend handler instance
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	//// ###### Test logic ######
	// Create a subscription with two filters
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL

	// add a second filter
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", eventingtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)

	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get cleaned subjects
	firstSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(firstSubject).To(Not(BeEmpty()))

	secondSubject, err := getCleanSubject(sub.Spec.Filter.Filters[1], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(secondSubject).To(Not(BeEmpty()))

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	expectedTotalNatsSubs := 2 * defaultSubsConfig.MaxInFlightMessages
	g.Expect(natsBackend.subscriptions).To(HaveLen(expectedTotalNatsSubs))

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	for key := range natsBackend.subscriptions {
		// set metadata on nats subscription
		if err := natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit); err != nil {
			t.Fatalf("set pending limits for nats subscription failed: %v", err)
		}
	}

	// Now, remove the second filter from subscription
	sub.Spec.Filter.Filters = sub.Spec.Filter.Filters[:1]

	// Sync the subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// Check if total existing NATS subscriptions are correct
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))

	// Verify that the nats subscriptions for first subject was not re-created
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, firstSubject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are same then it means that nats subscriptions
		// were not re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Equal(msgLimit))
		g.Expect(subBytesLimit).To(Equal(msgLimit))
	}

	// Test if subscription is working for first subject only
	// Send an event on first subject
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	// The sink should receive event for first subject
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("check event failed for first subject: %v", err)
	}

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	err = SendEventToNATSOnEventType(natsBackend, secondSubject, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	// The sink should not receive the event for second subject
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil && !strings.Contains(err.Error(), "check event after retries failed") {
		t.Fatalf("check event failed: %v", err)
	}
}

// Test_isNatsSubAssociatedWithKymaSub tests the isNatsSubAssociatedWithKymaSub method
func Test_isNatsSubAssociatedWithKymaSub(t *testing.T) {
	g := NewWithT(t)

	//// ######  Setup test assets ######
	// create subscription 1 and its nats subscription
	cleanSubject1 := "subOne"
	sub1 := eventingtesting.NewSubscription(cleanSubject1, "foo", eventingtesting.WithNotCleanEventTypeFilter)
	natsSub1Key := createKey(sub1, cleanSubject1, 0)
	natsSub1 := &nats.Subscription{
		Subject: cleanSubject1,
	}

	// create subscription 2 and its nats subscription
	cleanSubject2 := "subOneTwo"
	sub2 := eventingtesting.NewSubscription(cleanSubject2, "foo", eventingtesting.WithNotCleanEventTypeFilter)
	natsSub2Key := createKey(sub2, cleanSubject2, 0)
	natsSub2 := &nats.Subscription{
		Subject: cleanSubject2,
	}

	//// ###### Test logic ######
	// Should return true because natsSub1 is associated with sub1
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub1Key, natsSub1, sub1)).To(Equal(true))
	// Should return true because natsSub2 is associated with sub2
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub2Key, natsSub2, sub2)).To(Equal(true))

	// Should return false because natsSub1 is NOT associated with sub2
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub1Key, natsSub1, sub2)).To(Equal(false))
	// Should return false because natsSub2 is NOT associated with sub1
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub2Key, natsSub2, sub1)).To(Equal(false))

}

// TestNatsSubAfterSync_MultipleSubs tests the SyncSubscription method
// when there are two subscriptions and the filter is changed in one subscription
// it should not affect the NATS subscriptions of other Kyma subscriptions
func TestNatsSubAfterSync_MultipleSubs(t *testing.T) {
	g := NewWithT(t)

	// // ######  Setup test assets ######
	// setup logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}

	// create subscribers servers for testing
	// create subscribers servers for testing
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	defer subscriber.Shutdown() // defer the shutdown of subscriber
	defer subscriber.Shutdown()

	// check if the subscriber is running or not by checking the store
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer) // defer the shutdown of nats-server

	// Create NATS backend handler instance
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	if err := natsBackend.Initialize(env.Config{}); err != nil {
		t.Fatalf("connect to NATS server failed: %v", err)
	}

	// Prepare event-type cleaner
	application := applicationtest.NewApplication(eventingtesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	cleaner := eventtype.NewCleaner(eventingtesting.EventTypePrefix, applicationLister, defaultLogger)

	// // ###### Test logic ######
	// Create two subscriptions with single filter
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())

	sub2 := eventingtesting.NewSubscription("sub2", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub2.Spec.Sink = subscriberReceiveURL
	_, err = natsBackend.SyncSubscription(sub2, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	// check we have correct number of total subscriptions
	expectedTotalNatsSubs := 2 * defaultSubsConfig.MaxInFlightMessages // Because we have two subscriptions
	g.Expect(len(natsBackend.subscriptions)).To(Equal(expectedTotalNatsSubs))
	for key := range natsBackend.subscriptions {
		// set metadata on nats subscription
		if err := natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit); err != nil {
			t.Fatalf("set pending limits for nats subscription failed: %v", err)
		}
	}

	// Now, change the filter in subscription 1
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", eventingtesting.OrderCreatedEventTypeNotClean)
	// Sync the subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	if err != nil {
		t.Fatalf("sync subscription failed: %v", err)
	}

	// get new cleaned subject from subscription 1
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(newSubject).To(Not(BeEmpty()))

	// check we have correct number of total subscriptions
	expectedTotalNatsSubs = 2 * defaultSubsConfig.MaxInFlightMessages // Because we have two subscriptions
	g.Expect(len(natsBackend.subscriptions)).To(Equal(expectedTotalNatsSubs))

	// check if the NATS subscription are NOT the same after sync for subscription 1
	// because the subscriptions should have being re-created for new subject
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, newSubject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are NOT same then it means that nats subscriptions
		// were re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Not(Equal(msgLimit)))
		g.Expect(subBytesLimit).To(Not(Equal(msgLimit)))
	}

	// get cleaned subject for subscription 2
	cleanSubjectSub2, err := getCleanSubject(sub2.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(cleanSubjectSub2).To(Not(BeEmpty()))

	// check if the NATS subscription are same after sync for subscription 2
	// because the subscriptions should not have being re-created as
	// subscription 2 was not modified
	for i := 0; i < sub2.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub2, cleanSubjectSub2, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are same then it means that nats subscriptions
		// were not re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Equal(msgLimit))
		g.Expect(subBytesLimit).To(Equal(msgLimit))
	}
}

func TestMultipleSubscriptionsToSameEvent(t *testing.T) {
	g := NewWithT(t)
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
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
	defaultMaxInflight := 1
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, nil, defaultLogger)

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

	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

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
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}, nil, defaultLogger)

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

	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
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
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)

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

	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)

	// Start NATS server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer)

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
	natsClient := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)

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
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 1
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, nil, defaultLogger)

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
	natsPort := nextPort.get()
	subscriberPort := nextPort.get()
	subscriberCheckDataURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)
	subscriberCheckRetriesURL := fmt.Sprintf("http://127.0.0.1:%d/check_retries", subscriberPort)
	subscriberServerErrorURL := fmt.Sprintf("http://127.0.0.1:%d/return500", subscriberPort)

	// Start Nats server
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	defer eventingtesting.ShutDownNATSServer(natsServer)

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
	natsClient := NewNats(natsConfig, defaultSubscriptionConfig, nil, defaultLogger)

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

func TestSubscription_NATSServerRestart(t *testing.T) {
	g := NewWithT(t)
	natsServer, natsPort := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(t, kymalogger.INFO)
	// The reconnect configs should be large enough to cover the NATS server restart period
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 10,
		ReconnectWait: 3 * time.Second,
	}
	natsClient := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: 10}, nil, defaultLogger)
	g.Expect(natsClient.Initialize(env.Config{})).To(BeNil())

	subscriber, subscriberPort := startSubscriber()
	defer subscriber.Shutdown()
	subscriberReceiveURL := fmt.Sprintf("http://127.0.0.1:%d/store", subscriberPort)
	subscriberCheckURL := fmt.Sprintf("http://127.0.0.1:%d/check", subscriberPort)
	// Check subscriber is running or not by checking the store
	g.Expect(subscriber.CheckEvent("", subscriberCheckURL)).To(BeNil())

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriberReceiveURL
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	_, err := natsClient.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())

	ev1data := "sampledata"
	g.Expect(SendEventToNATS(natsClient, ev1data)).To(Succeed())
	expectedEv1Data := fmt.Sprintf("\"%s\"", ev1data)
	g.Expect(subscriber.CheckEvent(expectedEv1Data, subscriberCheckURL)).To(Succeed())

	natsServer.Shutdown()
	g.Eventually(func() bool {
		return natsClient.connection.IsConnected()
	}, 60*time.Second, 2*time.Second).Should(BeFalse())
	_ = eventingtesting.RunNatsServerOnPort(natsPort)
	g.Eventually(func() bool {
		return natsClient.connection.IsConnected()
	}, 60*time.Second, 2*time.Second).Should(BeTrue())

	// After reconnect, event delivery should work again
	ev2data := "newsampledata"
	g.Expect(SendEventToNATS(natsClient, ev2data)).To(Succeed())
	expectedEv2Data := fmt.Sprintf("\"%s\"", ev2data)
	g.Expect(subscriber.CheckEvent(expectedEv2Data, subscriberCheckURL)).To(Succeed())
}

func startNATSServer() (*server.Server, int) {
	natsPort := nextPort.get()
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	return natsServer, natsPort
}

func startSubscriber() (*eventingtesting.Subscriber, int) {
	subscriberPort := nextPort.get()
	subscriber := eventingtesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	return subscriber, subscriberPort
}

func createEventTypeCleaner(eventTypePrefix, applicationName string, logger *logger.Logger) eventtype.Cleaner {
	application := applicationtest.NewApplication(applicationName, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	return eventtype.NewCleaner(eventTypePrefix, applicationLister, logger)
}

func getLogger(t *testing.T, level kymalogger.Level) *logger.Logger {
	l, err := logger.New(string(kymalogger.JSON), string(level))
	if err != nil {
		t.Fatalf("initialize logger failed: %v", err)
	}
	return l
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

type portGenerator struct {
	lock sync.Mutex
	port int
}

func (pg *portGenerator) get() int {
	pg.lock.Lock()
	defer pg.lock.Unlock()
	p := pg.port
	pg.port++
	return p
}
