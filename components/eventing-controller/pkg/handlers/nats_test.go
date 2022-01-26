package handlers

import (
	"context"
	"errors"
	"fmt"
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
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultMaxInflight := 9
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriber.GetSinkURL()
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	_, err := natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())
	g.Expect(sub.Status.Config).NotTo(BeNil()) // It should apply the defaults
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(defaultMaxInflight))

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
// when there is no change in subscription then it should not re-create
// NATS subjects on nats-server
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
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriber.GetSinkURL()
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	_, err := natsBackend.SyncSubscription(sub, cleaner)
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
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
		// set metadata on nats subscription
		g.Expect(natsSub.SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}
	// Now, sync the subscription
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())

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
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber1, _ := startSubscriber()
	defer subscriber1.Shutdown()
	g.Expect(subscriber1.IsRunning()).To(BeTrue())
	subscriber2, _ := startSubscriber()
	defer subscriber2.Shutdown()
	g.Expect(subscriber2.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriber1.GetSinkURL()
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	_, err := natsBackend.SyncSubscription(sub, cleaner)
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
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
		// set metadata on nats subscription
		g.Expect(natsSub.SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}
	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2.GetSinkURL()
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())
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
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// Old sink should not have received the event, the new sink should have
	g.Expect(subscriber1.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
	g.Expect(subscriber2.CheckEvent(expectedDataInStore)).Should(Succeed())
}

// TestNatsSubAfterSync_SinkChange tests the SyncSubscription method
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
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriber.GetSinkURL()
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
	_, err := natsBackend.SyncSubscription(sub, cleaner)
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
	for i := 0; i < sub.Status.Config.MaxInFlightMessages; i++ {
		natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
		g.Expect(natsSub).To(Not(BeNil()))
		g.Expect(natsSub.IsValid()).To(BeTrue())
		// set metadata on nats subscription
		g.Expect(natsSub.SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	}

	// Now, change the filter in subscription and Sync the subscription
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", eventingtesting.OrderCreatedEventTypeNotClean)
	_, err = natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())

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
	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
	// The sink should not receive any event for old subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
	// Now, send an event on new subject
	g.Expect(SendEventToNATSOnEventType(natsBackend, newSubject, data)).Should(Succeed())
	// The sink should receive the event for new subject
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
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
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)

	// Create 3 subscriptions having the same sink and the same event type
	var subs [3]*eventingv1alpha1.Subscription
	for i := 0; i < len(subs); i++ {
		subs[i] = eventingtesting.NewSubscription(fmt.Sprintf("sub-%d", i), "foo", eventingtesting.WithNotCleanEventTypeFilter)
		subs[i].Spec.Sink = subscriber.GetSinkURL()
		_, err := natsBackend.SyncSubscription(subs[i], cleaner)
		g.Expect(err).To(BeNil())
		g.Expect(subs[i].Status.Config).NotTo(BeNil()) // It should apply the defaults
		g.Expect(subs[i].Status.Config.MaxInFlightMessages).To(Equal(defaultMaxInflight))
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
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

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
	sub.Spec.Sink = subscriber.GetSinkURL()
	idFunc := func(et string) (string, error) { return et, nil }
	_, err := natsBackend.SyncSubscription(sub, eventtype.CleanerFunc(idFunc))
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
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = fmt.Sprintf("http://127.0.0.1:%d/store", nextPort.get())
	_, err := natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())

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
	g.Expect(err).To(BeNil())

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
	natsServer, _ := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)

	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 2,
		ReconnectWait: time.Second,
	}
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithEventTypeFilter)
	sub.Spec.Sink = subscriber.GetSinkURL()
	_, err := natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())

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
		natsSub = natsBackend.subscriptions[key]
		g.Expect(natsSub).To(Not(BeNil()))
	}
	// check the mapping of Kyma subscription and Nats subscription
	nsn := createKymaSubscriptionNamespacedName(key, natsSub)
	g.Expect(nsn.Namespace).To(BeIdenticalTo(sub.Namespace))
	g.Expect(nsn.Name).To(BeIdenticalTo(sub.Name))
	// the associated Nats subscription should be valid
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
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(sub.Status.Config.MaxInFlightMessages))
	// restart NATS server
	_, _ = startNATSServer()
	// check that only one invalid subscription still exist, the controller is not running...
	invalidNsn = natsBackend.GetInvalidSubscriptions()
	g.Expect(len(*invalidNsn)).To(BeIdenticalTo(sub.Status.Config.MaxInFlightMessages))
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
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithEventTypeFilter)
	sub.Spec.Sink = subscriber.GetSinkURL()
	_, err := natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())
	g.Expect(sub.Status.Config).NotTo(BeNil()) // It should apply the defaults
	g.Expect(sub.Status.Config.MaxInFlightMessages).To(Equal(defaultMaxInflight))

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
	defaultSubscriptionConfig := env.DefaultSubscriptionConfig{
		MaxInFlightMessages:   1,
		DispatcherRetryPeriod: time.Second,
		DispatcherMaxRetries:  maxRetries,
	}
	natsBackend := NewNats(natsConfig, defaultSubscriptionConfig, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, subscriberPort := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithEventTypeFilter)
	subscriberServerErrorURL := fmt.Sprintf("http://127.0.0.1:%d/return500", subscriberPort)
	sub.Spec.Sink = subscriberServerErrorURL
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	_, err := natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())
	g.Expect(sub.Status.Config).NotTo(BeNil()) // It should apply the defaults

	subject := eventingtesting.CloudEventType
	g.Expect(SendStructuredCloudEventToNATS(natsBackend, subject, eventingtesting.StructuredCloudEvent)).Should(Succeed())
	// Check that the retries are done and that the sent data was correctly received
	g.Expect(subscriber.CheckRetries(maxRetries, "\""+eventingtesting.EventData+"\"")).Should(Succeed())
	g.Expect(natsBackend.DeleteSubscription(sub)).Should(Succeed())
}

