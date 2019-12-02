package subscription

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	messagingv1alpha1Client "github.com/knative/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"

	controllertesting "github.com/knative/eventing/pkg/reconciler/testing"

	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	eventingV1Alpha1Client "github.com/knative/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"

	messagingv1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	kySubName        = "ky-subscription"
	kyNamespace      = "default"
	eventType        = "testevent"
	eventTypeVersion = "v1"
	sourceID         = "testsourceid"

	subUID        = "sub-uid"
	chanUID       = "channel-uid"
	provisioner   = "natss"
	subscriberURI = "URL-test-susbscriber"

	testErrorMessage        = "test induced error"
	defaultMaxChannelLength = 25
)

var (
	// deletionTime is used when objects are marked as deleted. Rfc3339Copy()
	// truncates to seconds to match the loss of precision during serialization.
	deletionTime = metav1.Now().Rfc3339Copy()

	events = map[string]corev1.Event{
		subReconciled:      {Reason: subReconciled, Type: corev1.EventTypeNormal},
		subReconcileFailed: {Reason: subReconcileFailed, Type: corev1.EventTypeWarning},
	}

	knativeLib = NewMockKnativeLib()

	labels = map[string]string{
		"kyma-event-type":         "testevent",
		"kyma-event-type-version": "v1",
		"kyma-source-id":          "testsourceid",
	}
)

func init() {
	// Add types to scheme
	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add kyma eventing scheme")
	}
	err = evapisv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add knative eventing scheme")
	}
	err = messagingV1Alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add knative messaging scheme")
	}
}

func TestInjectClient(t *testing.T) {
	r := &reconciler{}
	orig := r.client
	n := fake.NewFakeClient()
	if orig == n {
		t.Errorf("Original and new clients are identical: %v", orig)
	}
	err := r.InjectClient(n)
	if err != nil {
		t.Errorf("Unexpected error injecting the client: %v", err)
	}
	if n != r.client {
		t.Errorf("Unexpected client. Expected: '%v'. Actual: '%v'", n, r.client)
	}
}

