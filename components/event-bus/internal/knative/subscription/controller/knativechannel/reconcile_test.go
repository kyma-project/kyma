package knativechannel

import (
	"fmt"
	"time"

	"knative.dev/pkg/apis"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	controllertesting "github.com/knative/eventing/pkg/reconciler/testing"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	//	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
)

const (
	subUID           = "test-uid"
	subName          = "my-sub-1"
	chName           = "my-kn-channel"
	sourceID         = "my-source-ID"
	eventType        = "my-event-type"
	eventTypeVersion = "my-event-type-version"

	testNamespace = "default"
)

func init() {
	err := subApis.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add Kyma eventing scheme")
	}
	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add Kyma application connector scheme")
	}
	err = messagingV1Alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add Knative messaging scheme")
	}
}

var testCases = []controllertesting.TestCase{
	{
		Name: "New channel adds finalizer",
		InitialState: []runtime.Object{
			makeNewChannel(chName, testNamespace),
			makeChannelActivatedSubscription(subName),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, chName),
		WantResult:   reconcile.Result{Requeue: true},
		WantPresent: []runtime.Object{
			addChannelFinalizer(makeNewChannel(chName, testNamespace), finalizerName),
		},
	}, {
		Name: "Marked to be deleted channel removes finalizer",
		InitialState: []runtime.Object{
			markedToBeDeletedChannel(addChannelFinalizer(makeNewChannel(chName, testNamespace), finalizerName)),
			makeChannelActivatedSubscription(subName),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, chName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			markedToBeDeletedChannel(
				makeNewChannel(chName, testNamespace)),
		},
	}, {
		Name: "New channel will activate subscription",
		InitialState: []runtime.Object{
			makeChannelDeactivatedSubscription(subName),
			makeChannelReady(addChannelFinalizer(
				makeNewChannel(chName, testNamespace), finalizerName)),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, chName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			makeChannelActivatedSubscription(subName),
		},
	}, {
		Name: "Marked to be deleted channel will deactivate subscription",
		InitialState: []runtime.Object{
			makeChannelActivatedSubscription(subName),
			markedToBeDeletedChannel(
				addChannelFinalizer(
					makeNewChannel(chName, testNamespace), finalizerName)),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, chName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			makeChannelDeactivatedSubscription(subName),
		},
	},
}

func TestAllCases(t *testing.T) {
	recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	for _, tc := range testCases {
		c := tc.GetClient()
		r := &reconciler{
			client:   c,
			recorder: recorder,
			time:     NewMockCurrentTime(),
		}
		tc.IgnoreTimes = true
		t.Logf("Running test %s", tc.Name)
		t.Run(tc.Name, tc.Runner(t, r, c, tc.GetEventRecorder()))
	}
}

func TestInjectClient(t *testing.T) {
	println("TestInjectClient()")
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

func makeChannelActivatedSubscription(name string) *subApis.Subscription {
	subscription := makeSubscription(name)
	subscription.Status.Conditions = []subApis.SubscriptionCondition{{
		Type:   subApis.EventsActivated,
		Status: subApis.ConditionTrue,
	}, {
		Type:   subApis.ChannelReady,
		Status: subApis.ConditionTrue,
	}, {
		Type:   subApis.SubscriptionReady,
		Status: subApis.ConditionTrue,
	}}
	return subscription
}

func makeChannelDeactivatedSubscription(name string) *subApis.Subscription {
	subscription := makeSubscription(name)
	subscription.Status.Conditions = []subApis.SubscriptionCondition{{
		Type:   subApis.EventsActivated,
		Status: subApis.ConditionTrue,
	}, {
		Type:   subApis.ChannelReady,
		Status: subApis.ConditionFalse,
	}, {
		Type:   subApis.SubscriptionReady,
		Status: subApis.ConditionTrue}}
	return subscription
}

func makeSubscription(name string) *subApis.Subscription {
	return &subApis.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: subApis.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
			UID:       subUID,
		},
		SubscriptionSpec: subApis.SubscriptionSpec{
			EventType:        eventType,
			EventTypeVersion: eventTypeVersion,
			SourceID:         sourceID,
		},
	}
}

func makeNewChannel(name string, ns string) *messagingV1Alpha1.Channel {
	return &messagingV1Alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: messagingV1Alpha1.SchemeGroupVersion.String(),
			Kind:       "Channel",
		}, ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"kyma-event-type":         eventType,
				"kyma-event-type-version": eventTypeVersion,
				"kyma-ns":                 ns,
				"kyma-source-id":          sourceID,
			},
		},
	}
}

func addChannelFinalizer(ch *messagingV1Alpha1.Channel, finalizer string) *messagingV1Alpha1.Channel {
	ch.ObjectMeta.Finalizers = append(ch.ObjectMeta.Finalizers, finalizer)
	return ch
}

func makeChannelReady(ch *messagingV1Alpha1.Channel) *messagingV1Alpha1.Channel {
	ch.Status.Conditions = []apis.Condition{{
		Type:   apis.ConditionReady,
		Status: corev1.ConditionTrue,
	}}
	return ch
}

func markedToBeDeletedChannel(ch *messagingV1Alpha1.Channel) *messagingV1Alpha1.Channel {
	deletedTime := metav1.Now().Rfc3339Copy()
	ch.DeletionTimestamp = &deletedTime
	return ch
}

type MockCurrentTime struct{}

func NewMockCurrentTime() util.CurrentTime {
	mockCurrentTime := new(MockCurrentTime)
	return mockCurrentTime
}

func (m *MockCurrentTime) GetCurrentTime() metav1.Time {
	return metav1.NewTime(time.Time{})
}