func TestSubscription_NATSServerRestart(t *testing.T) {
	g := NewWithT(t)
	natsServer, natsPort := startNATSServer()
	defer natsServer.Shutdown()
	defaultLogger := getLogger(g, kymalogger.INFO)
	// The reconnect configs should be large enough to cover the NATS server restart period
	natsConfig := env.NatsConfig{
		URL:           natsServer.ClientURL(),
		MaxReconnects: 10,
		ReconnectWait: 3 * time.Second,
	}
	natsBackend := NewNats(natsConfig, env.DefaultSubscriptionConfig{MaxInFlightMessages: 10}, nil, defaultLogger)
	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())

	subscriber, _ := startSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	// Create a subscription
	sub := eventingtesting.NewSubscription("sub", "foo", eventingtesting.WithNotCleanEventTypeFilter)
	sub.Spec.Sink = subscriber.GetSinkURL()
	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationName, defaultLogger)
	_, err := natsBackend.SyncSubscription(sub, cleaner)
	g.Expect(err).To(BeNil())

	ev1data := "sampledata"
	g.Expect(SendEventToNATS(natsBackend, ev1data)).Should(Succeed())
	expectedEv1Data := fmt.Sprintf("\"%s\"", ev1data)
	g.Expect(subscriber.CheckEvent(expectedEv1Data)).Should(Succeed())

	natsServer.Shutdown()
	g.Eventually(func() bool {
		return natsBackend.connection.IsConnected()
	}, 60*time.Second, 2*time.Second).Should(BeFalse())
	_ = eventingtesting.RunNatsServerOnPort(natsPort)
	g.Eventually(func() bool {
		return natsBackend.connection.IsConnected()
	}, 60*time.Second, 2*time.Second).Should(BeTrue())

	// After reconnect, event delivery should work again
	ev2data := "newsampledata"
	g.Expect(SendEventToNATS(natsBackend, ev2data)).Should(Succeed())
	expectedEv2Data := fmt.Sprintf("\"%s\"", ev2data)
	g.Expect(subscriber.CheckEvent(expectedEv2Data)).Should(Succeed())
}

func startNATSServer() (*server.Server, int) {
	natsPort := nextPort.get()
	natsServer := eventingtesting.RunNatsServerOnPort(natsPort)
	return natsServer, natsPort
}

func startSubscriber() (*eventingtesting.Subscriber, int) {
	subscriberPort := nextPort.get()
	subscriber := eventingtesting.NewSubscriber(subscriberPort)
	subscriber.Start()
	return subscriber, subscriberPort
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