var testCases = []controllertesting.TestCase{
	{
		Name: "Subscription not found",
	},
	{
		Name: "Error getting Subscription",
		Mocks: controllertesting.Mocks{
			MockGets: errorGettingSubscription(),
		},
		WantErrMsg: testErrorMessage,
	},
	{
		Name: "New kyma subscription adds finalizer",
		InitialState: []runtime.Object{
			makeKySubscription(),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantResult:   reconcile.Result{Requeue: true},
		WantPresent: []runtime.Object{
			makeSubscriptionWithFinalizer(),
		},
	},
	{
		Name: "Activated kyma subscription doesn't create a new channel if it exists, but will create a new kn subscription",
		InitialState: []runtime.Object{
			makeEventsActivatedSubscription(),
			makeKnativeLibChannel(),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantResult:   reconcile.Result{Requeue: false},
		AdditionalVerification: []func(*testing.T, *controllertesting.TestCase){
			func(t *testing.T, tc *controllertesting.TestCase) {
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
		WantPresent: []runtime.Object{
			makeReadySubscription(),
		},
		WantEvent: []corev1.Event{
			events[subReconciled],
		},
	},
	{
		Name: "Activated kyma subscription creates a new channel and a new knative subscription",
		InitialState: []runtime.Object{
			makeEventsActivatedSubscription(),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantResult:   reconcile.Result{Requeue: false},
		AdditionalVerification: []func(*testing.T, *controllertesting.TestCase){
			func(t *testing.T, tc *controllertesting.TestCase) {
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
		WantPresent: []runtime.Object{
			makeReadySubscription(),
		},
		WantEvent: []corev1.Event{
			events[subReconciled],
		},
	},
	{
		Name: "Deactivated kyma subscription deletes kn subscription and the channel",
		InitialState: []runtime.Object{
			makeEventsDeactivatedSubscription(),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantResult:   reconcile.Result{Requeue: false},
		AdditionalVerification: []func(*testing.T, *controllertesting.TestCase){
			func(t *testing.T, tc *controllertesting.TestCase) {
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
		WantPresent: []runtime.Object{
			makeNotReadySubscription(),
		},
		WantEvent: []corev1.Event{
			events[subReconciled],
		},
	},
	{
		Name: "Marked to be deleted kyma subscription remove finalizer",
		InitialState: []runtime.Object{
			makeDeletingSubscriptionWithFinalizer(),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", kyNamespace, kySubName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			makeDeletingSubscription(),
		},
		WantEvent: []corev1.Event{
			events[subReconciled],
		},
	},
}

func TestAllCases(t *testing.T) {
	//recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	for _, tc := range testCases {
		c := tc.GetClient()
		options := opts.Options{
			Port:           8080,
			ResyncPeriod:   10 * time.Second,
			ChannelTimeout: 10 * time.Second,
		}

		recorder := tc.GetEventRecorder()
		r := &reconciler{
			client:     c,
			recorder:   recorder,
			knativeLib: knativeLib,
			opts:       &options,
			time:       NewMockCurrentTime(),
		}
		t.Logf("Running test %s", tc.Name)
		if tc.ReconcileKey == "" {
			tc.ReconcileKey = fmt.Sprintf("/%s", kySubName)
		}
		tc.IgnoreTimes = true
		t.Run(tc.Name, tc.Runner(t, r, c, recorder))
	}
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

func errorGettingSubscription() []controllertesting.MockGet {
	return []controllertesting.MockGet{
		func(_ client.Client, _ context.Context, _ client.ObjectKey, obj runtime.Object) (controllertesting.MockHandled, error) {
			if _, ok := obj.(*eventingv1alpha1.Subscription); ok {
				return controllertesting.Handled, errors.New(testErrorMessage)
			}
			return controllertesting.Unhandled, nil
		},
	}
}

// Mock the current time for Status "LastTranscationTime"
type MockCurrentTime struct{}

func NewMockCurrentTime() util.CurrentTime {
	mockCurrentTime := new(MockCurrentTime)
	return mockCurrentTime
}

func (m *MockCurrentTime) GetCurrentTime() metav1.Time {
	return metav1.NewTime(time.Time{})
}

// Mock KnativeLib
var knSubscriptions = make(map[string]*evapisv1alpha1.Subscription)
var knChannels = make(map[string]*messagingV1Alpha1.Channel)
var knChannelNames = make(map[string]string)

type MockKnativeLib struct{}

func NewMockKnativeLib() util.KnativeAccessLib {
	return new(MockKnativeLib)
}
func (k *MockKnativeLib) GetChannel(name string, namespace string) (*messagingV1Alpha1.Channel, error) {
	channel, ok := knChannels[name]
	if !ok {
		gr := schema.GroupResource{Group: "test", Resource: "channel"}
		return nil, apierrors.NewNotFound(gr, name)
	}
	return channel, nil
}

func (k *MockKnativeLib) GetChannelByLabels(namespace string, labels map[string]string) (*messagingV1Alpha1.Channel, error) {
	var channelName string
	for name := range knChannels {
		channelName = name
		break
	}
	return k.GetChannel(channelName, namespace)
}

func (k *MockKnativeLib) CreateChannel(prefix, namespace string, labels map[string]string,
	readyFn ...util.ChannelReadyFunc) (*messagingV1Alpha1.Channel, error) {

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
func (k *MockKnativeLib) GetSubscription(name string, namespace string) (*evapisv1alpha1.Subscription, error) {
	knSub, ok := knSubscriptions[name]
	if !ok {
		gr := schema.GroupResource{Group: "test", Resource: "kn-subscriptoin"}
		return nil, apierrors.NewNotFound(gr, name)
	}
	return knSub, nil
}
func (k *MockKnativeLib) UpdateSubscription(sub *evapisv1alpha1.Subscription) (*evapisv1alpha1.Subscription, error) {
	return nil, nil
}
func (k *MockKnativeLib) SendMessage(channel *messagingV1Alpha1.Channel, headers *map[string][]string, message *string) error {
	return nil
}

func (k *MockKnativeLib) MsgChannelClient() messagingv1alpha1Client.MessagingV1alpha1Interface {
	return nil
}

// InjectClient injects a client, useful for running tests.
func (k *MockKnativeLib) InjectClient(evClient eventingV1Alpha1Client.EventingV1alpha1Interface, msgClient messagingv1alpha1.MessagingV1alpha1Interface) error {
	return nil
}

//  make channels
func makeKnChannel(prefix, namespace string, labels map[string]string) *messagingV1Alpha1.Channel {
	channelName := fmt.Sprint(prefix, "-", "23vwq3")
	knChannelNames[prefix] = channelName
	return &messagingV1Alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			Name:         channelName,
			GenerateName: prefix,
			Labels:       labels,
			UID:          chanUID,
		},
		Status: messagingV1Alpha1.ChannelStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
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

func makeKnativeLibChannel() *messagingV1Alpha1.Channel {
	chNamespace := util.GetDefaultChannelNamespace()
	channel, _ := knativeLib.CreateChannel(makeKnChannelNamePrefix(makeEventsActivatedSubscription()), chNamespace, labels)
	channel.SetClusterName("fake-channel") // use it as a marker
	knChannels[channel.Name] = channel
	return channel
}

func makeKnSubscription(kySub *eventingv1alpha1.Subscription) *evapisv1alpha1.Subscription {
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
