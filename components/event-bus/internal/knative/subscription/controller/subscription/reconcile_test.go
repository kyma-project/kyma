package subscription

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	fakeeventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/testing"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	eventingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingClientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	reconcilertesting "knative.dev/pkg/reconciler/testing"
)

const (
	kySubName        = "ky-subscription"
	kyNamespace      = "default"
	eventType        = "testevent"
	eventTypeVersion = "v1"
	sourceID         = "testsourceid"

	subUID        = "sub-uid"
	chanUID       = "channel-uid"
	subscriberURI = "URL-test-subscriber"
)

var (
	// deletionTime is used when objects are marked as deleted. Rfc3339Copy()
	// truncates to seconds to match the loss of precision during serialization.
	deletionTime = metav1.Now().Rfc3339Copy()
	events       = map[string]corev1.Event{
		subReconciled:      {Reason: subReconciled, Type: corev1.EventTypeNormal},
		subReconcileFailed: {Reason: subReconcileFailed, Type: corev1.EventTypeWarning},
	}
	knativeLib = NewMockKnativeLib()
	labels     = map[string]string{
		"kyma-event-type":         "testevent",
		"kyma-event-type-version": "v1",
		"kyma-source-id":          "testsourceid",
	}
)

var testCases = reconcilertesting.TableTest{
	{
		Name:    "Subscription not found",
		Key:     "invalid/invalid",
		WantErr: true,
	},
	{
		Name: "New kyma subscription adds finalizer",
		Objects: []runtime.Object{
			makeKySubscription(),
		},
		Key: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{

				Object: makeSubscriptionWithFinalizer(),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeSubscriptionWithFinalizer(),
			},
		},
	},
	{
		Name: "Activated kyma subscription doesn't create a new channel if it exists, but will create a new kn subscription",
		Objects: []runtime.Object{
			makeEventsActivatedSubscription(),
			makeKnativeLibChannel(),
		},
		Key: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		PostConditions: []func(*testing.T, *reconcilertesting.TableRow){
			func(t *testing.T, tc *reconcilertesting.TableRow) {
				dumpKnativeLibObjects(t)
				if _, ok := knSubscriptions[makeKnSubscriptionName(makeEventsActivatedSubscription())]; !ok {
					t.Errorf("Knative subscription was NOT created")
				}
				channelNamePrefix := makeKnChannelNamePrefix(makeEventsActivatedSubscription())
				channelName := knChannelNames[channelNamePrefix] // Get the channel name from the prefix
				if channel, ok := knChannels[channelName]; ok {
					if channel.GetClusterName() != "fake-channel" {
						t.Errorf("Knative channel should NOT be created in this case")
					}
				}
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeReadySubscription(),
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeReadySubscription(),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, events[subReconciled].Reason, "Subscription reconciled, name: %q; namespace: %q", kySubName, kyNamespace),
		},
	},
	{
		Name: "Activated kyma subscription creates a new channel and a new knative subscription",
		Objects: []runtime.Object{
			makeEventsActivatedSubscription(),
		},
		Key: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		PostConditions: []func(*testing.T, *reconcilertesting.TableRow){
			func(t *testing.T, tc *reconcilertesting.TableRow) {
				dumpKnativeLibObjects(t)
				if _, ok := knSubscriptions[makeKnSubscriptionName(makeEventsActivatedSubscription())]; !ok {
					t.Errorf("Knative subscription was NOT created")
				}
				channelNamePrefix := makeKnChannelNamePrefix(makeEventsActivatedSubscription())
				channelName := knChannelNames[channelNamePrefix] // Get the channel name from the prefix

				if ch, ok := knChannels[channelName]; !ok {
					t.Errorf("Knative channel was NOT created")
				} else {
					chLabels := ch.Labels
					ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
					if diff := cmp.Diff(labels, chLabels, ignore); diff != "" {
						t.Errorf("%s (-want, +got) = %v", "Activated kyma subscription creates a new channel and a new knative subscription", diff)
					}
				}
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeReadySubscription(),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeReadySubscription(),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, events[subReconciled].Reason, "Subscription reconciled, name: %q; namespace: %q", kySubName, kyNamespace),
		},
	},
	{
		Name: "Deactivated kyma subscription deletes kn subscription and the channel",
		Objects: []runtime.Object{
			makeEventsDeactivatedSubscription(),
		},
		Key: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		PostConditions: []func(*testing.T, *reconcilertesting.TableRow){
			func(t *testing.T, tc *reconcilertesting.TableRow) {
				dumpKnativeLibObjects(t)
				if _, ok := knSubscriptions[makeKnSubscriptionName(makeEventsActivatedSubscription())]; ok {
					t.Errorf("Knative subscription was NOT deleted")
				}
				channelNamePrefix := makeKnChannelNamePrefix(makeEventsActivatedSubscription())
				channelName := knChannelNames[channelNamePrefix] // Get the channel name from the prefix

				if _, ok := knChannels[channelName]; ok {
					t.Errorf("Knative channel was NOT deleted")
				}
			},
		},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeNotReadySubscription(),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeNotReadySubscription(),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, events[subReconciled].Reason, "Subscription reconciled, name: %q; namespace: %q", kySubName, kyNamespace),
		},
	},
	{
		Name: "Marked to be deleted kyma subscription remove finalizer",
		Objects: []runtime.Object{
			makeDeletingSubscriptionWithFinalizer(),
		},
		Key: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeDeletingSubscription(),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeDeletingSubscription(),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, events[subReconciled].Reason, "Subscription reconciled and deleted, name: %q; namespace: %q", kySubName, kyNamespace),
		},
	},
}

