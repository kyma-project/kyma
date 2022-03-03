package handlers

import (
	"context"
	"errors"
	"fmt"
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

var (
	nextSinkPort = &portGenerator{port: 8088}
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
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	data := "sampledata"
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())

	g.Expect(natsBackend.DeleteSubscription(sub)).Should(Succeed())

	newData := "test-data"
	g.Expect(SendEventToNATS(natsBackend, newData)).Should(Succeed())
	// Check for the event that it did not reach subscriber
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	g.Expect(subscriber.CheckEvent(notExpectedNewDataInStore)).ShouldNot(Succeed())
}

// TestNatsSubAfterSync_NoChange tests the SyncSubscription method
// when there is no change in the subscription then the method should
// not re-create NATS subjects on nats-server
func TestNatsSubAfterSync_NoChange(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
		// set metadata on nats subscription
		g.Expect(natsSub.SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}
	// Now, sync the subscription
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are now same then it means that NATS subscription
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
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
}

// TestNatsSubAfterSync_SinkChange tests the SyncSubscription method
// when only the sink is changed in subscription, then it should not re-create
// NATS subjects on nats-server
func TestNatsSubAfterSync_SinkChange(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber1 := eventingtesting.NewSubscriber()
	defer subscriber1.Shutdown()
	g.Expect(subscriber1.IsRunning()).To(BeTrue())
	subscriber2 := eventingtesting.NewSubscriber()
	defer subscriber2.Shutdown()
	g.Expect(subscriber2.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber1.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	g.Expect(subscriber1.CheckEvent(expectedDataInStore)).Should(Succeed())

	// set metadata on NATS subscriptions
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
		// set metadata on nats subscription
		g.Expect(natsSub.SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}
	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2.SinkURL
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())
	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())

		// check the metadata, if they are now same then it means that NATS subscription
		// were not re-created by SyncSubscription method
		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(subMsgLimit).To(Equal(msgLimit))
		g.Expect(subBytesLimit).To(Equal(msgLimit))
	}

	// Test if the subscription is working for new sink only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// Old sink should not have received the event, the new sink should have
	g.Expect(subscriber1.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
	g.Expect(subscriber2.CheckEvent(expectedDataInStore)).Should(Succeed())
}

