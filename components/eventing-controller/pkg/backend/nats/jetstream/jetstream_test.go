package jetstream

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	cleanerv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
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

// TestJetStreamInitialize_NoStreamExists tests if a stream is
// created when no stream exists in JetStream.
func TestJetStreamInitialize_NoStreamExists(t *testing.T) {
	// given
	for _, newCRD := range []bool{true, false} {
		t.Run(fmt.Sprintf("Enabled New Crd Version: %v", newCRD), func(t *testing.T) {
			testEnvironment := setupTestEnvironment(t, newCRD)
			natsConfig, jsClient := testEnvironment.natsConfig, testEnvironment.jsClient
			defer testEnvironment.natsServer.Shutdown()
			defer jsClient.natsConn.Close()

			// No stream exists
			_, err := jsClient.StreamInfo(natsConfig.JSStreamName)
			require.True(t, errors.Is(err, nats.ErrStreamNotFound))

			// when
			var initErr error
			if newCRD {
				initErr = testEnvironment.jsBackendv2.Initialize(nil)
			} else {
				initErr = testEnvironment.jsBackend.Initialize(nil)
			}

			// then
			// A stream is created
			require.NoError(t, initErr)
			streamInfo, err := jsClient.StreamInfo(natsConfig.JSStreamName)
			require.NoError(t, err)
			require.NotNil(t, streamInfo)
		})
	}
}