func TestAllCases(t *testing.T) {
	options := &opts.Options{
		ChannelTimeout: 10 * time.Second,
	}
	var ctor Ctor = func(ctx context.Context, ls *Listers) controller.Reconciler {
		rb := reconciler.NewBase(ctx, controllerAgentName, configmap.NewStaticWatcher())
		r := &Reconciler{
			Base:                  rb,
			subscriptionLister:    ls.GetSubscriptionLister(),
			eventActivationLister: ls.GetEventActivationLister(),
			kymaEventingClient:    fakeeventbusclient.Get(ctx).EventingV1alpha1(),
			knativeLib:            NewMockKnativeLib(),
			opts:                  options,
			time:                  NewMockCurrentTime(),
		}

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}

func makeKySubscription() *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kySubName,
			Namespace: kyNamespace,
			UID:       subUID,
		},
		SubscriptionSpec: eventingv1alpha1.SubscriptionSpec{
			EventType:        eventType,
			EventTypeVersion: eventTypeVersion,
			SourceID:         sourceID,
		},
	}
}

func makeReadySubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscriptionWithFinalizer()
	subscription.Status.Conditions = []eventingv1alpha1.SubscriptionCondition{
		{Type: eventingv1alpha1.EventsActivated, Status: eventingv1alpha1.ConditionTrue},
		{Type: eventingv1alpha1.SubscriptionReady, Status: eventingv1alpha1.ConditionTrue},
		{Type: eventingv1alpha1.Ready, Status: eventingv1alpha1.ConditionTrue},
	}
	return subscription
}

func makeNotReadySubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscriptionWithFinalizer()
	subscription.Status.Conditions = []eventingv1alpha1.SubscriptionCondition{
		{Type: eventingv1alpha1.EventsActivated, Status: eventingv1alpha1.ConditionFalse},
		{Type: eventingv1alpha1.Ready, Status: eventingv1alpha1.ConditionFalse},
	}
	return subscription
}

func makeEventsActivatedSubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscriptionWithFinalizer()
	subscription.Status.Conditions = []eventingv1alpha1.SubscriptionCondition{{
		Type:   eventingv1alpha1.EventsActivated,
		Status: eventingv1alpha1.ConditionTrue,
	}, {
		Type:   eventingv1alpha1.SubscriptionReady,
		Status: eventingv1alpha1.ConditionTrue,
	}}
	return subscription
}

func makeEventsDeactivatedSubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscriptionWithFinalizer()
	subscription.Status.Conditions = []eventingv1alpha1.SubscriptionCondition{{
		Type:   eventingv1alpha1.EventsActivated,
		Status: eventingv1alpha1.ConditionFalse,
	}}
	return subscription
}

func makeSubscriptionWithFinalizer() *eventingv1alpha1.Subscription {
	subscription := makeKySubscription()
	subscription.Finalizers = []string{finalizerName}
	return subscription
}

func makeDeletingSubscription() *eventingv1alpha1.Subscription {
	subscription := makeKySubscription()
	subscription.DeletionTimestamp = &deletionTime
	return subscription
}

func makeDeletingSubscriptionWithFinalizer() *eventingv1alpha1.Subscription {
	subscription := makeSubscriptionWithFinalizer()
	subscription.DeletionTimestamp = &deletionTime
	return subscription
}

// Mock the current time for Status "LastTransactionTime"
type MockCurrentTime struct{}

func NewMockCurrentTime() util.CurrentTime {
	mockCurrentTime := new(MockCurrentTime)
	return mockCurrentTime
}

func (m *MockCurrentTime) GetCurrentTime() metav1.Time {
	return metav1.NewTime(time.Time{})
}

// Mock KnativeLib
var knSubscriptions = make(map[string]*messagingv1alpha1.Subscription)
var knChannels = make(map[string]*messagingv1alpha1.Channel)
var knChannelNames = make(map[string]string)

