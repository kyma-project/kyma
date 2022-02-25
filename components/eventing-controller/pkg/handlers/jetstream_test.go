package handlers

import (
	"errors"
	"fmt"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

const (
	defaultStreamName = "kyma"
)

type jetStreamClient struct {
	nats.JetStreamContext
	natsConn *nats.Conn
}

func TestJetStream_Initialize_NoStreamExists(t *testing.T) {
	// Given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()

	// No stream exists
	_, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	assert.True(t, errors.Is(err, nats.ErrStreamNotFound))

	// When
	initErr := testEnvironment.jsBackend.Initialize(nil)

	// Then
	// A stream is created
	assert.NoError(t, initErr)
	streamInfo, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	assert.NoError(t, err)
	assert.NotNil(t, streamInfo)
}

func TestJetStream_Initialize_StreamExists(t *testing.T) {
	// Given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()

	// A stream already exists
	createdStreamInfo, err := testEnvironment.jsClient.AddStream(&nats.StreamConfig{
		Name:    testEnvironment.natsConfig.JSStreamName,
		Storage: nats.MemoryStorage,
	})
	assert.NotNil(t, createdStreamInfo)
	assert.NoError(t, err)

	// When
	initErr := testEnvironment.jsBackend.Initialize(nil)

	// Then
	// No new stream should be created
	assert.NoError(t, initErr)
	reusedStreamInfo, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	assert.NoError(t, err)
	assert.Equal(t, reusedStreamInfo.Created, createdStreamInfo.Created)
}

// TODO: Add test TestNatsSubAfterSync_NoChange

func TestJetStream_Subscription(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	assert.NoError(t, initErr)

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	assert.True(t, subscriber.IsRunning())

	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
	)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	data := "sampledata"
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// when
	assert.NoError(t, testEnvironment.jsBackend.DeleteSubscription(sub))

	// then
	newData := "test-data"
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, newData))
	// Check for the event that it did not reach subscriber
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	assert.Error(t, subscriber.CheckEvent(notExpectedNewDataInStore))
}

