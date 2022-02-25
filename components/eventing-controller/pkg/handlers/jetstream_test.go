package handlers

import (
	"errors"
	"fmt"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	g := NewWithT(t)
	// Given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()

	// No stream exists
	_, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	g.Expect(errors.Is(err, nats.ErrStreamNotFound)).To(BeTrue())

	// When
	initErr := testEnvironment.jsBackend.Initialize(nil)

	// Then
	// A stream is created
	g.Expect(initErr).To(BeNil())
	g.Expect(testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)).ShouldNot(BeNil())
}

func TestJetStream_Initialize_StreamExists(t *testing.T) {
	g := NewWithT(t)
	// Given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()

	// A stream already exists
	createdStreamInfo, err := testEnvironment.jsClient.AddStream(&nats.StreamConfig{
		Name:    testEnvironment.natsConfig.JSStreamName,
		Storage: nats.MemoryStorage,
	})
	g.Expect(createdStreamInfo).ToNot(BeNil())
	g.Expect(err).To(BeNil())

	// When
	initErr := testEnvironment.jsBackend.Initialize(nil)

	// Then
	// No new stream should be created
	g.Expect(initErr).To(BeNil())
	reusedStreamInfo, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	g.Expect(err).To(BeNil())
	g.Expect(reusedStreamInfo.Created).To(Equal(createdStreamInfo.Created))
}

func TestJetStream_Subscription(t *testing.T) {
	g := NewWithT(t)
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	g.Expect(initErr).To(BeNil())

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	g.Expect(subscriber.IsRunning()).To(BeTrue())

	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
	)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	g.Expect(err).To(BeNil())
	streamInfo, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	g.Expect(streamInfo).ShouldNot(BeNil())
	g.Expect(streamInfo.State.Consumers).To(Equal(1))
	consumerName := generateHashForString(createKeyPrefix(sub) + string(types.Separator) + evtesting.OrderCreatedEventType)
	g.Expect(testEnvironment.jsClient.ConsumerInfo(testEnvironment.natsConfig.JSStreamName, consumerName)).ShouldNot(BeNil())

	data := "sampledata"
	g.Expect(SendEventToJetStream(testEnvironment.jsBackend, data)).Should(Succeed())
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())

	// TODO: Uncomment after DeleteSubscription implementation
	//g.Expect(testEnvironment.jsBackend.DeleteSubscription(sub)).Should(Succeed())
	//
	//newData := "test-data"
	//g.Expect(SendEventToJetStream(testEnvironment.jsBackend, newData)).Should(Succeed())
	//// Check for the event that it did not reach subscriber
	//notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	//g.Expect(subscriber.CheckEvent(notExpectedNewDataInStore)).ShouldNot(Succeed())
}

// TestNatsSubAfterSync_SinkChange tests the SyncSubscription method
// when only the sink is changed in subscription, then it should not re-create
// NATS subjects on nats-server
func TestJetStreamSubAfterSync_SinkChange(t *testing.T) {
	g := NewWithT(t)
	// given
	testEnvironment := setupTestEnvironment(t)
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := testEnvironment.jsBackend.Initialize(nil)
	g.Expect(initErr).To(BeNil())

	// create New Subscribers
	subscriber1 := evtesting.NewSubscriber()
	defer subscriber1.Shutdown()
	g.Expect(subscriber1.IsRunning()).To(BeTrue())
	subscriber2 := evtesting.NewSubscriber()
	defer subscriber2.Shutdown()
	g.Expect(subscriber2.IsRunning()).To(BeTrue())

	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber1.SinkURL),
	)
	addCleanEventTypesToStatus(sub, testEnvironment.cleaner)

	// when
	err := testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	g.Expect(err).To(BeNil())
	streamInfo, err := testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	g.Expect(streamInfo).ShouldNot(BeNil())
	g.Expect(streamInfo.State.Consumers).To(Equal(1))
	consumerName := generateHashForString(createKeyPrefix(sub) + string(types.Separator) + evtesting.OrderCreatedEventType)
	g.Expect(testEnvironment.jsClient.ConsumerInfo(testEnvironment.natsConfig.JSStreamName, consumerName)).ShouldNot(BeNil())

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(subject).To(Not(BeEmpty()))

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToJetStream(testEnvironment.jsBackend, data)).Should(Succeed())
	g.Expect(subscriber1.CheckEvent(expectedDataInStore)).Should(Succeed())

	// given
	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2.SinkURL

	// when
	err = testEnvironment.jsBackend.SyncSubscription(sub)

	// then
	g.Expect(err).To(BeNil())
	streamInfo, err = testEnvironment.jsClient.StreamInfo(testEnvironment.natsConfig.JSStreamName)
	g.Expect(streamInfo).ShouldNot(BeNil())
	g.Expect(streamInfo.State.Consumers).To(Equal(1))
	consumerName = generateHashForString(createKeyPrefix(sub) + string(types.Separator) + evtesting.OrderCreatedEventType)
	g.Expect(testEnvironment.jsClient.ConsumerInfo(testEnvironment.natsConfig.JSStreamName, consumerName)).ShouldNot(BeNil())

	// Test if the subscription is working for new sink only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	g.Expect(SendEventToJetStream(testEnvironment.jsBackend, data)).Should(Succeed())
	// Old sink should not have received the event, the new sink should have
	g.Expect(subscriber1.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
	g.Expect(subscriber2.CheckEvent(expectedDataInStore)).Should(Succeed())
}

