package jetstream

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	cleanerv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	defaultStreamName    = "kyma"
	defaultMaxReconnects = 10
	defaultMaxInFlights  = 10

	// maxJetStreamConsumerNameLength is the maximum preferred length for the JetStream consumer names
	// as per https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming
	maxJetStreamConsumerNameLength = 32
)

type jetStreamClient struct {
	nats.JetStreamContext
	natsConn *nats.Conn
}

// TestJetStream_SubscriptionDeletion tests the creation and deletion
// of a JetStream subscription on the NATS server.
func TestJetStream_SubscriptionDeletion(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
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
	jsSubject := jsBackend.GetJetStreamSubject(subject)
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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
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
	jsSubject := jsBackend.GetJetStreamSubject(subject)
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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject := jsBackend.GetJetStreamSubject(subject)
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
	oldJsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, oldJsSub)
	require.True(t, oldJsSub.IsValid())

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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject
	newSubject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	require.Len(t, jsBackend.subscriptions, 1)
	jsSubject = jsBackend.GetJetStreamSubject(newSubject)
	jsSubKey = NewSubscriptionSubjectIdentifier(sub, jsSubject)
	newJsSub := jsBackend.subscriptions[jsSubKey]
	require.NotNil(t, newJsSub)
	require.True(t, newJsSub.IsValid())
	// make sure old filter doesn't have any JetStream consumer
	oldCon, err := oldJsSub.ConsumerInfo()
	require.Nil(t, oldCon)
	require.ErrorIs(t, err, nats.ErrConsumerNotFound)

	// check the metadata, if they are NOT same then it means that NATS subscription
	// were re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := newJsSub.PendingLimits()
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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	firstSubject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
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
	secondSubject, err := backendnats.GetCleanSubject(newFilter, testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, secondSubject)
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	// Because we have two filters (i.e. two subjects)
	require.Len(t, jsBackend.subscriptions, 2)
	// Verify that the nats subscriptions for first subject was not re-created
	jsSubject := jsBackend.GetJetStreamSubject(firstSubject)
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
	require.NoError(t, SendEventToJetStreamOnEventType(jsBackend, secondSubject, data))
	// The sink should receive the event for second subject
	require.NoError(t, subscriber.CheckEvent(expectedDataInStore))
}

// TestJetStreamSubAfterSync_FilterRemoved tests the SyncSubscription method
// when a filter is removed from subscription.
func TestJetStreamSubAfterSync_FilterRemoved(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subjects
	firstSubject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, firstSubject)

	secondSubject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[1], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, secondSubject)
	secondJsSubject := jsBackend.GetJetStreamSubject(secondSubject)
	secondJsSubKey := NewSubscriptionSubjectIdentifier(sub, secondJsSubject)
	secondJsSub := jsBackend.subscriptions[secondJsSubKey]
	require.NotNil(t, secondJsSub)
	require.True(t, secondJsSub.IsValid())

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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// Check if total existing NATS subscriptions are correct
	require.Len(t, jsBackend.subscriptions, 1)
	// Verify that the nats subscriptions for first subject was not re-created
	firstJsSubject := jsBackend.GetJetStreamSubject(firstSubject)
	firstJsSubKey := NewSubscriptionSubjectIdentifier(sub, firstJsSubject)
	firstJsSub := jsBackend.subscriptions[firstJsSubKey]
	require.NotNil(t, firstJsSub)
	require.True(t, firstJsSub.IsValid())
	// make sure old filter doesn't have any JetStream consumer
	secondCon, err := secondJsSub.ConsumerInfo()
	require.Nil(t, secondCon)
	require.ErrorIs(t, err, nats.ErrConsumerNotFound)

	// check the metadata, if they are now same then it means that NATS subscription
	// were not re-created by SyncSubscription method
	subMsgLimit, subBytesLimit, err := firstJsSub.PendingLimits()
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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
	require.NoError(t, addJSCleanEventTypesToStatus(sub2, testEnvironment.cleaner))

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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject from subscription 1
	newSubject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// Check if total existing NATS subscriptions are correct
	// Because we have two subscriptions
	require.Len(t, jsBackend.subscriptions, 2)

	// check if the NATS subscription are NOT the same after sync for subscription 1
	// because the subscriptions should have being re-created for new subject
	jsSubject := jsBackend.GetJetStreamSubject(newSubject)
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
	cleanSubjectSub2, err := backendnats.GetCleanSubject(sub2.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, cleanSubjectSub2)

	// check if the NATS subscription are same after sync for subscription 2
	// because the subscriptions should NOT have being re-created as
	// subscription 2 was not modified
	jsSubject = jsBackend.GetJetStreamSubject(cleanSubjectSub2)
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

