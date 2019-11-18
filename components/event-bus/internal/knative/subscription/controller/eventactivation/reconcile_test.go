package eventactivation

import (
	"context"
	"fmt"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"time"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	controllertesting "knative.dev/pkg/reconciler/testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	fakeeventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client/fake"

	"testing"

	. "github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/testing"
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
		//Mocks: controllertesting.Mocks{
		//	MockLists: mockSubscriptionEmptyList(),
		//},
		Key:        fmt.Sprintf("%s/%s", testNamespace, eaName),
		//WantResult: reconcile.Result{Requeue: true},
		//WantPresent: []runtime.Object{
		//	addEventActivationFinalizer(
		//		makeNewEventActivation(testNamespace, eaName), finalizerName),
		//},
	},
	//{
	//	Name: "Marked to be deleted event activation removes finalizer",
	//	InitialState: []runtime.Object{
	//		markedToBeDeletedEventActivation(
	//			addEventActivationFinalizer(
	//				makeNewEventActivation(testNamespace, eaName), finalizerName)),
	//	},
	//	Mocks: controllertesting.Mocks{
	//		MockLists: mockSubscriptionEmptyList(),
	//	},
	//	ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, eaName),
	//	WantResult:   reconcile.Result{},
	//	WantPresent: []runtime.Object{
	//		markedToBeDeletedEventActivation(
	//			makeNewEventActivation(testNamespace, eaName)),
	//	},
	//},
	//{
	//	Name: "New event activation will activate subscription",
	//	InitialState: []runtime.Object{
	//		makeEventsDeactivatedSubscription(subName),
	//		addEventActivationFinalizer(
	//			makeNewEventActivation(testNamespace, eaName), finalizerName),
	//	},
	//	ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, eaName),
	//	WantResult:   reconcile.Result{},
	//	WantPresent: []runtime.Object{
	//		makeEventsActivatedSubscription(subName),
	//	},
	//},
	//{
	//	Name: "Marked to be deleted event activation will deactivate subscription",
	//	InitialState: []runtime.Object{
	//		makeEventsActivatedSubscription(subName),
	//		markedToBeDeletedEventActivation(
	//			addEventActivationFinalizer(
	//				makeNewEventActivation(testNamespace, eaName), finalizerName)),
	//	},
	//	ReconcileKey: fmt.Sprintf("%s/%s", testNamespace, eaName),
	//	WantResult:   reconcile.Result{},
	//	WantPresent: []runtime.Object{
	//		makeEventsDeactivatedSubscription(subName),
	//	},
	//},
}

func TestAllCases(t *testing.T) {
	//recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	//for _, tc := range testCases {
	//	c := tc.GetClient()
	//	r := &reconciler{
	//		client:   c,
	//		recorder: recorder,
	//		time:     NewMockCurrentTime(),
	//	}
	//	tc.IgnoreTimes = true
	//	t.Logf("Running test %s", tc.Name)
	//	t.Run(tc.Name, tc.Runner(t, r, c, tc.GetEventRecorder()))
	//}
	var ctor Ctor = func(ctx context.Context, ls *Listers) controller.Reconciler {
		rb := reconciler.NewBase(ctx, controllerAgentName, configmap.NewStaticWatcher())
		r := &Reconciler{
			Base:                       rb,
			eventActivationLister:      ls.GetEventActivationLister(),
			applicationconnectorClient: fakeeventbusclient.Get(ctx).ApplicationconnectorV1alpha1(),
			eventingClient:             fakeeventbusclient.Get(ctx).EventingV1alpha1(),
		}

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}

//func mockSubscriptionEmptyList() []controllertesting.MockList {
//	return []controllertesting.MockList{
//		func(_ client.Client, _ context.Context, _ *client.ListOptions, obj runtime.Object) (controllertesting.MockHandled, error) {
//			if _, ok := obj.(*subApis.SubscriptionList); ok {
//				return controllertesting.Handled, nil
//			}
//			return controllertesting.Unhandled, nil
//		},
//	}
//}

//func mockSubscriptionActivatedGet() []controllertesting.MockGet {
//	return []controllertesting.MockGet{
//		func(innerClient client.Client, ctx context.Context, key client.ObjectKey, obj runtime.Object) (controllertesting.MockHandled, error) {
//			if _, ok := obj.(*subApis.Subscription); ok {
//				obj = makeEventsActivatedSubscription(subName)
//				return controllertesting.Handled, nil
//			}
//			return controllertesting.Unhandled, nil
//		},
//	}
//}
//
//func mockSubscriptionUpdate() []controllertesting.MockUpdate {
//	return []controllertesting.MockUpdate{
//		func(innerClient client.Client, ctx context.Context, obj runtime.Object) (controllertesting.MockHandled, error) {
//			if _, ok := obj.(*subApis.Subscription); ok {
//				return controllertesting.Handled, nil
//			}
//			return controllertesting.Unhandled, nil
//		},
//	}
//}
//
//func mockSubscriptionDeactivatedList() []controllertesting.MockList {
//	return []controllertesting.MockList{
//		func(_ client.Client, _ context.Context, _ *client.ListOptions, obj runtime.Object) (controllertesting.MockHandled, error) {
//			if l, ok := obj.(*subApis.SubscriptionList); ok {
//				l.Items = []subApis.Subscription{
//					*makeEventsDeactivatedSubscription(subName),
//				}
//				return controllertesting.Handled, nil
//			}
//			return controllertesting.Unhandled, nil
//		},
//	}
//}

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

// Mock the current time for Status "LastTranscationTime"
type MockCurrentTime struct{}

func NewMockCurrentTime() util.CurrentTime {
	mockCurrentTime := new(MockCurrentTime)
	return mockCurrentTime
}

func (m *MockCurrentTime) GetCurrentTime() metav1.Time {
	return metav1.NewTime(time.Time{})
}
