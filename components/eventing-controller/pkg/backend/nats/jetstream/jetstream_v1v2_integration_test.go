package jetstream

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	evtestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
)

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
				jetstreamv2.AddJSCleanEventTypesToStatus(subv2, testEnvironment.cleanerv2)

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
				jetstreamv2.AddJSCleanEventTypesToStatus(subv2, testEnvironment.cleanerv2)

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