// TestJetStream_isJsSubAssociatedWithKymaSub tests the isJsSubAssociatedWithKymaSub method.
func TestJetStream_isJsSubAssociatedWithKymaSub(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create subscription 1 and its JetStream subscription
	cleanSubject1 := "subOne"
	sub1 := evtesting.NewSubscription(cleanSubject1, "foo", evtesting.WithNotCleanFilter())
	jsSub1Key := NewSubscriptionSubjectIdentifier(sub1, cleanSubject1)

	// create subscription 2 and its JetStream subscription
	cleanSubject2 := "subOneTwo"
	sub2 := evtesting.NewSubscription(cleanSubject2, "foo", evtesting.WithNotCleanFilter())
	jsSub2Key := NewSubscriptionSubjectIdentifier(sub2, cleanSubject2)

	testCases := []struct {
		name            string
		givenJSSubKey   SubscriptionSubjectIdentifier
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
			gotResult := jsBackend.isJsSubAssociatedWithKymaSub(tC.givenJSSubKey, tC.givenKymaSubKey)
			require.Equal(t, tC.wantResult, gotResult)
		})
	}
}

// TestMultipleJSSubscriptionsToSameEvent tests the behaviour of JS
// when multiple subscriptions need to receive the same event.
func TestMultipleJSSubscriptionsToSameEvent(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
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
		require.NoError(t, addJSCleanEventTypesToStatus(subs[i], testEnvironment.cleaner))
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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
		consumerName := NewSubscriptionSubjectIdentifier(sub, jsBackend.GetJetStreamSubject(sub.Status.CleanEventTypes[0])).ConsumerName()
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
	testEnvironment := setupTestEnvironment(t, false)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// create New Subscriber
	subscriber := evtesting.NewSubscriber()
	subscriber.Shutdown() // shutdown the subscriber intentionally
	require.False(t, subscriber.IsRunning())

	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}
	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL(subscriber.SinkURL),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

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
	testEnvironment := setupTestEnvironment(t, false)
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	subject := evtesting.CloudEventType
	require.NoError(t, SendBinaryCloudEventToJetStream(jsBackend, jsBackend.GetJetStreamSubject(subject), evtesting.CloudEventData))
	require.NoError(t, subscriber.CheckEvent(evtesting.CloudEventData))
	require.NoError(t, SendStructuredCloudEventToJetStream(jsBackend, jsBackend.GetJetStreamSubject(subject), evtesting.StructuredCloudEvent))
	require.NoError(t, subscriber.CheckEvent("\""+evtesting.EventData+"\""))
	require.NoError(t, jsBackend.DeleteSubscription(sub))
}

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                     url,
		MaxReconnects:           defaultMaxReconnects,
		ReconnectWait:           3 * time.Second,
		JSStreamName:            defaultStreamName,
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

func addJSCleanEventTypesToStatus(sub *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner) error {
	cleanEventType, err := backendnats.GetCleanSubjects(sub, cleaner)
	if err != nil {
		return err
	}
	sub.Status.CleanEventTypes = cleanEventType
	return nil
}

// TestEnvironment provides mocked resources for tests.
type TestEnvironment struct {
	jsBackend   *JetStream
	jsBackendv2 *jetstreamv2.JetStream
	logger      *logger.Logger
	natsServer  *server.Server
	jsClient    *jetStreamClient
	natsConfig  env.NatsConfig
	cleaner     eventtype.Cleaner
	cleanerv2   cleanerv1alpha2.Cleaner
	natsPort    int
}

// setupTestEnvironment is a TestEnvironment constructor.
func setupTestEnvironment(t *testing.T, newCRD bool) *TestEnvironment {
	natsServer, natsPort, err := natstesting.StartNATSServer(evtesting.WithJetStreamEnabled())
	require.NoError(t, err)
	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	jsClient := getJetStreamClient(t, natsConfig.URL)

	cleanerv2 := cleanerv1alpha2.NewJetStreamCleaner(defaultLogger)
	cleaner := backendnats.CreateEventTypeCleaner(evtesting.EventTypePrefix, evtesting.ApplicationNameNotClean, defaultLogger)
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}

	var jsBackend *JetStream
	var jsBackendNew *jetstreamv2.JetStream
	if newCRD {
		jsBackendNew = jetstreamv2.NewJetStream(natsConfig, metricsCollector, cleanerv2, defaultSubsConfig, defaultLogger)
	} else {
		jsBackend = NewJetStream(natsConfig, metricsCollector, defaultLogger)
	}

	return &TestEnvironment{
		jsBackend:   jsBackend,
		jsBackendv2: jsBackendNew,
		logger:      defaultLogger,
		natsServer:  natsServer,
		jsClient:    jsClient,
		natsConfig:  natsConfig,
		cleaner:     cleaner,
		cleanerv2:   cleanerv2,
		natsPort:    natsPort,
	}
}

