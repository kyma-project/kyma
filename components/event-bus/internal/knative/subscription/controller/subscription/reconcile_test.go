package subscription

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	controllertesting "github.com/knative/eventing/pkg/controller/testing"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	subUid             = "test-uid"
	provisioner        = "natss"
	knativeSubsName    = "knativeSubscription"
	knativeChannelName = "knativeChannel"

	testErrorMessage = "test induced error"
)

var (
	// deletionTime is used when objects are marked as deleted. Rfc3339Copy()
	// truncates to seconds to match the loss of precision during serialization.
	deletionTime = metav1.Now().Rfc3339Copy()
)

func init() {
	// Add types to scheme
	eventingv1alpha1.AddToScheme(scheme.Scheme)
	evapisv1alpha1.AddToScheme(scheme.Scheme)
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
}

func TestAllCases(t *testing.T) {
	recorder := record.NewBroadcaster().NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	for _, tc := range testCases {
		c := tc.GetClient()
		r := &reconciler{
			client:   c,
			recorder: recorder,
		}
		t.Logf("Running test %s", tc.Name)
		if tc.ReconcileKey == "" {
			tc.ReconcileKey = fmt.Sprintf("/%s", Name)
		}
		tc.IgnoreTimes = true
		t.Run(tc.Name, tc.Runner(t, r, c))
	}
}

func makeSubscription() *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: Name,
			UID:  subUid,
		},
		SubscriptionSpec: eventingv1alpha1.SubscriptionSpec{},
	}
}

func makeEventsActivatedSubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscription()
	subscription.Status.Conditions = []eventingv1alpha1.SubscriptionCondition{{
		Type:   eventingv1alpha1.EventsActivated,
		Status: eventingv1alpha1.ConditionTrue,
	}}
	return subscription
}

func makeEventsDeactivatedSubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscription()
	subscription.Status.Conditions = []eventingv1alpha1.SubscriptionCondition{{
		Type:   eventingv1alpha1.EventsActivated,
		Status: eventingv1alpha1.ConditionFalse,
	}}
	return subscription
}

func makeSubscriptionWithFinalizer() *eventingv1alpha1.Subscription {
	subscription := makeSubscription()
	subscription.Finalizers = []string{finalizerName}
	return subscription
}

func makeDeletingSubscription() *eventingv1alpha1.Subscription {
	subscription := makeSubscription()
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
