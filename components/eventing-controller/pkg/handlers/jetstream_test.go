package handlers

import (
	"errors"
	"fmt"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

const (
	defaultStreamName = "kyma"
)

type jetStreamClient struct {
	nats.JetStreamContext
	natsConn *nats.Conn
}

// TestJetStreamInitialize_NoStreamExists tests if a stream is
// created when no stream exists in JetStream.
func TestJetStreamInitialize_NoStreamExists(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	natsConfig, jsClient := testEnvironment.natsConfig, testEnvironment.jsClient
	defer testEnvironment.natsServer.Shutdown()
	defer jsClient.natsConn.Close()

	// No stream exists
	_, err := jsClient.StreamInfo(natsConfig.JSStreamName)
	require.True(t, errors.Is(err, nats.ErrStreamNotFound))

	// when
	initErr := testEnvironment.jsBackend.Initialize(nil)

	// then
	// A stream is created
	require.NoError(t, initErr)
	streamInfo, err := jsClient.StreamInfo(natsConfig.JSStreamName)
	require.NoError(t, err)
	require.NotNil(t, streamInfo)
}

// TestJetStreamInitialize_StreamExists tests if a stream is
// reused and not created when a stream exists in JetStream.
func TestJetStreamInitialize_StreamExists(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	natsConfig, jsClient := testEnvironment.natsConfig, testEnvironment.jsClient
	defer testEnvironment.natsServer.Shutdown()
	defer jsClient.natsConn.Close()

	// A stream already exists
	createdStreamInfo, err := jsClient.AddStream(&nats.StreamConfig{
		Name:    natsConfig.JSStreamName,
		Storage: nats.MemoryStorage,
	})
	require.NotNil(t, createdStreamInfo)
	require.NoError(t, err)

	// when
	initErr := testEnvironment.jsBackend.Initialize(nil)

	// then
	// No new stream should be created
	require.NoError(t, initErr)
	reusedStreamInfo, err := jsClient.StreamInfo(natsConfig.JSStreamName)
	require.NoError(t, err)
	require.Equal(t, reusedStreamInfo.Created, createdStreamInfo.Created)
}

// TestJetStream_SubscriptionDeletion tests the creation and deletion
// of a JetStream subscription on the NATS server.
func TestJetStream_SubscriptionDeletion(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	data := "sampledata"
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// when
	require.NoError(t, jsBackend.DeleteSubscription(sub))

	// then
	newData := "test-data"
	require.NoError(t, SendEventToJetStream(jsBackend, newData))
	// Check for the event that it did not reach subscriber
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	require.Error(t, subscriber.CheckEvent(notExpectedNewDataInStore))
}

// TestJetStreamSubAfterSync_NoChange tests the SyncSubscription method
// when there is no change in the subscription then the method should
// not re-create NATS subjects on nats-server.
func TestJetStreamSubAfterSync_NoChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscribers
	subscriber1 := evtesting.NewSubscriber()
	defer subscriber1.Shutdown()
	require.True(t, subscriber1.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber1.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	require.NoError(t, subscriber1.CheckEvent(expectedDataInStore))

	// set metadata on NATS subscriptions
	msgLimit, bytesLimit := 2048, 2048
	require.Len(t, jsBackend.subscriptions, 1)
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// no change in subscription

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject := jsBackend.GetJsSubjectToSubscribe(subject)
	jsSubKey := jsBackend.generateJsSubKey(jsSubject, sub)
	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.Equal(t, subMsgLimit, msgLimit)
	require.Equal(t, subBytesLimit, bytesLimit)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	require.NoError(t, subscriber1.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_SinkChange tests the SyncSubscription method
// when only the sink is changed in subscription, then it should not re-create
// NATS subjects on nats-server.
func TestJetStreamSubAfterSync_SinkChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscribers
	subscriber1 := evtesting.NewSubscriber()
	defer subscriber1.Shutdown()
	require.True(t, subscriber1.IsRunning())
	subscriber2 := evtesting.NewSubscriber()
	defer subscriber2.Shutdown()
	require.True(t, subscriber2.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber1.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	require.NoError(t, subscriber1.CheckEvent(expectedDataInStore))

	// set metadata on NATS subscriptions
	msgLimit, bytesLimit := 2048, 2048
	require.Len(t, jsBackend.subscriptions, 1)
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// NATS subscription should not be re-created in sync when sink is changed.
	// change the sink
	sub.Spec.Sink = subscriber2.SinkURL

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// check if the NATS subscription are the same (have same metadata)
	// by comparing the metadata of nats subscription
	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject := jsBackend.GetJsSubjectToSubscribe(subject)
	jsSubKey := jsBackend.generateJsSubKey(jsSubject, sub)
	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.Equal(t, subMsgLimit, msgLimit)
	require.Equal(t, subBytesLimit, bytesLimit)

	// Test if the subscription is working for new sink only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	// Old sink should not have received the event, the new sink should have
	require.Error(t, subscriber1.CheckEvent(expectedDataInStore))
	require.NoError(t, subscriber2.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FiltersChange tests the SyncSubscription method
// when the filters are changed in subscription.
func TestJetStreamSubAfterSync_FiltersChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	// test if subscription is working properly by sending an event
	// and checking if it is received by the subscriber
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	require.Len(t, jsBackend.subscriptions, 1)
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// Now, change the filter in subscription
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", evtesting.OrderCreatedEventTypeNotClean)
	// Sync the subscription status
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject := jsBackend.GetJsSubjectToSubscribe(newSubject)
	jsSubKey := jsBackend.generateJsSubKey(jsSubject, sub)

	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are NOT same then it means that NATS subscription
	// were re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.NotEqual(t, subMsgLimit, msgLimit)
	require.NotEqual(t, subBytesLimit, bytesLimit)

	// Test if subscription is working for new subject only
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	// Send an event on old subject
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	// The sink should not receive any event for old subject
	require.Error(t, subscriber.CheckEvent(expectedDataInStore))
	// Now, send an event on new subject
	require.NoError(t, SendEventToJetStreamOnEventType(jsBackend, newSubject, data))
	// The sink should receive the event for new subject
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FilterAdded tests the SyncSubscription method
// when a new filter is added in subscription.
func TestJetStreamSubAfterSync_FilterAdded(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// Create a new subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	// Create a subscription with single filter
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	firstSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, firstSubject)

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	require.Len(t, jsBackend.subscriptions, 1)
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// Now, add a new filter to subscription
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", evtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)

	// get new cleaned subject
	secondSubject, err := getCleanSubject(newFilter, testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, secondSubject)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	require.Len(t, jsBackend.subscriptions, 2)
	// Verify that the nats subscriptions for first subject was not re-created
	jsSubject := jsBackend.GetJsSubjectToSubscribe(firstSubject)
	jsSubKey := jsBackend.generateJsSubKey(jsSubject, sub)

	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.Equal(t, subMsgLimit, msgLimit)
	require.Equal(t, subBytesLimit, bytesLimit)

	// Test if subscription is working for both subjects
	// Send an event on first subject
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	// The sink should receive event for first subject
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStreamOnEventType(jsBackend, secondSubject, data))
	// The sink should receive the event for second subject
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FilterRemoved tests the SyncSubscription method
// when a filter is removed from subscription.
func TestJetStreamSubAfterSync_FilterRemoved(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// Create a new subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	// Create a subscription with two filters
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	// add a second filter
	newFilter := sub.Spec.Filter.Filters[0].DeepCopy()
	newFilter.EventType.Value = fmt.Sprintf("%snew1", evtesting.OrderCreatedEventTypeNotClean)
	sub.Spec.Filter.Filters = append(sub.Spec.Filter.Filters, newFilter)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subjects
	firstSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, firstSubject)

	secondSubject, err := getCleanSubject(sub.Spec.Filter.Filters[1], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, secondSubject)

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	require.Len(t, jsBackend.subscriptions, 2)

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// given
	// Now, remove the second filter from subscription
	sub.Spec.Filter.Filters = sub.Spec.Filter.Filters[:1]
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	require.Len(t, jsBackend.subscriptions, 1)
	// Verify that the nats subscriptions for first subject was not re-created
	jsSubject := jsBackend.GetJsSubjectToSubscribe(firstSubject)
	jsSubKey := jsBackend.generateJsSubKey(jsSubject, sub)

	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.Equal(t, subMsgLimit, msgLimit)
	require.Equal(t, subBytesLimit, bytesLimit)

	// Test if subscription is working for first subject only
	// Send an event on first subject
	data := fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	// The sink should receive event for first subject
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))

	// Now, send an event on second subject
	data = fmt.Sprintf("data-%s", time.Now().Format(time.RFC850))
	expectedDataInStore = fmt.Sprintf("\"%s\"", data)
	require.NoError(t, SendEventToJetStreamOnEventType(jsBackend, secondSubject, data))
	// The sink should NOT receive the event for second subject
	require.Error(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_MultipleSubs tests the SyncSubscription method
// when there are two subscriptions and the filter is changed in one subscription
// it should not affect the NATS subscriptions of other Kyma subscriptions.
func TestJetStreamSubAfterSync_MultipleSubs(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// Create a new subscriber
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	defaultMaxInflight := 9
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}

	// Create two subscriptions with single filter
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// given
	sub2 := evtesting.NewSubscription("sub2", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub2, testEnvironment.cleaner, jsBackend)

	// when
	err = jsBackend.SyncSubscription(sub2)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	// Because we have two subscriptions
	require.Len(t, jsBackend.subscriptions, 2)

	// set metadata on NATS subscriptions
	// so that we can later verify if the nats subscriptions are the same (not re-created by Sync)
	msgLimit, bytesLimit := 2048, 2048
	for _, jsSub := range jsBackend.subscriptions {
		require.True(t, jsSub.IsValid())
		require.NoError(t, jsSub.SetPendingLimits(msgLimit, bytesLimit))
	}

	// Now, change the filter in subscription 1
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", evtesting.OrderCreatedEventTypeNotClean)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject from subscription 1
	newSubject, err := getCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// Check if total existing NATS subscriptions are correct
	// Because we have two subscriptions
	require.Len(t, jsBackend.subscriptions, 2)

	// check if the NATS subscription are NOT the same after sync for subscription 1
	// because the subscriptions should have being re-created for new subject
	jsSubject := jsBackend.GetJsSubjectToSubscribe(newSubject)
	jsSubKey := jsBackend.generateJsSubKey(jsSubject, sub)

	jsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := jsSub.PendingLimits()
	require.NoError(t, err)
	require.NotEqual(t, subMsgLimit, msgLimit)
	require.NotEqual(t, subBytesLimit, bytesLimit)

	// get cleaned subject for subscription 2
	cleanSubjectSub2, err := getCleanSubject(sub2.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, cleanSubjectSub2)

	// check if the NATS subscription are same after sync for subscription 2
	// because the subscriptions should NOT have being re-created as
	// subscription 2 was not modified
	jsSubject = jsBackend.GetJsSubjectToSubscribe(cleanSubjectSub2)
	jsSubKey = jsBackend.generateJsSubKey(jsSubject, sub2)

	jsSub = jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, jsSub)
	require.True(t, jsSub.IsValid())
	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err = jsSub.PendingLimits()
	require.NoError(t, err)
	require.Equal(t, subMsgLimit, msgLimit)
	require.Equal(t, subBytesLimit, bytesLimit)
}

// TestJetStream_isJsSubAssociatedWithKymaSub tests the isJsSubAssociatedWithKymaSub method.
func TestJetStream_isJsSubAssociatedWithKymaSub(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create subscription 1 and its JetStream subscription
	cleanSubject1 := "subOne"
	sub1 := evtesting.NewSubscription(cleanSubject1, "foo", evtesting.WithNotCleanFilter())
	jsSub1Key := jsBackend.generateJsSubKey(
		jsBackend.GetJsSubjectToSubscribe(cleanSubject1),
		sub1)

	// create subscription 2 and its JetStream subscription
	cleanSubject2 := "subOneTwo"
	sub2 := evtesting.NewSubscription(cleanSubject2, "foo", evtesting.WithNotCleanFilter())
	jsSub2Key := jsBackend.generateJsSubKey(
		jsBackend.GetJsSubjectToSubscribe(cleanSubject2),
		sub2)

	testCases := []struct {
		name            string
		givenJSSubKey   string
		givenKymaSubKey *eventingv1alpha1.Subscription
		wantResult      bool
	}{
		{
			name:            "",
			givenJSSubKey:   jsSub1Key,
			givenKymaSubKey: sub1,
			wantResult:      true,
		},
		{
			name:            "",
			givenJSSubKey:   jsSub2Key,
			givenKymaSubKey: sub2,
			wantResult:      true,
		},
		{
			name:            "",
			givenJSSubKey:   jsSub1Key,
			givenKymaSubKey: sub2,
			wantResult:      false,
		},
		{
			name:            "",
			givenJSSubKey:   jsSub2Key,
			givenKymaSubKey: sub1,
			wantResult:      false,
		},
	}

	for _, tC := range testCases {
		testCase := tC
		t.Run(testCase.name, func(t *testing.T) {
			gotResult, err := jsBackend.isJsSubAssociatedWithKymaSub(
				tC.givenJSSubKey,
				tC.givenKymaSubKey)
			require.Equal(t, gotResult, tC.wantResult)
			require.NoError(t, err)
		})
	}
}

// TestMultipleJSSubscriptionsToSameEvent tests the behaviour of JS
// when multiple subscriptions need to receive the same event.
func TestMultipleJSSubscriptionsToSameEvent(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	defaultMaxInflight := 1
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	// Create 3 subscriptions having the same sink and the same event type
	var subs [3]*eventingv1alpha1.Subscription
	for i := 0; i < len(subs); i++ {
		subs[i] = evtesting.NewSubscription(fmt.Sprintf("sub-%d", i), "foo",
			evtesting.WithNotCleanFilter(),
			evtesting.WithSinkURL(subscriber.SinkURL),
			evtesting.WithStatusConfig(defaultSubsConfig),
		)
		addJSCleanEventTypesToStatus(subs[i], testEnvironment.cleaner, jsBackend)
		// when
		err := jsBackend.SyncSubscription(subs[i])
		// then
		require.NoError(t, err)
	}

	// Send only one event. It should be multiplexed to 3 by NATS, cause 3 subscriptions exist
	data := "sampledata"
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	// Check for the 3 events that should be received by the subscriber
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	for i := 0; i < len(subs); i++ {
		require.NoError(t, subscriber.CheckEvent(expectedDataInStore))
	}
	// Delete all 3 subscription
	for i := 0; i < len(subs); i++ {
		require.NoError(t, jsBackend.DeleteSubscription(subs[i]))
	}
	// Check if all subscriptions are deleted in NATS
	// Send an event again which should not be delivered to subscriber
	newData := "test-data"
	require.NoError(t, SendEventToJetStream(jsBackend, newData))
	// Check for the event that did not reach the subscriber
	// Store should never return newdata hence CheckEvent should fail to match newdata
	notExpectedNewDataInStore := fmt.Sprintf("\"%s\"", newData)
	require.Error(t, subscriber.CheckEvent(notExpectedNewDataInStore))
}

// TestJSSubscriptionWithDuplicateFilters tests the subscription behaviour
// when duplicate filters are added.
func TestJSSubscriptionWithDuplicateFilters(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithFilter("", evtesting.OrderCreatedEventType),
		evtesting.WithFilter("", evtesting.OrderCreatedEventType),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	data := "sampledata"
	require.NoError(t, SendEventToJetStream(jsBackend, data))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))
	// There should be no more!
	require.Error(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJSSubscriptionWithMaxInFlightChange tests the maxAckPending
// to be equal to the MaxInFlightMessages when the server is not running.
func TestJSSubscriptionWithMaxInFlightChange(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	subscriber.Shutdown() // shutdown the subscriber intentionally here
	require.False(t, subscriber.IsRunning())

	defaultMaxInflight := 16
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInflight}
	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// when
	// send 2 * defaultMaxInflight number of events
	for i := 0; i < 2*defaultMaxInflight; i++ {
		data := fmt.Sprintf("sampledata%d", i)
		require.NoError(t, SendEventToJetStream(jsBackend, data))
	}

	// then
	require.Eventually(t, func() bool {
		consumerName := jsBackend.generateJsSubKey(sub.Status.CleanEventTypes[0], sub)
		// fetch consumer info from JetStream
		consumerInfo, err := jsBackend.jsCtx.ConsumerInfo(jsBackend.config.JSStreamName, consumerName)
		require.NoError(t, err)

		// since our subscriber is not in running state,
		// so these events will be pending for receiving an ACK from dispatchers
		// check consumer current maxAckPending
		return consumerInfo.NumAckPending == defaultMaxInflight
	}, 10*time.Second, 200*time.Millisecond)
}