// TestSubscriptionSubjectIdentifierEqual checks the equality of two SubscriptionSubjectIdentifier instances and their consumer names.
func TestSubscriptionSubjectIdentifierEqual(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		givenIdentifier1 SubscriptionSubjectIdentifier
		givenIdentifier2 SubscriptionSubjectIdentifier
		wantEqual        bool
	}{
		// instances are equal
		{
			name: "should be equal if the two instances are identical",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: true,
		},
		// instances are not equal
		{
			name: "should not be equal if only name is different",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-2", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
		{
			name: "should not be equal if only namespace is different",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-2"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
		{
			name: "should not be equal if only subject is different",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v2",
			),
			wantEqual: false,
		},
		// possible naming collisions
		{
			name: "should not be equal if subject is the same but name and namespace are swapped",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"),
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("ns-1", "sub-1"),
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
		{
			name: "should not be equal if subject is the same but name and namespace are only equal if joined together",
			givenIdentifier1: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1", "ns-1"), // evaluates to "sub-1ns-1" when joined
				"prefix.app.event.operation.v1",
			),
			givenIdentifier2: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub-1n", "s-1"), // evaluates to "sub-1ns-1" when joined
				"prefix.app.event.operation.v1",
			),
			wantEqual: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotInstanceEqual := reflect.DeepEqual(tc.givenIdentifier1, tc.givenIdentifier2)
			assert.Equal(t, tc.wantEqual, gotInstanceEqual)

			gotConsumerNameEqual := tc.givenIdentifier1.ConsumerName() == tc.givenIdentifier2.ConsumerName()
			assert.Equal(t, tc.wantEqual, gotConsumerNameEqual)
		})
	}
}

// TestSubscriptionSubjectIdentifierConsumerNameLength checks that the SubscriptionSubjectIdentifier consumer name
// length is equal to the recommended length by JetStream.
func TestSubscriptionSubjectIdentifierConsumerNameLength(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                   string
		givenIdentifier        SubscriptionSubjectIdentifier
		wantConsumerNameLength int
	}{
		{
			name: "short string values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub", "ns"),
				"app.event.operation.v1",
			),
			wantConsumerNameLength: maxJetStreamConsumerNameLength,
		},
		{
			name: "long string values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("some-test-subscription", "some-test-namespace"),
				"some.test.prefix.some-test-application.some-test-event-name.some-test-operation.some-test-version",
			),
			wantConsumerNameLength: maxJetStreamConsumerNameLength,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantConsumerNameLength, len(tc.givenIdentifier.ConsumerName()))
		})
	}
}

// TestSubscriptionSubjectIdentifierNamespacedName checks the syntax of the SubscriptionSubjectIdentifier namespaced name.
func TestSubscriptionSubjectIdentifierNamespacedName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name               string
		givenIdentifier    SubscriptionSubjectIdentifier
		wantNamespacedName string
	}{
		{
			name: "short name and namespace values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("sub", "ns"),
				"app.event.operation.v1",
			),
			wantNamespacedName: "ns/sub",
		},
		{
			name: "long name and namespace values",
			givenIdentifier: NewSubscriptionSubjectIdentifier(
				evtesting.NewSubscription("some-test-subscription", "some-test-namespace"),
				"app.event.operation.v1",
			),
			wantNamespacedName: "some-test-namespace/some-test-subscription",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantNamespacedName, tc.givenIdentifier.NamespacedName())
		})
	}
}

