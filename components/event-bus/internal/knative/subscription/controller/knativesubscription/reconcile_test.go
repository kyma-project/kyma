package knativesubscription

import (
	"context"
	"fmt"
	"testing"
	"time"

	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	fakeeventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/testing"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	evapisv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	controllertesting "knative.dev/pkg/reconciler/testing"

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

var testCases = controllertesting.TableTest{
	{
		Name: "New Knative Channel adds finalizer",
		Objects: []runtime.Object{
			makeNewKnSubscription(testNamespace, knSubName),
			makeKnSubActivatedSubscription(subName),
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
			makeKnSubActivatedSubscription(subName),
			markedToBeDeletedKnSub(
				addKnSubFinalizer(
					makeNewKnSubscription(testNamespace, knSubName), finalizerName)),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		//WantPresent: []runtime.Object{
		//	markedToBeDeletedKnSub(
		//		makeNewKnSubscription(testNamespace, knSubName)),
		//},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: markedToBeDeletedKnSub(makeNewKnSubscription(testNamespace, knSubName)),
			},
		},
	},
	{
		Name: "New Knative Subscription will activate Kyma subscription",
		Objects: []runtime.Object{
			makeKnSubDeactivatedSubscription(subName),
			addKnSubFinalizer(
				makeNewKnSubscription(testNamespace, knSubName), finalizerName),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		//WantPresent: []runtime.Object{
		//	makeKnSubActivatedSubscription(subName),
		//},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubActivatedSubscription(subName),
			},
		},
	},
	{
		Name: "Marked to be deleted Knative Subscription will deactivate Kyma Subscription",
		Objects: []runtime.Object{
			makeKnSubActivatedSubscription(subName),
			markedToBeDeletedKnSub(
				addKnSubFinalizer(
					makeNewKnSubscription(testNamespace, knSubName), finalizerName)),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, knSubName),
		//WantPresent: []runtime.Object{
		//	makeKnSubDeactivatedSubscription(subName),
		//},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeKnSubDeactivatedSubscription(subName),
			},
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