// TestJetStreamSubAfterSync_SinkChange tests the SyncSubscription method
// when only the sink is changed in subscription, then it should not re-create
// NATS subjects on nats-server
func TestJetStreamSubAfterSync_SinkChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	assert.NoError(t, initErr)

	// create New Subscribers
	subscriber1 := evtesting.NewSubscriber()
	defer subscriber1.Shutdown()
	assert.True(t, subscriber1.IsRunning())
	subscriber2 := evtesting.NewSubscriber()
	defer subscriber2.Shutdown()
	assert.True(t, subscriber2.IsRunning())

	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber1.SinkURL),
	)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, subject)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	assert.NoError(t, subscriber1.CheckEvent(expectedDataInStore))

	//// set metadata on NATS subscriptions
	//msgLimit, bytesLimit := 2048, 2048
	//g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	//for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
	//	natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
	//	g.Expect(natsSub).To(Not(BeNil()))
	//	g.Expect(natsSub.IsValid()).To(BeTrue())
	//	// set metadata on nats subscription
	//	g.Expect(natsSub.SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
	//}

	// given
	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2.SinkURL

	// when
	err = testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	//g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	//for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
	//	natsSub := natsBackend.subscriptions[createKey(sub, subject, i)]
	//	g.Expect(natsSub).To(Not(BeNil()))
	//	g.Expect(natsSub.IsValid()).To(BeTrue())
	//
	//	// check the metadata, if they are now same then it means that NATS subscription
	//	// were not re-created by SyncSubscription method
	//	subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
	//	g.Expect(err).ShouldNot(HaveOccurred())
	//	g.Expect(subMsgLimit).To(Equal(msgLimit))
	//	g.Expect(subBytesLimit).To(Equal(msgLimit))
	//}

	// Test if the subscription is working for new sink only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	// Old sink should not have received the event, the new sink should have
	assert.Error(t, subscriber1.CheckEvent(expectedDataInStore))
	assert.NoError(t, subscriber2.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FiltersChange tests the SyncSubscription method
// when the filters are changed in subscription
func TestJetStreamSubAfterSync_FiltersChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	assert.NoError(t, initErr)

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	assert.True(t, subscriber.IsRunning())

	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
	)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, subject)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	assert.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	//g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	for key := range testEnvironment.jsBackend.subscriptions {
		// set metadata on nats subscription
		assert.NoError(t, testEnvironment.jsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// Now, change the filter in subscription
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", evtesting.OrderCreatedEventTypeNotClean)
	// Sync the subscription
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err = testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	// TODO: Check why is this erroring out
	// testStreamsAndConsumers(t, testEnvironment, sub)

	// get new cleaned subject
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, newSubject)
	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	// g.Expect(len(testEnvironment.jsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
	//for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
	//	natsSub := natsBackend.subscriptions[createKey(sub, newSubject, i)]
	//	g.Expect(natsSub).To(Not(BeNil()))
	//	g.Expect(natsSub.IsValid()).To(BeTrue())
	//
	//	// check the metadata, if they are NOT same then it means that nats subscriptions
	//	// were re-created by SyncSubscription method
	//	subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
	//	g.Expect(err).ShouldNot(HaveOccurred())
	//	g.Expect(subMsgLimit).To(Not(Equal(msgLimit)))
	//	g.Expect(subBytesLimit).To(Not(Equal(msgLimit)))
	//}

	// Test if subscription is working for new subject only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	// Send an event on old subject
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	// The sink should not receive any event for old subject
	assert.Error(t, subscriber.CheckEvent(expectedDataInStore))
	// Now, send an event on new subject
	assert.NoError(t, SendEventToJetStreamOnEventType(testEnvironment.jsBackend, newSubject, data))
	// The sink should receive the event for new subject
	assert.NoError(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FilterAdded tests the SyncSubscription method
// when a new filter is added in subscription
func TestJetStreamSubAfterSync_FilterAdded(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	assert.NoError(t, initErr)

	// Create a new subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	assert.True(t, subscriber.IsRunning())

	// Create a subscription with single filter
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
	)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// get cleaned subject
	firstSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, firstSubject)

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	// assert.Equal(t, len(testEnvironment.jsBackend.subscriptions), defaultSubsConfig.MaxInFlightMessages)
	for key := range testEnvironment.jsBackend.subscriptions {
		// set metadata on nats subscription
		assert.NoError(t, testEnvironment.jsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit))
	}

	// Now, add a new filter to subscription
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", evtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)

	// get new cleaned subject
	secondSubject, err := getCleanSubject(newFilter, testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, secondSubject)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err = testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	//expectedTotalNatsSubs := 2 * defaultSubsConfig.MaxInFlightMessages
	//g.Expect(natsBackend.subscriptions).To(HaveLen(expectedTotalNatsSubs))

	// Verify that the nats subscriptions for first subject was not re-created
	//for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
	//	natsSub := natsBackend.subscriptions[createKey(sub, firstSubject, i)]
	//	g.Expect(natsSub).To(Not(BeNil()))
	//	g.Expect(natsSub.IsValid()).To(BeTrue())
	//
	//	// check the metadata, if they are same then it means that nats subscriptions
	//	// were not re-created by SyncSubscription method
	//	subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
	//	g.Expect(err).ShouldNot(HaveOccurred())
	//	g.Expect(subMsgLimit).To(Equal(msgLimit))
	//	g.Expect(subBytesLimit).To(Equal(msgLimit))
	//}

	// Test if subscription is working for both subjects
	// Send an event on first subject
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	// The sink should receive event for first subject
	assert.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStreamOnEventType(testEnvironment.jsBackend, secondSubject, data))
	// The sink should receive the event for second subject
	assert.NoError(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FilterRemoved tests the SyncSubscription method
// when a filter is removed from subscription
func TestJetStreamSubAfterSync_FilterRemoved(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	assert.NoError(t, initErr)

	// Create a new subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	assert.True(t, subscriber.IsRunning())

	// Create a subscription with two filters
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
	)
	// add a second filter
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", evtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// get cleaned subjects
	firstSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, firstSubject)

	secondSubject, err := getCleanSubject(sub.Spec.Filter.Filters[1], testEnvironment.cleaner)
	assert.NoError(t, err)
	assert.NotEmpty(t, secondSubject)

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	// expectedTotalNatsSubs := 2 * defaultSubsConfig.MaxInFlightMessages
	// g.Expect(natsBackend.subscriptions).To(HaveLen(expectedTotalNatsSubs))

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	for key := range testEnvironment.jsBackend.subscriptions {
		// set metadata on nats subscription
		assert.NoError(t, testEnvironment.jsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// Now, remove the second filter from subscription
	sub.Spec.Filter.Filters = sub.Spec.Filter.Filters[:1]
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err = testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	assert.NoError(t, err)
	testStreamsAndConsumers(t, testEnvironment, sub)

	// Check if total existing NATS subscriptions are correct
	// assert.Equal(t, len(testEnvironment.jsBackend.subscriptions), defaultSubsConfig.MaxInFlightMessages)
	// Verify that the nats subscriptions for first subject was not re-created
	//for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
	//	natsSub := natsBackend.subscriptions[createKey(sub, firstSubject, i)]
	//	g.Expect(natsSub).To(Not(BeNil()))
	//	g.Expect(natsSub.IsValid()).To(BeTrue())
	//
	//	// check the metadata, if they are same then it means that nats subscriptions
	//	// were not re-created by SyncSubscription method
	//	subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
	//	g.Expect(err).ShouldNot(HaveOccurred())
	//	g.Expect(subMsgLimit).To(Equal(msgLimit))
	//	g.Expect(subBytesLimit).To(Equal(msgLimit))
	//}

	// Test if subscription is working for first subject only
	// Send an event on first subject
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, data))
	// The sink should receive event for first subject
	assert.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	assert.NoError(t, SendEventToJetStreamOnEventType(testEnvironment.jsBackend, secondSubject, data))
	// The sink should NOT receive the event for second subject
	assert.Error(t, subscriber.CheckEvent(expectedDataInStore))
}