// TestJetStream_NoNATSSubscription tests if the error is being triggered
// when expected entries in js.subscriptions map are missing.
func TestJetStream_NATSSubscriptionCount(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
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
		subOpts                         []evtesting.SubscriptionOpt
		givenManuallyDeleteSubscription bool
		givenFilterToDelete             string
		wantNatsSubsLen                 int
		wantErr                         bool
		wantErrText                     string
	}{
		{
			name: "No error should happen, when there are no filters",
			subOpts: []evtesting.SubscriptionOpt{
				evtesting.WithSinkURL(subscriber.SinkURL),
				evtesting.WithEmptyFilter(),
				evtesting.WithStatusConfig(env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}),
			},
			givenManuallyDeleteSubscription: false,
			wantNatsSubsLen:                 0,
			wantErr:                         false,
		},
		{
			name: "No error expected when js.subscriptions map has entries for all the eventTypes",
			subOpts: []evtesting.SubscriptionOpt{
				evtesting.WithFilter("", evtesting.OrderCreatedEventType),
				evtesting.WithFilter("", evtesting.NewOrderCreatedEventType),
				evtesting.WithStatusConfig(env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}),
			},
			givenManuallyDeleteSubscription: false,
			wantNatsSubsLen:                 2,
			wantErr:                         false,
		},
		{
			name: "An error is expected, when we manually delete a subscription from js.subscriptions map",
			subOpts: []evtesting.SubscriptionOpt{
				evtesting.WithFilter("", evtesting.OrderCreatedEventType),
				evtesting.WithFilter("", evtesting.NewOrderCreatedEventType),
				evtesting.WithStatusConfig(env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}),
			},
			givenManuallyDeleteSubscription: true,
			givenFilterToDelete:             evtesting.NewOrderCreatedEventType,
			wantNatsSubsLen:                 2,
			wantErr:                         true,
			wantErrText:                     fmt.Sprintf(MissingNATSSubscriptionMsgWithInfo, evtesting.NewOrderCreatedEventType),
		},
	}
	for i, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// create a new subscription with no filters
			sub := evtesting.NewSubscription("sub"+fmt.Sprint(i), "foo",
				tc.subOpts...,
			)
			require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

			// when
			err := jsBackend.SyncSubscription(sub)
			require.NoError(t, err)
			assert.Equal(t, len(jsBackend.subscriptions), tc.wantNatsSubsLen)

			if tc.givenManuallyDeleteSubscription {
				// manually delete the subscription from map
				jsSubject := jsBackend.GetJetStreamSubject(tc.givenFilterToDelete)
				jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)
				delete(jsBackend.subscriptions, jsSubKey)
			}

			err = jsBackend.SyncSubscription(sub)
			testEnvironment.logger.WithContext().Error(err)
			require.Equal(t, err != nil, tc.wantErr)

			if tc.wantErr {
				// the createConsumer function won't create a new Subscription,
				// because the subscription was manually deleted from the js.subscriptions map
				// hence the consumer will be shown in the NATS Backend as still bound
				err = jsBackend.SyncSubscription(sub)
				require.True(t, strings.Contains(err.Error(), tc.wantErrText))
			}

			// empty the js.subscriptions map
			require.NoError(t, jsBackend.DeleteSubscription(sub))
		})
	}
}

// TestJetStream_CheckConsumerConfig that the latest Subscription Config changes will be propagated to the consumer.
func TestJetStream_CheckConsumerConfig(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	initErr := jsBackend.Initialize(nil)
	require.NoError(t, initErr)

	// set the initial MaxInFlight
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}

	// create a new Subscription
	sub := evtesting.NewSubscription("sub", "foo",
		evtesting.WithNotCleanFilter(),
		evtesting.WithSinkURL("test.svc.local"),
		evtesting.WithStatusConfig(defaultSubsConfig),
	)
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// init the required strings
	jsSubject := jsBackend.GetJetStreamSubject(sub.Status.CleanEventTypes[0])
	jsSubKey := NewSubscriptionSubjectIdentifier(sub, jsSubject)

	// when
	require.NoError(t, jsBackend.SyncSubscription(sub))

	// then
	// check that the consumer info war created with the expected maxAckPending value
	consumerInfo, _ := jsBackend.jsCtx.ConsumerInfo(jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
	require.NotNil(t, consumerInfo)
	require.Equal(t, consumerInfo.Config.MaxAckPending, defaultMaxInFlights)

	// given
	// set the new MaxInFlight value in the Subscription
	newMaxInFlight := 20
	sub.Spec.Config = &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: newMaxInFlight}

	// when
	require.NoError(t, jsBackend.SyncSubscription(sub))

	// then
	// check that the new value was propagated dto the consumer
	consumerInfo, err := jsBackend.jsCtx.ConsumerInfo(jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
	require.NotNil(t, consumerInfo)
	require.NoError(t, err)

	require.Equal(t, consumerInfo.Config.MaxAckPending, newMaxInFlight)

	// given
	// unset the config which should lead to returning to defaults
	sub.Spec.Config = nil
	// imitate the reconciler call, which will set the maxInFlight value back to the default value
	sub.Status.Config.MaxInFlightMessages = defaultMaxInFlights

	// when
	require.NoError(t, jsBackend.SyncSubscription(sub))

	// then
	// check that the new value was propagated dto the consumer
	consumerInfo, err = jsBackend.jsCtx.ConsumerInfo(jsBackend.Config.JSStreamName, jsSubKey.ConsumerName())
	require.NotNil(t, consumerInfo)
	require.NoError(t, err)

	require.Equal(t, consumerInfo.Config.MaxAckPending, defaultMaxInFlights)
}