// TestJSSubscriptionUsingCESDK tests that eventing works with Cloud events.
func TestJSSubscriptionUsingCESDK(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 1}

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithOrderCreatedFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	subject := evtesting.CloudEventType
	require.NoError(t, SendBinaryCloudEventToJetStream(jsBackend, jsBackend.GetJsSubjectToSubscribe(subject), evtesting.CloudEventData))
	require.NoError(t, subscriber.CheckEvent(evtesting.CloudEventData))
	require.NoError(t, SendStructuredCloudEventToJetStream(jsBackend, jsBackend.GetJsSubjectToSubscribe(subject), evtesting.StructuredCloudEvent))
	require.NoError(t, subscriber.CheckEvent("\""+evtesting.EventData+"\""))
	require.NoError(t, jsBackend.DeleteSubscription(sub))
}

// TODO: Enable this test once the ConnCloseHandler is implemented
/*func TestSubscription_JetStreamServerRestart(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 10}

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	// Create a subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner, jsBackend)

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	ev1data := "sampledata"
	require.NoError(t, SendEventToJetStream(jsBackend, ev1data))
	expectedEv1Data := fmt.Sprintf("\"%s\"", ev1data)
	require.NoError(t, subscriber.CheckEvent(expectedEv1Data))

	testEnvironment.natsServer.Shutdown()
	require.Eventually(t, func() bool {
		return !jsBackend.conn.IsConnected()
	}, 30*time.Second, 2*time.Second)

	_ = evtesting.RunNatsServerOnPort(
		evtesting.WithPort(testEnvironment.natsPort),
		evtesting.WithJetStreamEnabled())

	require.Eventually(t, func() bool {
		return jsBackend.conn.IsConnected()
	}, 60*time.Second, 2*time.Second)

	// After reconnect, event delivery should work again
	ev2data := "newsampledata"
	require.NoError(t, SendEventToJetStream(jsBackend, ev2data))
	expectedEv2Data := fmt.Sprintf("\"%s\"", ev2data)
	require.NoError(t, subscriber.CheckEvent(expectedEv2Data))
}*/

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                     url,
		JSStreamName:            defaultStreamName,
		JSStreamStorageType:     JetStreamStorageTypeMemory,
		JSStreamRetentionPolicy: JetStreamRetentionPolicyInterest,
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

func addJSCleanEventTypesToStatus(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, jsBackend *JetStream) {
	cleanedSubjects, _ := GetCleanSubjects(sub, cleaner)
	sub.Status.CleanEventTypes = jsBackend.GetJetStreamSubjects(cleanedSubjects)
}

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	jsBackend  *JetStream
	logger     *logger.Logger
	natsServer *server.Server
	jsClient   *jetStreamClient
	natsConfig env.NatsConfig
	cleaner    eventtype.Cleaner
	natsPort   int
}

// setupTestEnvironment is a TestEnvironment constructor
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer, natsPort := startNATSServer(evtesting.WithJetStreamEnabled())
	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

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
		natsPort:   natsPort,
	}
}
