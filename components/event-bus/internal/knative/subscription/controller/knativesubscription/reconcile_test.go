package knativesubscription

import (
	"context"
	"fmt"
	"testing"
	"time"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	fakeeventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/testing"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	kneventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	reconcilertesting "knative.dev/pkg/reconciler/testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
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

var testCases = reconcilertesting.TableTest{
	{
		Name: "New Knative Subscription adds a finalizer",
		Objects: []runtime.Object{
			makeNewKnSubscription(testNamespace, knSubName),
			makeKnSubActivatedKymaSubscription(subName),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: addKnSubFinalizer(makeNewKnSubscription(testNamespace, knSubName), finalizerName),
			},
		},
	},
	{
		Name: "Marked to be deleted Knative Subscription removes finalizer",
		Objects: []runtime.Object{
			makeKnSubActivatedKymaSubscription(subName),
			markedToBeDeletedKnSub(
				addKnSubFinalizer(
					makeNewKnSubscription(testNamespace, knSubName), finalizerName)),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubDeactivatedKymaSubscription(subName),
			},
			{
				Object: markedToBeDeletedKnSub(makeNewKnSubscription(testNamespace, knSubName)),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubDeactivatedKymaSubscription(subName),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, knativeSubscriptionReconciled, "KnativeSubscription reconciled, name: %q; namespace: %q", knSubName, testNamespace),
		},
	},
	{
		Name: "New Knative Subscription will activate Kyma subscription",
		Objects: []runtime.Object{
			makeKnSubDeactivatedKymaSubscription(subName),
			addKnSubFinalizer(
				makeNewKnSubscription(testNamespace, knSubName), finalizerName),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubActivatedKymaSubscription(subName),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubActivatedKymaSubscription(subName),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, knativeSubscriptionReconciled, "KnativeSubscription reconciled, name: %q; namespace: %q", knSubName, testNamespace),
		},
	},
	{
		Name: "Marked to be deleted Knative Subscription will deactivate Kyma Subscription",
		Objects: []runtime.Object{
			makeKnSubActivatedKymaSubscription(subName),
			markedToBeDeletedKnSub(
				addKnSubFinalizer(
					makeNewKnSubscription(testNamespace, knSubName), finalizerName)),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubDeactivatedKymaSubscription(subName),
			},
			{
				Object: markedToBeDeletedKnSub(makeNewKnSubscription(testNamespace, knSubName)),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubDeactivatedKymaSubscription(subName),
			},
		},
		WantEvents: []string{
			reconcilertesting.Eventf(corev1.EventTypeNormal, knativeSubscriptionReconciled, "KnativeSubscription reconciled, name: %q; namespace: %q", knSubName, testNamespace),
		},
	},
}

func TestAllCases(t *testing.T) {
	var ctor Ctor = func(ctx context.Context, ls *Listers) controller.Reconciler {
		rb := reconciler.NewBase(ctx, controllerAgentName, configmap.NewStaticWatcher())
		r := &Reconciler{
			Base:               rb,
			subscriptionLister: ls.GetKnativeSubscriptionLister(),
			kymaEventingClient: fakeeventbusclient.Get(ctx).EventingV1alpha1(),
			time:               NewMockCurrentTime(),
		}

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}

func makeNewKnSubscription(namespace string, name string) *messagingv1alpha1.Subscription {
	subSpec := messagingv1alpha1.SubscriptionSpec{}
	return &messagingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kneventingv1alpha1.SchemeGroupVersion.String(),
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
		Status: messagingv1alpha1.SubscriptionStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				}},
			},
		},
	}
}

func makeKnSubActivatedKymaSubscription(name string) *kymaeventingv1alpha1.Subscription {
	subscription := makeKymaSubscription(name)
	subscription.Status.Conditions = []kymaeventingv1alpha1.SubscriptionCondition{{
		Type:   kymaeventingv1alpha1.SubscriptionReady,
		Status: kymaeventingv1alpha1.ConditionTrue,
	}}
	return subscription
}

func makeKnSubDeactivatedKymaSubscription(name string) *kymaeventingv1alpha1.Subscription {
	subscription := makeKymaSubscription(name)
	subscription.Status.Conditions = []kymaeventingv1alpha1.SubscriptionCondition{{
		Type:   kymaeventingv1alpha1.SubscriptionReady,
		Status: kymaeventingv1alpha1.ConditionFalse,
	}}
	return subscription
}

func makeKymaSubscription(name string) *kymaeventingv1alpha1.Subscription {
	return &kymaeventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kymaeventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
			UID:       subUID,
		},
		SubscriptionSpec: kymaeventingv1alpha1.SubscriptionSpec{
			EventType:        eventType,
			EventTypeVersion: eventTypeVersion,
			SourceID:         sourceID,
		},
	}
}

func addKnSubFinalizer(knSub *messagingv1alpha1.Subscription, finalizer string) *messagingv1alpha1.Subscription {
	knSub.ObjectMeta.Finalizers = append(knSub.ObjectMeta.Finalizers, finalizer)
	return knSub
}

func markedToBeDeletedKnSub(knSub *messagingv1alpha1.Subscription) *messagingv1alpha1.Subscription {
	deletedTime := metav1.Now().Rfc3339Copy()
	knSub.DeletionTimestamp = &deletedTime
	return knSub
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