// TestJetStreamSubAfterSync_ForExplicitlyBoundSubscriptionDeletion tests the SyncSubscription method
// when the filters are changed in subscription after NATS JetStream restart. It also
// verifies that explicitly bound subscription is deleted if it is filter is gone.
func TestJetStreamSubAfterSync_ForExplicitlyBoundSubscriptionDeletion(t *testing.T) {
	// given
	testEnvironment := setupTestEnvironment(t, false)
	jsBackend := testEnvironment.jsBackend
	defer testEnvironment.natsServer.Shutdown()
	defer testEnvironment.jsClient.natsConn.Close()
	defer func() { _ = testEnvironment.jsClient.DeleteStream(defaultStreamName) }()

	testEnvironment.jsBackend.Config.JSStreamStorageType = StorageTypeFile
	testEnvironment.jsBackend.Config.MaxReconnects = 0
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
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err := jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get cleaned subject
	subject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, subject)

	require.Len(t, jsBackend.subscriptions, 1)
	oldJsSubject := jsBackend.GetJetStreamSubject(subject)
	oldJsSubKey := NewSubscriptionSubjectIdentifier(sub, oldJsSubject)
	oldJsSub := jsBackend.subscriptions[oldJsSubKey]
	require.NotNil(t, oldJsSub)
	require.True(t, oldJsSub.IsValid())

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

	// shutdown the JetStream and start so that existing subscription gets invalid.
	testEnvironment.natsServer.Shutdown()
	require.Eventually(t, func() bool {
		return !testEnvironment.jsBackend.conn.IsConnected()
	}, 30*time.Second, 2*time.Second)

	// when
	// restart the NATS server
	_ = evtesting.RunNatsServerOnPort(
		evtesting.WithPort(testEnvironment.natsPort),
		evtesting.WithJetStreamEnabled())
	// the unacknowledged message must still be present in the stream

	require.Eventually(t, func() bool {
		info, streamErr := testEnvironment.jsClient.StreamInfo(defaultStreamName)
		require.NoError(t, streamErr)
		return info != nil && streamErr == nil
	}, 60*time.Second, 5*time.Second)

	// SyncSubscription binds the existing subscription to JetStream created one
	err = jsBackend.SyncSubscription(sub)
	// then
	require.NoError(t, err)

	// given
	// Now, change the filter in subscription
	sub.Spec.Filter.Filters[0].EventType.Value = fmt.Sprintf("%schanged", evtesting.OrderCreatedEventTypeNotClean)
	// Sync the subscription status
	require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

	// when
	err = jsBackend.SyncSubscription(sub)

	// then
	require.NoError(t, err)

	// get new cleaned subject
	newSubject, err := backendnats.GetCleanSubject(sub.Spec.Filter.Filters[0], testEnvironment.cleaner)
	require.NoError(t, err)
	require.NotEmpty(t, newSubject)

	// check if the NATS subscription are NOT the same after sync
	// because the subscriptions should have being re-created for new subject
	require.Len(t, jsBackend.subscriptions, 1)
	newJsSubject := jsBackend.GetJetStreamSubject(newSubject)
	newJsSubKey := NewSubscriptionSubjectIdentifier(sub, newJsSubject)
	newJsSub := jsBackend.subscriptions[newJsSubKey]
	require.NotNil(t, newJsSub)
	require.True(t, newJsSub.IsValid())
	// make sure new filter does have JetStream consumer
	newCon, err := jsBackend.jsCtx.ConsumerInfo(jsBackend.Config.JSStreamName, newJsSubKey.consumerName)
	require.NotNil(t, newCon)
	require.NoError(t, err)
	// make sure old filter doesn't have any JetStream consumer
	oldCon, err := jsBackend.jsCtx.ConsumerInfo(jsBackend.Config.JSStreamName, oldJsSubKey.consumerName)
	require.Nil(t, oldCon)
	require.ErrorIs(t, err, nats.ErrConsumerNotFound)
}