type MockKnativeLib struct{}

func NewMockKnativeLib() util.KnativeAccessLib {
	return new(MockKnativeLib)
}

func (k *MockKnativeLib) GetChannel(name string, namespace string) (*messagingv1alpha1.Channel, error) {
	channel, ok := knChannels[name]
	if !ok {
		gr := schema.GroupResource{Group: "test", Resource: "channel"}
		return nil, apierrors.NewNotFound(gr, name)
	}
	return channel, nil
}
func (k *MockKnativeLib) GetChannelByLabels(namespace string, labels map[string]string) (*messagingv1alpha1.Channel, error) {
	var channelName string
	for name := range knChannels {
		channelName = name
		break
	}
	return k.GetChannel(channelName, namespace)
}
func (k *MockKnativeLib) CreateChannel(prefix, namespace string, labels map[string]string,
	readyFn ...util.ChannelReadyFunc) (*messagingv1alpha1.Channel, error) {

	channel := makeKnChannel(prefix, namespace, labels)
	knChannels[channel.Name] = channel
	return channel, nil
}
func (k *MockKnativeLib) DeleteChannel(name string, namespace string) error {
	delete(knChannels, name)
	return nil
}
func (k *MockKnativeLib) CreateSubscription(name string, namespace string, channelName string, uri *string, labels map[string]string) error {
	knSub := makeKnSubscription(makeEventsActivatedSubscription())
	knSubscriptions[knSub.Name] = knSub
	return nil
}
func (k *MockKnativeLib) DeleteSubscription(name string, namespace string) error {
	delete(knSubscriptions, name)
	return nil
}
func (k *MockKnativeLib) GetSubscription(name string, namespace string) (*messagingv1alpha1.Subscription, error) {
	knSub, ok := knSubscriptions[name]
	if !ok {
		gr := schema.GroupResource{Group: "test", Resource: "kn-subscriptoin"}
		return nil, apierrors.NewNotFound(gr, name)
	}
	return knSub, nil
}
func (k *MockKnativeLib) UpdateSubscription(sub *messagingv1alpha1.Subscription) (*messagingv1alpha1.Subscription, error) {
	return nil, nil
}
func (k *MockKnativeLib) SendMessage(channel *messagingv1alpha1.Channel, headers *map[string][]string, message *string) error {
	return nil
}

//
// InjectClient injects a client, useful for running tests.
func (k *MockKnativeLib) InjectClient(evClient eventingclientv1alpha1.EventingV1alpha1Interface, msgClient messagingClientv1alpha1.MessagingV1alpha1Interface) error {
	return nil
}

//
//  make channels
func makeKnChannel(prefix, namespace string, labels map[string]string) *messagingv1alpha1.Channel {
	channelName := fmt.Sprint(prefix, "-", "23vwq3")
	knChannelNames[prefix] = channelName
	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			Name:         channelName,
			GenerateName: prefix,
			Labels:       labels,
			UID:          chanUID,
		},
		Status: messagingv1alpha1.ChannelStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					apis.Condition{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
	}
}

func makeKnChannelNamePrefix(kySub *eventingv1alpha1.Subscription) string {
	return kySub.EventType
}

func makeKnSubscriptionName(kySub *eventingv1alpha1.Subscription) string {
	return util.GetKnSubscriptionName(&kySub.Name, &kySub.Namespace)
}

func makeKnativeLibChannel() *messagingv1alpha1.Channel {
	chNamespace := util.GetDefaultChannelNamespace()
	channel, _ := knativeLib.CreateChannel(makeKnChannelNamePrefix(makeEventsActivatedSubscription()), chNamespace, labels)
	channel.SetClusterName("fake-channel") // use it as a marker
	knChannels[channel.Name] = channel
	return channel
}

func makeKnSubscription(kySub *eventingv1alpha1.Subscription) *messagingv1alpha1.Subscription {
	knSubName := util.GetKnSubscriptionName(&kySub.Name, &kySub.Namespace)
	knChannelName := knChannelNames[makeKnChannelNamePrefix(kySub)]
	subscriberURL := subscriberURI
	chNamespace := util.GetDefaultChannelNamespace()
	return util.Subscription(knSubName, chNamespace, labels).ToChannel(knChannelName).ToURI(&subscriberURL).EmptyReply().Build()
}

func dumpKnativeLibObjects(t *testing.T) {
	t.Log("--- Knative Subscriptions ---")
	for key, value := range knSubscriptions {
		t.Logf("key: %v", key)
		t.Logf("subscription: %v", *value)
	}
	t.Log("--- Knative Channels ---")
	for key, value := range knChannels {
		t.Logf("key: %v", key)
		t.Logf("channel: %v", *value)
	}
}
