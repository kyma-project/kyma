package knativesubscription

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"knative.dev/pkg/apis"

	controllertesting "github.com/knative/eventing/pkg/reconciler/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	//	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
	"testing"

	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
)

const (
	subUID           = "test-uid"
	subName          = "my-sub"
	knSubName        = "my-kn-sub"
	sourceID         = "my-source-id"
	eventType        = "my-event-type"
	eventTypeVersion = "my-event-type-version"
	testNamespace    = "default"
)

func init() {
	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add Application Connector scheme")
	}
	err = evapisv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Error(err, "Failed to add Knative Eventing scheme")
	}
}

var testCases = []controllertesting.TestCase{
	{
		Name: "New Knative Channel adds finalizer",
		InitialState: []runtime.Object{
			makeNewKnSubscription(testNamespace, knSubName),
			makeKnSubActivatedSubscription(subName),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantResult:   reconcile.Result{Requeue: true},
		WantPresent: []runtime.Object{
			addKnSubFinalizer(
				makeNewKnSubscription(testNamespace, knSubName), finalizerName),
		},
	},
	{
		Name: "Marked to be deleted Knative Subscription removes finalizer",
		InitialState: []runtime.Object{
			makeKnSubActivatedSubscription(subName),
			markedToBeDeletedKnSub(
				addKnSubFinalizer(
					makeNewKnSubscription(testNamespace, knSubName), finalizerName)),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			markedToBeDeletedKnSub(
				makeNewKnSubscription(testNamespace, knSubName)),
		},
	},
	{
		Name: "New Knative Subscription will activate Kyma subscription",
		InitialState: []runtime.Object{
			makeKnSubDeactivatedSubscription(subName),
			addKnSubFinalizer(
				makeNewKnSubscription(testNamespace, knSubName), finalizerName),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			makeKnSubActivatedSubscription(subName),
		},
	},
	{
		Name: "Marked to be deleted Knative Subscription will deactivate Kyma Subscription",
		InitialState: []runtime.Object{
			makeKnSubActivatedSubscription(subName),
			markedToBeDeletedKnSub(
				addKnSubFinalizer(
					makeNewKnSubscription(testNamespace, knSubName), finalizerName)),
		},
		ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantResult:   reconcile.Result{},
		WantPresent: []runtime.Object{
			makeKnSubDeactivatedSubscription(subName),
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

func makeNewKnSubscription(namespace string, name string) *evapisv1alpha1.Subscription {
	subSpec := evapisv1alpha1.SubscriptionSpec{}
	return &evapisv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: evapisv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				"kyma-event-type":         eventType,
				"kyma-event-type-version": eventTypeVersion,
				"kyma-ns":                 namespace,
				"kyma-source-id":          sourceID,
			},
		},
		Spec: subSpec,
		Status: evapisv1alpha1.SubscriptionStatus{
			Status: duckv1beta1.Status{
				Conditions: []apis.Condition{{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
	}
}

func makeKnSubActivatedSubscription(name string) *subApis.Subscription {
	subscription := makeSubscription(name)
	subscription.Status.Conditions = []subApis.SubscriptionCondition{{
		Type:   subApis.SubscriptionReady,
		Status: subApis.ConditionTrue,
	}}
	return subscription
}

func makeKnSubDeactivatedSubscription(name string) *subApis.Subscription {
	subscription := makeSubscription(name)
	subscription.Status.Conditions = []subApis.SubscriptionCondition{{
		Type:   subApis.SubscriptionReady,
		Status: subApis.ConditionFalse,
	}}
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

func addKnSubFinalizer(knSub *evapisv1alpha1.Subscription, finalizer string) *evapisv1alpha1.Subscription {
	knSub.ObjectMeta.Finalizers = append(knSub.ObjectMeta.Finalizers, finalizer)
	return knSub
}

func markedToBeDeletedKnSub(knSub *evapisv1alpha1.Subscription) *evapisv1alpha1.Subscription {
	deletedTime := metav1.Now().Rfc3339Copy()
	knSub.DeletionTimestamp = &deletedTime
	return knSub
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