// TestJetStreamInitialize_StreamExists tests if a stream is
// reused and not created when a stream exists in JetStream.
func TestJetStreamInitialize_StreamExists(t *testing.T) {
	// given
	for _, newCRD := range []bool{true, false} {
		t.Run(fmt.Sprintf("Enabled New Crd Version: %v", newCRD), func(t *testing.T) {
			testEnvironment := setupTestEnvironment(t, newCRD)
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
			var initErr error
			if newCRD {
				initErr = testEnvironment.jsBackendv2.Initialize(nil)
			} else {
				initErr = testEnvironment.jsBackend.Initialize(nil)
			}

			// then
			// No new stream should be created
			require.NoError(t, initErr)
			reusedStreamInfo, err := jsClient.StreamInfo(natsConfig.JSStreamName)
			require.NoError(t, err)
			require.Equal(t, reusedStreamInfo.Created, createdStreamInfo.Created)
		})
	}
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
	jsSubject := jsBackend.GetJetStreamSubject(newSubject)
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

// TestJetStream_ServerRestart tests that eventing works when NATS server is restarted
// for scenarios involving the stream storage type and when reconnect attempts are exhausted or not.
func TestJetStream_ServerRestart(t *testing.T) { //nolint:gocognit
	// given
	subscriber := evtesting.NewSubscriber()
	defer subscriber.Shutdown()
	require.True(t, subscriber.IsRunning())
	defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}

	testCases := []struct {
		name                  string
		givenMaxReconnects    int
		givenStorageType      string
		givenEnableCRDVersion bool
	}{
		{
			name:                  "with reconnects disabled and memory storage for streams",
			givenMaxReconnects:    0,
			givenStorageType:      StorageTypeMemory,
			givenEnableCRDVersion: false,
		},
		{
			name:                  "with reconnects enabled and memory storage for streams",
			givenMaxReconnects:    defaultMaxReconnects,
			givenStorageType:      StorageTypeMemory,
			givenEnableCRDVersion: false,
		},
		{
			name:                  "with reconnects disabled and file storage for streams",
			givenMaxReconnects:    0,
			givenStorageType:      StorageTypeFile,
			givenEnableCRDVersion: false,
		},
		{
			name:                  "with reconnects enabled and file storage for streams",
			givenMaxReconnects:    defaultMaxReconnects,
			givenStorageType:      StorageTypeFile,
			givenEnableCRDVersion: false,
		},
		{
			name:                  "with reconnects disabled and memory storage for streams",
			givenMaxReconnects:    0,
			givenStorageType:      StorageTypeMemory,
			givenEnableCRDVersion: true,
		},
		{
			name:                  "with reconnects enabled and memory storage for streams",
			givenMaxReconnects:    defaultMaxReconnects,
			givenStorageType:      StorageTypeMemory,
			givenEnableCRDVersion: true,
		},
		{
			name:                  "with reconnects disabled and file storage for streams",
			givenMaxReconnects:    0,
			givenStorageType:      StorageTypeFile,
			givenEnableCRDVersion: true,
		},
		{
			name:                  "with reconnects enabled and file storage for streams",
			givenMaxReconnects:    defaultMaxReconnects,
			givenStorageType:      StorageTypeFile,
			givenEnableCRDVersion: true,
		},
	}

	for id, tc := range testCases {
		tc, id := tc, id
		t.Run(tc.name, func(t *testing.T) {
			// given
			testEnvironment := setupTestEnvironment(t, tc.givenEnableCRDVersion)
			defer testEnvironment.natsServer.Shutdown()
			defer testEnvironment.jsClient.natsConn.Close()
			defer func() { _ = testEnvironment.jsClient.DeleteStream(defaultStreamName) }()
			var err error
			if tc.givenEnableCRDVersion {
				testEnvironment.jsBackendv2.Config.JSStreamStorageType = tc.givenStorageType
				testEnvironment.jsBackendv2.Config.MaxReconnects = tc.givenMaxReconnects
				err = testEnvironment.jsBackendv2.Initialize(nil)
			} else {
				testEnvironment.jsBackend.Config.JSStreamStorageType = tc.givenStorageType
				testEnvironment.jsBackend.Config.MaxReconnects = tc.givenMaxReconnects
				err = testEnvironment.jsBackend.Initialize(nil)
			}
			require.NoError(t, err)

			// Create a subscription
			subName := fmt.Sprintf("%s%d", "sub", id)
			var sub *eventingv1alpha1.Subscription
			var subv2 *eventingv1alpha2.Subscription
			if tc.givenEnableCRDVersion {
				subv2 = evtestingv2.NewSubscription(subName, "foo",
					evtestingv2.WithNotCleanEventSourceAndType(),
					evtestingv2.WithSinkURL(subscriber.SinkURL),
					evtestingv2.WithTypeMatchingStandard(),
					evtestingv2.WithMaxInFlight(defaultMaxInFlights),
				)
				require.NoError(t, jetstreamv2.AddJSCleanEventTypesToStatus(subv2, testEnvironment.cleanerv2))

				// when
				err = testEnvironment.jsBackendv2.SyncSubscription(subv2)
			} else {
				sub = evtesting.NewSubscription(subName, "foo",
					evtesting.WithNotCleanFilter(),
					evtesting.WithSinkURL(subscriber.SinkURL),
					evtesting.WithStatusConfig(defaultSubsConfig),
				)
				require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

				// when
				err = testEnvironment.jsBackend.SyncSubscription(sub)
			}

			// then
			require.NoError(t, err)

			ev1data := fmt.Sprintf("%s%d", "sampledata", id)
			if tc.givenEnableCRDVersion {
				require.NoError(t, jetstreamv2.SendEventToJetStream(testEnvironment.jsBackendv2, ev1data))
			} else {
				require.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, ev1data))
			}
			expectedEv1Data := fmt.Sprintf("%q", ev1data)
			require.NoError(t, subscriber.CheckEvent(expectedEv1Data))

			// given
			testEnvironment.natsServer.Shutdown()
			require.Eventually(t, func() bool {
				if tc.givenEnableCRDVersion {
					return !testEnvironment.jsBackendv2.Conn.IsConnected()
				}
				return !testEnvironment.jsBackend.conn.IsConnected()
			}, 30*time.Second, 2*time.Second)

			// when
			_ = evtesting.RunNatsServerOnPort(
				evtesting.WithPort(testEnvironment.natsPort),
				evtesting.WithJetStreamEnabled())

			// then
			if tc.givenMaxReconnects > 0 {
				require.Eventually(t, func() bool {
					if tc.givenEnableCRDVersion {
						return testEnvironment.jsBackendv2.Conn.IsConnected()
					}
					return testEnvironment.jsBackend.conn.IsConnected()
				}, 30*time.Second, 2*time.Second)
			}

			_, err = testEnvironment.jsClient.StreamInfo(defaultStreamName)
			if tc.givenStorageType == StorageTypeMemory && tc.givenMaxReconnects == 0 {
				// for memory storage with reconnects disabled
				require.True(t, errors.Is(err, nats.ErrStreamNotFound))
			} else {
				// check that the stream is still present for file storage
				// or recreated via reconnect handler for memory storage
				require.NoError(t, err)
			}

			// sync the subscription again to recreate invalid subscriptions or consumers, if any
			if tc.givenEnableCRDVersion {
				err = testEnvironment.jsBackendv2.SyncSubscription(subv2)
			} else {
				err = testEnvironment.jsBackend.SyncSubscription(sub)
			}

			require.NoError(t, err)

			// stream exists
			_, err = testEnvironment.jsClient.StreamInfo(defaultStreamName)
			require.NoError(t, err)

			ev2data := fmt.Sprintf("%s%d", "newsampledata", id)
			if tc.givenEnableCRDVersion {
				require.NoError(t, jetstreamv2.SendEventToJetStream(testEnvironment.jsBackendv2, ev2data))
			} else {
				require.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, ev2data))
			}
			expectedEv2Data := fmt.Sprintf("%q", ev2data)
			require.NoError(t, subscriber.CheckEvent(expectedEv2Data))
		})
	}
}