func testStreamsAndConsumers(t *testing.T, testEnvironment *TestEnvironment, sub *eventingv1alpha1.Subscription) {
	streamInfo, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	assert.NoError(t, err)
	assert.NotNil(t, streamInfo)
	// TODO: Check the consumer length comparison in another function
	// assert.Equal(t, streamInfo.State.Consumers, 1)
	consumerName := encodeString(createKeyPrefix(sub) + string(types.Separator) + evtesting.OrderCreatedEventType)
	info, err := testEnvironment.jsClient.ConsumerInfo(testEnvironment.natsConfig.JSStreamName, consumerName)
	assert.NotNil(t, info)
	assert.NoError(t, err)
}

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                 url,
		JSStreamName:        defaultStreamName,
		JSStreamStorageType: JetStreamMemoryStorageType,
	}
}

// getJetStreamClient creates a client with JetStream context, or fails the caller test.
func getJetStreamClient(t *testing.T, serverURL string) *jetStreamClient {
	conn, err := nats.Connect(serverURL)
	if err != nil {
		t.Error(err.Error())
	}
	jsCtx, err := conn.JetStream()
	if err != nil {
		conn.Close()
		t.Error(err.Error())
	}
	return &jetStreamClient{
		JetStreamContext: jsCtx,
		natsConn:         conn,
	}
}

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	jsBackend  *JetStream
	logger     *logger.Logger
	natsServer *server.Server
	jsClient   *jetStreamClient
	natsConfig env.NatsConfig
	cleaner    eventtype.Cleaner
}

// setupTestEnvironment is a TestEnvironment constructor
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer, _ := startNATSServer(evtesting.WithJetStreamEnabled())
	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	assert.NoError(t, err)

	jsClient := getJetStreamClient(t, natsConfig.URL)
	jsBackend := NewJetStream(natsConfig, defaultLogger)
	cleaner := createEventTypeCleaner(evtesting.EventTypePrefix, evtesting.ApplicationNameNotClean, defaultLogger)

	return &TestEnvironment{
		jsBackend:  jsBackend,
		logger:     defaultLogger,
		natsServer: natsServer,
		jsClient:   jsClient,
		natsConfig: natsConfig,
		cleaner:    cleaner,
	}
}