// TestNatsSubAfterSync_FiltersChange tests the SyncSubscription method
// when the filters are changed in subscription
func TestNatsSubAfterSync_FiltersChange(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for key := range natsBackend.subscriptions {
		// set metadata on nats subscription
		g.Expect(natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}

	// Now, change the filter in subscription
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", eventingtesting.OrderCreatedEventTypeNotClean)
	// Sync the subscription
	addCleanEventTypesToStatus(sub, cleaner)
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	// get new cleaned subject
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(newSubject).To(Not(BeEmpty()))

	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
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
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// The sink should not receive any event for old subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
	// Now, send an event on new subject
	g.Expect(SendEventToNATSOnEventType(natsBackend, newSubject, data)).Should(Succeed())
	// The sink should receive the event for new subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
}

// TestNatsSubAfterSync_FilterAdded tests the SyncSubscription method
// when a new filter is added in subscription
func TestNatsSubAfterSync_FilterAdded(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	// // ###### Test logic ######
	// Create a subscription with single filter
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

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
		g.Expect(natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}

	// Now, add a new filter to subscription
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", eventingtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)

	// get new cleaned subject
	secondSubject, err := getCleanSubject(newFilter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(secondSubject).To(Not(BeEmpty()))

	addCleanEventTypesToStatus(sub, cleaner)
	// Sync the subscription
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

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
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// The sink should receive event for first subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATSOnEventType(natsBackend, secondSubject, data)).Should(Succeed())
	// The sink should receive the event for second subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
}

// TestNatsSubAfterSync_FilterRemoved tests the SyncSubscription method
// when a filter is removed from subscription
func TestNatsSubAfterSync_FilterRemoved(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)

	// // ###### Test logic ######
	// Create a subscription with two filters
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	// add a second filter
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", eventingtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

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
		g.Expect(natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}

	// Now, remove the second filter from subscription
	sub.Spec.Filter.Filters = sub.Spec.Filter.Filters[:1]
	addCleanEventTypesToStatus(sub, cleaner)
	// Sync the subscription
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

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
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// The sink should receive event for first subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATSOnEventType(natsBackend, secondSubject, data)).Should(Succeed())
	// The sink should NOT receive the event for second subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
}

// TestNatsSubAfterSync_MultipleSubs tests the SyncSubscription method
// when there are two subscriptions and the filter is changed in one subscription
// it should not affect the NATS subscriptions of other Kyma subscriptions
func TestNatsSubAfterSync_MultipleSubs(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	// Create a new subscriber
	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)

	// // ###### Test logic ######
	// Create two subscriptions with single filter
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).ShouldNot(HaveOccurred())

	sub2 := eventingtesting.NewSubscription("sub2", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	addCleanEventTypesToStatus(sub2, cleaner)
	err = natsBackend.SyncSubscription(sub2)
	g.Expect(err).ShouldNot(HaveOccurred())

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	// check we have correct number of total subscriptions
	expectedTotalNatsSubs := 2 * defaultSubsConfig.MaxInFlightMessages // Because we have two subscriptions
	g.Expect(len(natsBackend.subscriptions)).To(Equal(expectedTotalNatsSubs))
	for key := range natsBackend.subscriptions {
		// set metadata on nats subscription
		g.Expect(natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}

	// Now, change the filter in subscription 1
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", eventingtesting.OrderCreatedEventTypeNotClean)
	addCleanEventTypesToStatus(sub, cleaner)
	// Sync the subscription
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).ShouldNot(HaveOccurred())

	// get new cleaned subject from subscription 1
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(newSubject).To(Not(BeEmpty()))

	// check we have correct number of total subscriptions
	expectedTotalNatsSubs = 2 * defaultSubsConfig.MaxInFlightMessages // Because we have two subscriptions
	g.Expect(len(natsBackend.subscriptions)).To(Equal(expectedTotalNatsSubs))

	// check if the NATS subscription are NOT the same after sync for subscription 1
	// because the subscriptions should have being re-created for new subject
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
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
	// because the subscriptions should NOT have being re-created as
	// subscription 2 was not modified
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
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

// Test_isNatsSubAssociatedWithKymaSub tests the isNatsSubAssociatedWithKymaSub method
func Test_isNatsSubAssociatedWithKymaSub(t *testing.T) {
	g := NewWithT(t)

	// // ######  Setup test assets ######
	// create subscription 1 and its nats subscription
	cleanSubject1 := "subOne"
	sub1 := eventingtesting.NewSubscription(cleanSubject1, "foo", eventingtesting.WithNotCleanFilter())
	natsSub1Key := createKey(sub1, cleanSubject1, 0)
	natsSub1 := &nats.Subscription{
		Subject: cleanSubject1,
	}

	// create subscription 2 and its nats subscription
	cleanSubject2 := "subOneTwo"
	sub2 := eventingtesting.NewSubscription(cleanSubject2, "foo", eventingtesting.WithNotCleanFilter())
	natsSub2Key := createKey(sub2, cleanSubject2, 0)
	natsSub2 := &nats.Subscription{
		Subject: cleanSubject2,
	}

	// // ###### Test logic ######
	// Should return true because natsSub1 is associated with sub1
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub1Key, natsSub1, sub1)).To(Equal(true))
	// Should return true because natsSub2 is associated with sub2
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub2Key, natsSub2, sub2)).To(Equal(true))

	// Should return false because natsSub1 is NOT associated with sub2
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub1Key, natsSub1, sub2)).To(Equal(false))
	// Should return false because natsSub2 is NOT associated with sub1
	g.Expect(isNatsSubAssociatedWithKymaSub(natsSub2Key, natsSub2, sub1)).To(Equal(false))
}

