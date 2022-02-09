package beb

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func Test_isInDeletion(t *testing.T) {
	var testCases = []struct {
		name              string
		givenSubscription func() *eventingv1alpha1.Subscription
		isInDeletion      bool
	}{
		{
			name: "Deletion timestamp uninitialized",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				sub := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				sub.DeletionTimestamp = nil
				return sub
			},
			isInDeletion: false,
		},
		{
			name: "Deletion timestamp is zero",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				zero := metav1.Time{}
				sub := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				sub.DeletionTimestamp = &zero
				return sub
			},
			isInDeletion: false,
		},
		{
			name: "Deletion timestamp is set to a useful time",
			givenSubscription: func() *eventingv1alpha1.Subscription {
				newTime := metav1.NewTime(time.Now())
				sub := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				sub.DeletionTimestamp = &newTime
				return sub
			},
			isInDeletion: true,
		},
	}
	g := NewGomegaWithT(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			givenSubscription := tt.givenSubscription()
			g.Expect(isInDeletion(givenSubscription)).To(Equal(tt.isInDeletion))
		})
	}
}
func Test_replaceStatusCondition(t *testing.T) {
	var testCases = []struct {
		name              string
		giveSubscription  *eventingv1alpha1.Subscription
		giveCondition     eventingv1alpha1.Condition
		wantStatusChanged bool
		wantStatus        *eventingv1alpha1.SubscriptionStatus
		wantReady         bool
	}{
		{
			name: "Updating a condition marks the status as changed",
			giveSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter())
				subscription.Status.InitializeConditions()
				return subscription
			}(),
			giveCondition: func() eventingv1alpha1.Condition {
				return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, "")
			}(),
			wantStatusChanged: true,
			wantReady:         false,
		},
		{
			name: "All conditions true means status is ready",
			giveSubscription: func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("some-name", "some-namespace",
					reconcilertesting.WithNotCleanFilter(),
					reconcilertesting.WithWebhookAuthForBEB())
				subscription.Status.InitializeConditions()
				subscription.Status.Ready = false

				// mark all conditions as true
				subscription.Status.Conditions = []eventingv1alpha1.Condition{
					{
						Type:               eventingv1alpha1.ConditionSubscribed,
						LastTransitionTime: metav1.Now(),
						Status:             corev1.ConditionTrue,
					},
					{
						Type:               eventingv1alpha1.ConditionSubscriptionActive,
						LastTransitionTime: metav1.Now(),
						Status:             corev1.ConditionTrue,
					},
				}
				return subscription
			}(),
			giveCondition: func() eventingv1alpha1.Condition {
				return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, "")
			}(),
			wantStatusChanged: true, // readiness changed
			wantReady:         true, // all conditions are true
		},
	}

	g := NewGomegaWithT(t)
	r := Reconciler{}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			subscription := tt.giveSubscription
			condition := tt.giveCondition
			statusChanged := r.replaceStatusCondition(subscription, condition)
			g.Expect(statusChanged).To(BeEquivalentTo(tt.wantStatusChanged))
			g.Expect(subscription.Status.Conditions).To(ContainElement(condition))
			g.Expect(subscription.Status.Ready).To(BeEquivalentTo(tt.wantReady))
		})
	}
}

func Test_getRequiredConditions(t *testing.T) {
	var emptySubscriptionStatus eventingv1alpha1.SubscriptionStatus
	emptySubscriptionStatus.InitializeConditions()
	expectedConditions := emptySubscriptionStatus.Conditions

	testCases := []struct {
		name                   string
		subscriptionConditions []eventingv1alpha1.Condition
		wantConditions         []eventingv1alpha1.Condition
	}{
		{
			name:                   "should get expected conditions if the subscription has no conditions",
			subscriptionConditions: []eventingv1alpha1.Condition{},
			wantConditions:         expectedConditions,
		},
		{
			name: "should get subscription conditions if the all the expected conditions are present",
			subscriptionConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionFalse},
			},
			wantConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionFalse},
			},
		},
		{
			name: "should get latest conditions Status compared to the expected condition status",
			subscriptionConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
			},
			wantConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionUnknown},
			},
		},
		{
			name: "should get rid of unwanted conditions in the subscription, if present",
			subscriptionConditions: []eventingv1alpha1.Condition{
				{Type: "Fake Condition Type", Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
			},
			wantConditions: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionUnknown},
			},
		},
	}

	g := NewGomegaWithT(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotConditions := getRequiredConditions(tc.subscriptionConditions, expectedConditions)
			if !conditionsEquals(gotConditions, tc.wantConditions) {
				t.Errorf("ShouldUpdateReadyStatus is not valid, want: %v but got: %v", tc.wantConditions, gotConditions)
			}
			g.Expect(len(gotConditions)).To(BeEquivalentTo(len(expectedConditions)))
		})
	}
}
