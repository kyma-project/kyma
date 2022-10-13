package jetstreamv2

import (
	"fmt"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backenderrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2/errors"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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

	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber1.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
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
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
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

	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber1.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
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
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
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

	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
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
	sub.Spec.Types[0] = fmt.Sprintf("%schanged", evtestingv2.OrderCreatedUncleanEvent)
	// Sync the subscription status
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject
	newSubject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, newSubject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
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
	require.NoError(t, sendEventToJetStreamOnEventType(jsBackend, newSubject, data, sub.Spec.TypeMatching))
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

	// Create a subscription with single filter
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	firstSubject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
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
	newType := sub.Spec.Types[0]
	newType = fmt.Sprintf("%snew1", newType)
	sub.Spec.Types = append(sub.Spec.Types, newType)

	// get new cleaned subject
	secondSubject, err := getCleanEventType(newType, testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, secondSubject)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	require.Len(t, jsBackend.subscriptions, 2)
	// Verify that the nats subscriptions for first subject was not re-created
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, firstSubject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
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
	require.NoError(t, sendEventToJetStreamOnEventType(jsBackend, secondSubject, data, sub.Spec.TypeMatching))
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

	// Create a subscription with two filters
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	// add a second filter
	newType := sub.Spec.Types[0]
	newType = fmt.Sprintf("%snew1", newType)
	sub.Spec.Types = append(sub.Spec.Types, newType)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subjects
	firstSubject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, firstSubject)

	secondSubject, err := getCleanEventType(sub.Spec.Types[1], testEnvironment.cleaner)
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
	sub.Spec.Types = sub.Spec.Types[:1]
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	require.Len(t, jsBackend.subscriptions, 1)
	// Verify that the nats subscriptions for first subject was not re-created
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, firstSubject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
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
	require.NoError(t, sendEventToJetStreamOnEventType(jsBackend, secondSubject, data, sub.Spec.TypeMatching))
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

	// Create two subscriptions with single filter
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// given
	sub2 := evtestingv2.NewSubscription("sub2", "foo",
		evtestingv2.WithCleanEventTypeOld(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingExact(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub2, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub2)

	// then
	require.NoError(t, err)

	// test for exact type matching subscription
	data := "sampledata"
	require.NoError(t, sendEventToJetStreamOnEventType(jsBackend, evtestingv2.OrderCreatedEventType, data, sub2.Spec.TypeMatching))
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))

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
	sub.Spec.Types[0] = fmt.Sprintf("%schanged", evtestingv2.OrderCreatedEventTypeNotClean)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject from subscription 1
	newSubject, err := getCleanEventType(sub.Spec.Types[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// Check if total existing NATS subscriptions are correct
	// Because we have two subscriptions
	require.Len(t, jsBackend.subscriptions, 2)

	// check if the NATS subscription are NOT the same after sync for subscription 1
	// because the subscriptions should have being re-created for new subject
	jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, newSubject, sub.Spec.TypeMatching)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
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
	cleanSubjectSub2, err := getCleanEventType(sub2.Spec.Types[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, cleanSubjectSub2)

	// check if the NATS subscription are same after sync for subscription 2
	// because the subscriptions should NOT have being re-created as
	// subscription 2 was not modified
	jsSubject = jsBackend.getJetStreamSubject(sub2.Spec.Source, cleanSubjectSub2, sub2.Spec.TypeMatching)
	jsSubKey = NewSubscriptionSubjectIdentifier(sub2, jsSubject)
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

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	// Create 3 subscriptions having the same sink and the same event type
	var subs [3]*eventingv1alpha2.Subscription
	for i := 0; i < len(subs); i++ {
		subs[i] = evtestingv2.NewSubscription(fmt.Sprintf("sub-%d", i), "foo",
			evtestingv2.WithNotCleanEventSourceAndType(),
			evtestingv2.WithSinkURL(subscriber.SinkURL),
			evtestingv2.WithTypeMatchingStandard(),
			evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
		)
		require.NoError(t, AddJSCleanEventTypesToStatus(subs[i], testEnvironment.cleaner))
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
	// defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(defaultMaxInflight),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
		// fetch consumer info from JetStream
		consumerName := NewSubscriptionSubjectIdentifier(sub, jsBackend.getJetStreamSubject(sub.Spec.Source, sub.Status.Types[0].CleanType, sub.Spec.TypeMatching)).ConsumerName()
		consumerInfo, err := jsBackend.jsCtx.ConsumerInfo(jsBackend.Config.JSStreamName, consumerName)
		require.NoError(t, err)

		// since our subscriber is not in running state,
		// so these events will be pending for receiving an ACK from dispatchers
		// check consumer current maxAckPending
		return consumerInfo.NumAckPending == defaultMaxInflight
	}, 10*time.Second, 10*time.Millisecond)
}