func TestMultipleSubscriptionsToSameEvent(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 1
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)

	// Create 3 subscriptions having the same sink and the same event type
	var subs [3]*eventingv1alpha1.Subscription
	for i := 0; i < len(subs); i++ {
		subs[i] = eventingtesting.NewSubscription(fmt.Sprintf("sub-%d", i), "foo",
			eventingtesting.WithNotCleanFilter(),
			eventingtesting.WithSinkURL(subscriber.SinkURL),
			eventingtesting.WithStatusConfig(defaultSubsConfig),
		)
		addCleanEventTypesToStatus(subs[i], cleaner)
		err := natsBackend.SyncSubscription(subs[i])
		g.Expect(err).To(BeNil())
	}

	// Send only one event. It should be multiplexed to 3 by NATS, cause 3 subscriptions exist
	data := "sampledata"
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// Check for the 3 events that should be received by the subscriber
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	for i := 0; i < len(subs); i++ {
		g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
	}
	// Delete all 3 subscription
	for i := 0; i < len(subs); i++ {
		g.Expect(natsBackend.DeleteSubscription(subs[i])).Should(Succeed())
	}
	// Check if all subscriptions are deleted in NATS
	// Send an event again which should not be delivered to subscriber
	newData := "test-data"
	g.Expect(SendEventToNATS(natsBackend, newData)).Should(Succeed())
	// Check for the event that did not reach the subscriber
	// Store should never return newdata hence CheckEvent should fail to match newdata
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	g.Expect(subscriber.CheckEvent(notExpectedNewDataInStore)).ShouldNot(Succeed())
}

func TestSubscriptionWithDuplicateFilters(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithFilter("", eventingtesting.OrderCreatedEventType),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	idFunc := func(et string) (string, error) { return et, nil }
	addCleanEventTypesToStatus(sub, eventtype.CleanerFunc(idFunc))
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	data := "sampledata"
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
	// There should be no more!
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
}

func TestSubscriptionWithMaxInFlightChange(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	sub.Spec.Sink = fmt.Sprintf("http://127.0.0.1:%d/store", nextSinkPort.get())
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	filter := sub.Spec.Filter.Filters[0]
	subject, err := getCleanSubject(filter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// get internal key
	var key string
	var natsSub *nats.Subscription
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		key = createKey(sub, subject, i)
		natsSub = natsBackend.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
	}
	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))

	// check that no invalid subscriptions exist
	invalidNsn := natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeZero())

	addCleanEventTypesToStatus(sub, cleaner)
	sub.Status.Config = &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: 7}
	err = natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		key = createKey(sub, subject, i)
		natsSub = natsBackend.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
	}
	g.Expect(len(natsBackend.subscriptions)).To(Equal(sub.Status.Config.MaxInFlightMessages))
	// check that no invalid subscriptions exist
	invalidNsn = natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeZero())
}

func TestIsValidSubscription(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithOrderCreatedFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	filter := sub.Spec.Filter.Filters[0]
	subject, err := getCleanSubject(filter, cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// get internal key
	var key string
	var natsSub *nats.Subscription
	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
		key = createKey(sub, subject, i)
		g.Expect(key).To(Not(BeEmpty()))
		natsSub = natsBackend.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
	}
	// check the mapping of Kyma subscription and Nats subscription
	nsn := createKymaSubscriptionNamespacedName(key, natsSub)
	g.Expect(nsn.Namespace).To(BeIdenticalTo(sub.Namespace))
	g.Expect(nsn.Name).To(BeIdenticalTo(sub.Name))
	// the associated NATS subscription should be valid
	g.Expect(natsSub.IsValid()).To(BeTrue())
	// check that no invalid subscriptions exist
	invalidNsn := natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeZero())
	natsBackend.connection.Close()
	// the associated NATS subscription should not be valid anymore
	g.Expect(checkIsNotValid(natsSub, t)).Should(Succeed())

	natsServer.Shutdown()
	// the associated NATS subscription should not be valid anymore
	g.Expect(checkIsNotValid(natsSub, t)).Should(Succeed())
	// check that only one invalid subscription exist
	invalidNsn = natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(defaultSubsConfig.MaxInFlightMessages))
	// restart NATS server
	_, _ = startNATSServer()
	// check that only one invalid subscription still exist, the controller is not running...
	invalidNsn = natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(defaultSubsConfig.MaxInFlightMessages))
}

