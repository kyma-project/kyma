package eventactivation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	controllertesting "knative.dev/pkg/reconciler/testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	fakeeventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/testing"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

const (
	subUID   = "test-uid"
	subName  = "my-sub"
	eaName   = "my-event-activation"
	sourceID = "my_source_ID"

	testNamespace = "default"
)

var testCases = controllertesting.TableTest{
	{
		Name: "New event activation adds finalizer",
		Objects: []runtime.Object{
			makeNewEventActivation(testNamespace, eaName),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, eaName),
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: addEventActivationFinalizer(makeNewEventActivation(testNamespace, eaName), finalizerName),
		}},
	},
	{
		Name: "Marked to be deleted event activation removes finalizer",
		Objects: []runtime.Object{
			markedToBeDeletedEventActivation(
				addEventActivationFinalizer(
					makeNewEventActivation(testNamespace, eaName), finalizerName)),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, eaName),
		WantUpdates: []clientgotesting.UpdateActionImpl{{
			Object: markedToBeDeletedEventActivation(makeNewEventActivation(testNamespace, eaName)),
		}},
		WantEvents: []string{
			controllertesting.Eventf(corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", eaName, testNamespace),
		},
	},
	{
		Name: "New event activation will activate subscription",
		Objects: []runtime.Object{
			makeEventsDeactivatedSubscription(subName),
			addEventActivationFinalizer(
				makeNewEventActivation(testNamespace, eaName), finalizerName),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, eaName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeEventsActivatedSubscription(subName),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeEventsActivatedSubscription(subName),
			},
		},
		WantEvents: []string{
			controllertesting.Eventf(corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", eaName, testNamespace),
		},
	},
	{
		Name: "Marked to be deleted event activation will deactivate subscription",
		Objects: []runtime.Object{
			makeEventsActivatedSubscription(subName),
			markedToBeDeletedEventActivation(
				addEventActivationFinalizer(
					makeNewEventActivation(testNamespace, eaName), finalizerName)),
		},
		Key: fmt.Sprintf("%s/%s", testNamespace, eaName),
		WantUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeEventsDeactivatedSubscription(subName),
			},
			{
				Object: markedToBeDeletedEventActivation(makeNewEventActivation(testNamespace, eaName)),
			},
		},
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{
			{
				Object: makeEventsDeactivatedSubscription(subName),
			},
		},
		WantEvents: []string{
			controllertesting.Eventf(corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", eaName, testNamespace),
		},
	},
}

func TestAllCases(t *testing.T) {
	var ctor Ctor = func(ctx context.Context, ls *Listers) controller.Reconciler {
		rb := reconciler.NewBase(ctx, controllerAgentName, configmap.NewStaticWatcher())
		r := &Reconciler{
			Base:                       rb,
			eventActivationLister:      ls.GetEventActivationLister(),
			applicationconnectorClient: fakeeventbusclient.Get(ctx).ApplicationconnectorV1alpha1(),
			kymaEventingClient:         fakeeventbusclient.Get(ctx).EventingV1alpha1(),
			time:                       NewMockCurrentTime(),
		}

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}

func makeNewEventActivation(namespace string, name string) *eventingv1alpha1.EventActivation {
	eas := eventingv1alpha1.EventActivationSpec{
		DisplayName: "display_name",
		SourceID:    sourceID,
	}
	return &eventingv1alpha1.EventActivation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta:          om(namespace, name),
		EventActivationSpec: eas,
	}
}

func makeEventsActivatedSubscription(name string) *subApis.Subscription {
	subscription := makeSubscription(name)
	subscription.Status.Conditions = []subApis.SubscriptionCondition{{
		Type:   subApis.EventsActivated,
		Status: subApis.ConditionTrue,
	}}
	return subscription
}

func makeEventsDeactivatedSubscription(name string) *subApis.Subscription {
	subscription := makeSubscription(name)
	subscription.Status.Conditions = []subApis.SubscriptionCondition{{
		Type:   subApis.EventsActivated,
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
			SourceID: sourceID,
		},
	}
}

func addEventActivationFinalizer(ea *eventingv1alpha1.EventActivation, finalizer string) *eventingv1alpha1.EventActivation {
	ea.ObjectMeta.Finalizers = append(ea.ObjectMeta.Finalizers, finalizer)
	return ea
}

func markedToBeDeletedEventActivation(ea *eventingv1alpha1.EventActivation) *eventingv1alpha1.EventActivation {
	deletedTime := metav1.Now().Rfc3339Copy()
	ea.DeletionTimestamp = &deletedTime
	return ea
}

func om(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
		SelfLink:  fmt.Sprintf("/apis/eventing/v1alpha1/namespaces/%s/object/%s", namespace, name),
	}
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