// TestJSSubscriptionRedeliverWithFailedDispatch tests the redelivering
// of event when the dispatch fails.
func TestJSSubscriptionRedeliverWithFailedDispatch(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	subscriber.Shutdown() // shutdown the subscriber intentionally
	require.False(t, subscriber.IsRunning())

	// create a new Subscription
	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// when
	// send an event
	ev2data := "newsampledata"
	require.NoError(t, SendEventToJetStream(jsBackend, ev2data))

	// then
	// it should have failed to dispatch
	expectedEv2Data := fmt.Sprintf("\"%s\"", ev2data)
	require.Error(t, subscriber.CheckEvent(expectedEv2Data))

	// when
	// start a new subscriber
	subscriber = evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())
	// and update sink in the subscription
	sub.Spec.Sink = subscriber.SinkURL
	require.NoError(t, jsBackend.SyncSubscription(sub))

	// then
	// the same event should be redelivered
	require.Eventually(t, func() bool {
		return subscriber.CheckEvent(expectedEv2Data) == nil
	}, 60*time.Second, 5*time.Second)
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

	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())

	sub := evtestingv2.NewSubscription("sub", "foo",
		evtestingv2.WithNotCleanEventSourceAndType(),
		evtestingv2.WithSinkURL(subscriber.SinkURL),
		evtestingv2.WithTypeMatchingStandard(),
		evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
	)
	require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	subject := evtestingv2.OrderCreatedCleanEvent
	require.NoError(t, sendCloudEventToJetStream(jsBackend, jsBackend.getJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching), evtestingv2.CloudEventData, types.ContentModeBinary))
	require.NoError(t, subscriber.CheckEvent(evtestingv2.CloudEventData))
	require.NoError(t, sendCloudEventToJetStream(jsBackend, jsBackend.getJetStreamSubject(sub.Spec.Source, subject, sub.Spec.TypeMatching), evtestingv2.StructuredCloudEvent, types.ContentModeStructured))
	require.NoError(t, subscriber.CheckEvent("\""+evtestingv2.EventData+"\""))
	require.NoError(t, jsBackend.DeleteSubscription(sub))
}

// TestJetStream_NoNATSSubscription tests if the error is being triggered
// when expected entries in js.subscriptions map are missing.
func TestJetStream_NATSSubscriptionCount(t *testing.T) {
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

	testCases := []struct {
		name                            string
		subOpts                         []evtestingv2.SubscriptionOpt
		givenManuallyDeleteSubscription bool
		givenFilterToDelete             string
		wantNatsSubsLen                 int
		wantErr                         func(t *testing.T, givenError error)
	}{
		{
			name: "No error should happen, when there is only one type",
			subOpts: []evtestingv2.SubscriptionOpt{
				evtestingv2.WithSinkURL(subscriber.SinkURL),
				evtestingv2.WithNotCleanEventSourceAndType(),
				evtestingv2.WithTypeMatchingStandard(),
				evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
			},
			givenManuallyDeleteSubscription: false,
			wantNatsSubsLen:                 1,
			wantErr:                         nil,
		},
		{
			name: "No error expected when js.subscriptions map has entries for all the eventTypes",
			subOpts: []evtestingv2.SubscriptionOpt{
				evtestingv2.WithNotCleanEventSourceAndType(),
				evtestingv2.WithCleanEventTypeOld(),
				evtestingv2.WithTypeMatchingStandard(),
				evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
			},
			givenManuallyDeleteSubscription: false,
			wantNatsSubsLen:                 2,
			wantErr:                         nil,
		},
		{
			name: "An error is expected, when we manually delete a subscription from js.subscriptions map",
			subOpts: []evtestingv2.SubscriptionOpt{
				evtestingv2.WithNotCleanEventSourceAndType(),
				evtestingv2.WithCleanEventTypeOld(),
				evtestingv2.WithTypeMatchingStandard(),
				evtestingv2.WithMaxInFlight(DefaultMaxInFlights),
			},
			givenManuallyDeleteSubscription: true,
			givenFilterToDelete:             evtestingv2.OrderCreatedEventType,
			wantNatsSubsLen:                 2,
			wantErr: func(t *testing.T, givenError error) {
				var wantError *backenderrors.MissingSubscriptionError
				require.ErrorAs(t, givenError, &wantError)
				assert.Equal(t, wantError.Subject.CleanType, evtestingv2.OrderCreatedEventType)
			},
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// create a new subscription with no filters
			sub := evtestingv2.NewSubscription("sub"+fmt.Sprint(i), "foo",
				tc.subOpts...,
			)
			require.NoError(t, AddJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

			// when
			err := jsBackend.SyncSubscription(sub)
			require.NoError(t, err)
			require.Equal(t, len(jsBackend.subscriptions), tc.wantNatsSubsLen)

			if tc.givenManuallyDeleteSubscription {
				// manually delete the subscription from map
				jsSubject := jsBackend.getJetStreamSubject(sub.Spec.Source, tc.givenFilterToDelete, sub.Spec.TypeMatching)
				jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
				delete(jsBackend.subscriptions, jsSubKey)
			}

			err = jsBackend.SyncSubscription(sub)
			testEnvironment.logger.WithContext().Error(err)

			if tc.wantErr != nil {
				// the createConsumer function won't create a new Subscription,
				// because the subscription was manually deleted from the js.subscriptions map
				// hence the consumer will be shown in the NATS Backend as still bound
				err = jsBackend.SyncSubscription(sub)
				tc.wantErr(t, err)
			}

			// empty the js.subscriptions map
			require.NoError(t, jsBackend.DeleteSubscription(sub))
		})
	}
}

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                     url,
		MaxReconnects:           DefaultMaxReconnects,
		ReconnectWait:           3 * time.Second,
		JSStreamName:            DefaultStreamName,
		JSStreamStorageType:     StorageTypeMemory,
		JSStreamRetentionPolicy: RetentionPolicyInterest,
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

// setupTestEnvironment is a TestEnvironment constructor.
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer, natsPort, err := natstesting.StartNATSServer(evtesting.WithJetStreamEnabled())
	require.NoError(t, err)
	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	jsClient := getJetStreamClient(t, natsConfig.URL)
	jsCleaner := cleaner.NewJetStreamCleaner(defaultLogger)
	jsBackend := NewJetStream(natsConfig, metricsCollector, jsCleaner, defaultLogger)

	return &TestEnvironment{
		jsBackend:  jsBackend,
		logger:     defaultLogger,
		natsServer: natsServer,
		jsClient:   jsClient,
		natsConfig: natsConfig,
		cleaner:    jsCleaner,
		natsPort:   natsPort,
	}
}