func TestSubscriptionUsingCESDK(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 1
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithOrderCreatedFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	subject := eventingtesting.CloudEventType
	g.Expect(SendBinaryCloudEventToNATS(natsBackend, subject, eventingtesting.CloudEventData)).Should(Succeed())
	g.Expect(subscriber.CheckEvent(eventingtesting.CloudEventData)).Should(Succeed())
	g.Expect(SendStructuredCloudEventToNATS(natsBackend, subject, eventingtesting.StructuredCloudEvent)).Should(Succeed())
	g.Expect(subscriber.CheckEvent("\"" + eventingtesting.EventData + "\"")).Should(Succeed())
	g.Expect(natsBackend.DeleteSubscription(sub)).Should(Succeed())
}

func TestRetryUsingCESDK(t *testing.T) {
	g := NewWithT(t)
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	maxRetries := 3
	defaultSubsConfig := env.DefaultSubscriptionConfig{
		MaxInFlightMessages:   1,
		DispatcherRetryPeriod: time.Second,
		DispatcherMaxRetries:  maxRetries,
	}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithOrderCreatedFilter(),
		eventingtesting.WithSinkURL(subscriber.InternalErrorURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	subject := eventingtesting.CloudEventType
	g.Expect(SendStructuredCloudEventToNATS(natsBackend, subject, eventingtesting.StructuredCloudEvent)).Should(Succeed())
	// Check that the retries are done and that the published data was correctly received
	g.Expect(subscriber.CheckRetries(maxRetries, "\""+eventingtesting.EventData+"\"")).Should(Succeed())
	g.Expect(natsBackend.DeleteSubscription(sub)).Should(Succeed())
}

func TestSubscription_NATSServerRestart(t *testing.T) {
	g := NewWithT(t)
	natsServer, natsPort := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	// The `reconnects` configs should be large enough to cover the NATS server restart period
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 10,
		ReconnectWait: 3 * time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 10}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, defaultLogger)
	g.Expect(natsBackend.Initialize(nil)).Should(Succeed())

	subscriber := eventingtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo",
		eventingtesting.WithNotCleanFilter(),
		eventingtesting.WithSinkURL(subscriber.SinkURL),
		eventingtesting.WithStatusConfig(defaultSubsConfig),
	)
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	addCleanEventTypesToStatus(sub, cleaner)
	err := natsBackend.SyncSubscription(sub)
	g.Expect(err).To(BeNil())

	ev1data := "sampledata"
	g.Expect(SendEventToNATS(natsBackend, ev1data)).Should(Succeed())
	expectedEv1Data := fmt.Sprintf("\"%s\"", ev1data)
	g.Expect(subscriber.CheckEvent(expectedEv1Data)).Should(Succeed())

	natsServer.Shutdown()
	g.Eventually(func() bool {
		return natsBackend.connection.IsConnected()
	}, 60*time.Second, 2*time.Second).Should(BeFalse())
	_ = eventingtesting.RunNatsServerOnPort(eventingtesting.WithPort(natsPort))
	g.Eventually(func() bool {
		return natsBackend.connection.IsConnected()
	}, 60*time.Second, 2*time.Second).Should(BeTrue())

	// After reconnect, event delivery should work again
	ev2data := "newsampledata"
	g.Expect(SendEventToNATS(natsBackend, ev2data)).Should(Succeed())
	expectedEv2Data := fmt.Sprintf("\"%s\"", ev2data)
	g.Expect(subscriber.CheckEvent(expectedEv2Data)).Should(Succeed())
}

func addCleanEventTypesToStatus(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner) {
	sub.Status.CleanEventTypes, _ = GetCleanSubjects(sub, cleaner)
}

func createEventTypeCleaner(eventTypePrefix, applicationName string, logger *logger.Logger) eventtype.Cleaner { //nolint:unparam
	application := applicationtest.NewApplication(applicationName, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), application)
	return eventtype.NewCleaner(eventTypePrefix, applicationLister, logger)
}

func getLogger(g *GomegaWithT, level kymalogger.Level) *logger.Logger { //nolint:unparam
	l, err := logger.New(string(kymalogger.JSON), string(level))
	g.Expect(err).To(BeNil())
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