//// TestNatsSubAfterSync_FiltersChange tests the SyncSubscription method
//// when the filters are changed in subscription
//func TestNatsSubAfterSync_FiltersChange(t *testing.T) {
//	g := NewWithT(t)
//	natsServer, _ := startNATSServer()
//	defer natsServer.Shutdown()
//	defaultLogger := getLogger(g, kymalogger.INFO)
//	natsConfig := env.NatsConfig{
//		URL:           natsServer.ClientURL(),
//		MaxReconnects: 2,
//		ReconnectWait: time.Second,
//	}
//	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 5}
//	natsBackend := NewNats(natsConfig, defaultSubsConfig, nil, defaultLogger)
//	g.Expect(natsBackend.Initialize(env.Config{})).Should(Succeed())
//
//	subscriber := eventingtesting.NewSubscriber()
//	defer subscriber.Shutdown()
//	g.Expect(subscriber.IsRunning()).To(BeTrue())
//
//	sub := eventingtesting.NewSubscription("sub", "foo",
//		eventingtesting.WithNotCleanFilter(),
//		eventingtesting.WithSinkURL(subscriber.SinkURL),
//		eventingtesting.WithStatusConfig(defaultSubsConfig),
//	)
//	cleaner := createEventTypeCleaner(eventingtesting.EventTypePrefix, eventingtesting.ApplicationNameNotClean, defaultLogger)
//	addCleanEventTypesToStatus(sub, cleaner)
//	_, err := natsBackend.SyncSubscription(sub, cleaner)
//	g.Expect(err).To(BeNil())
//
//	// get cleaned subject
//	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
//	g.Expect(err).ShouldNot(HaveOccurred())
//	g.Expect(subject).To(Not(BeEmpty()))
//
//	// test if subscription is working properly by sending an event
//	// and checking if it is received by the subscriber
//	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
//	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
//	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
//	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
//
//	// set metadata on NATS subscriptions
//	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
//	msgLimit, bytesLimit := 2048, 2048
//	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
//	for key := range natsBackend.subscriptions {
//		// set metadata on nats subscription
//		g.Expect(natsBackend.subscriptions[key].SetPendingLimits(msgLimit, bytesLimit)).Should(Succeed())
//	}
//
//	// Now, change the filter in subscription
//	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", eventingtesting.OrderCreatedEventTypeNotClean)
//	// Sync the subscription
//	addCleanEventTypesToStatus(sub, cleaner)
//	_, err = natsBackend.SyncSubscription(sub, cleaner)
//	g.Expect(err).To(BeNil())
//
//	// get new cleaned subject
//	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], cleaner)
//	g.Expect(err).ShouldNot(HaveOccurred())
//	g.Expect(newSubject).To(Not(BeEmpty()))
//
//	// check if the NATS subscription are NOT the same after sync
//	// because the subscriptions should have being re-created for new subject
//	g.Expect(len(natsBackend.subscriptions)).To(Equal(defaultSubsConfig.MaxInFlightMessages))
//	for i := 0; i < defaultSubsConfig.MaxInFlightMessages; i++ {
//		natsSub := natsBackend.subscriptions[createKey(sub, newSubject, i)]
//		g.Expect(natsSub).To(Not(BeNil()))
//		g.Expect(natsSub.IsValid()).To(BeTrue())
//
//		// check the metadata, if they are NOT same then it means that nats subscriptions
//		// were re-created by SyncSubscription method
//		subMsgLimit, subBytesLimit, err := natsSub.PendingLimits()
//		g.Expect(err).ShouldNot(HaveOccurred())
//		g.Expect(subMsgLimit).To(Not(Equal(msgLimit)))
//		g.Expect(subBytesLimit).To(Not(Equal(msgLimit)))
//	}
//
//	// Test if subscription is working for new subject only
//	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
//	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
//	// Send an event on old subject
//	g.Expect(SendEventToNATS(natsBackend, data)).Should(Succeed())
//	// The sink should not receive any event for old subject
//	g.Expect(subscriber.CheckEvent(expectedDataInStore)).ShouldNot(Succeed())
//	// Now, send an event on new subject
//	g.Expect(SendEventToNATSOnEventType(natsBackend, newSubject, data)).Should(Succeed())
//	// The sink should receive the event for new subject
//	g.Expect(subscriber.CheckEvent(expectedDataInStore)).Should(Succeed())
//}

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                 url,
		JSStreamName:        defaultStreamName,
		JSStreamStorageType: JetStreamMemoryStorageType,
	}
}

// getJetStreamClient creates a client with JetStream context, or fails the caller test.
func getJetStreamClient(serverURL string) *jetStreamClient {
	conn, err := nats.Connect(serverURL)
	if err != nil {
		ginkgo.Fail(err.Error())
	}
	jsCtx, err := conn.JetStream()
	if err != nil {
		conn.Close()
		ginkgo.Fail(err.Error())
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
	g := NewWithT(t)

	natsServer, _ := startNATSServer(evtesting.WithJetStreamEnabled())
	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger := getLogger(g, kymalogger.INFO)

	jsClient := getJetStreamClient(natsConfig.URL)

	jsBackend := NewJetStream(natsConfig, defaultLogger)
	cleaner := createEventTypeCleaner(evtesting.EventTypePrefix, evtesting.ApplicationNameNotClean, defaultLogger)

	return &TestEnvironment{
		jsBackend:  jsBackend,
		logger:     defaultLogger,
		jsClient:   jsClient,
		natsConfig: natsConfig,
		cleaner:    cleaner,
	}
}