// TestJetStream_ServerAndSinkRestart tests that the messages persisted (not ack'd) in the stream
// when the sink is down reach the subscriber even when the NATS server is restarted.
func TestJetStream_ServerAndSinkRestart(t *testing.T) {
	for _, newCRD := range []bool{true, false} {
		t.Run(fmt.Sprintf("Enabled New Crd Version: %v", newCRD), func(t *testing.T) {
			// given
			subscriber := evtesting.NewSubscriber()
			defer subscriber.Shutdown()
			require.True(t, subscriber.IsRunning())
			listener := subscriber.GetSubscriberListener()
			listenerNetwork, listenerAddress := listener.Addr().Network(), listener.Addr().String()
			defaultSubsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: defaultMaxInFlights}

			testEnvironment := setupTestEnvironment(t, newCRD)
			defer testEnvironment.natsServer.Shutdown()
			defer testEnvironment.jsClient.natsConn.Close()
			defer func() { _ = testEnvironment.jsClient.DeleteStream(defaultStreamName) }()

			var err error
			if newCRD {
				testEnvironment.jsBackendv2.Config.JSStreamStorageType = StorageTypeFile
				testEnvironment.jsBackendv2.Config.MaxReconnects = 0
				err = testEnvironment.jsBackendv2.Initialize(nil)
			} else {
				testEnvironment.jsBackend.Config.JSStreamStorageType = StorageTypeFile
				testEnvironment.jsBackend.Config.MaxReconnects = 0
				err = testEnvironment.jsBackend.Initialize(nil)
			}
			require.NoError(t, err)

			var sub *eventingv1alpha1.Subscription
			var subv2 *eventingv1alpha2.Subscription
			if newCRD {
				subv2 = evtestingv2.NewSubscription("sub", "foo",
					evtestingv2.WithNotCleanEventSourceAndType(),
					evtestingv2.WithSinkURL(subscriber.SinkURL),
					evtestingv2.WithTypeMatchingStandard(),
					evtestingv2.WithMaxInFlight(defaultMaxInFlights),
				)
				require.NoError(t, jetstreamv2.AddJSCleanEventTypesToStatus(subv2, testEnvironment.cleanerv2))

				// when
				err = testEnvironment.jsBackendv2.SyncSubscription(subv2)
			} else {
				sub = evtesting.NewSubscription("sub", "foo",
					evtesting.WithNotCleanFilter(),
					evtesting.WithSinkURL(subscriber.SinkURL),
					evtesting.WithStatusConfig(defaultSubsConfig),
				)
				require.NoError(t, addJSCleanEventTypesToStatus(sub, testEnvironment.cleaner))

				// when
				err = testEnvironment.jsBackend.SyncSubscription(sub)
			}

			// then
			require.NoError(t, err)
			ev1data := "sampledata"
			if newCRD {
				require.NoError(t, jetstreamv2.SendEventToJetStream(testEnvironment.jsBackendv2, ev1data))
			} else {
				require.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, ev1data))
			}
			expectedEv1Data := fmt.Sprintf("%q", ev1data)
			require.NoError(t, subscriber.CheckEvent(expectedEv1Data))

			// given
			subscriber.Shutdown() // shutdown the subscriber intentionally here
			require.False(t, subscriber.IsRunning())
			ev2data := "newsampletestdata"
			if newCRD {
				require.NoError(t, jetstreamv2.SendEventToJetStream(testEnvironment.jsBackendv2, ev2data))
			} else {
				require.NoError(t, SendEventToJetStream(testEnvironment.jsBackend, ev2data))
			}

			// check that the stream contains one message that was not acknowledged
			const expectedNotAcknowledgedMsgs = uint64(1)
			var info *nats.StreamInfo

			require.Eventually(t, func() bool {
				info, err = testEnvironment.jsClient.StreamInfo(defaultStreamName)
				require.NoError(t, err)
				return info.State.Msgs == expectedNotAcknowledgedMsgs
			}, 60*time.Second, 5*time.Second)

			// shutdown the nats server
			testEnvironment.natsServer.Shutdown()
			require.Eventually(t, func() bool {
				if newCRD {
					return !testEnvironment.jsBackendv2.Conn.IsConnected()
				}
				return !testEnvironment.jsBackend.conn.IsConnected()
			}, 30*time.Second, 2*time.Second)

			// when
			// restart the NATS server
			_ = evtesting.RunNatsServerOnPort(
				evtesting.WithPort(testEnvironment.natsPort),
				evtesting.WithJetStreamEnabled())
			// the unacknowledged message must still be present in the stream
			require.Eventually(t, func() bool {
				info, err = testEnvironment.jsClient.StreamInfo(defaultStreamName)
				require.NoError(t, err)
				return info.State.Msgs == expectedNotAcknowledgedMsgs
			}, 60*time.Second, 5*time.Second)
			// sync the subscription again to recreate invalid subscriptions or consumers, if any
			if newCRD {
				err = testEnvironment.jsBackendv2.SyncSubscription(subv2)
			} else {
				err = testEnvironment.jsBackend.SyncSubscription(sub)
			}
			require.NoError(t, err)
			// restart the subscriber
			listener, err = net.Listen(listenerNetwork, listenerAddress)
			require.NoError(t, err)
			newSubscriber := evtesting.NewSubscriber(evtesting.WithListener(listener))
			defer newSubscriber.Shutdown()
			require.True(t, newSubscriber.IsRunning())

			// then
			// no messages should be present in the stream
			require.Eventually(t, func() bool {
				info, err = testEnvironment.jsClient.StreamInfo(defaultStreamName)
				require.NoError(t, err)
				return info.State.Msgs == uint64(0)
			}, 60*time.Second, 5*time.Second)
			// check if the event is received
			expectedEv2Data := fmt.Sprintf("%q", ev2data)
			require.NoError(t, newSubscriber.CheckEvent(expectedEv2Data))
		})
	}
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

	var jsBackend *JetStream
	var jsBackendNew *jetstreamv2.JetStream
	if newCRD {
		jsBackendNew = jetstreamv2.NewJetStream(natsConfig, metricsCollector, cleanerv2, defaultLogger)
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
