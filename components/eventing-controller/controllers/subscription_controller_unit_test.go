package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func Test_replaceStatusCondition(t *testing.T) {
	var tests = []struct {
		name              string
		giveSubscription  *eventingv1alpha1.Subscription
		giveCondition     eventingv1alpha1.Condition
		wantStatusChanged bool
		wantError         bool
		wantStatus        *eventingv1alpha1.SubscriptionStatus
		wantReady         bool
	}{
		{
			name: "Updating a condition marks the status as changed",
			giveSubscription: func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("some-name", "some-namespace")
				subscription.Status.InitializeConditions()
				return subscription
			}(),
			giveCondition: func() eventingv1alpha1.Condition {
				return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue)
			}(),
			wantStatusChanged: true,
			wantError:         false,
			wantReady:         false,
		},
		{
			name: "All conditions true means status is ready",
			giveSubscription: func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("some-name", "some-namespace")
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
				return eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, corev1.ConditionTrue)
			}(),
			wantStatusChanged: true, // readyness changed
			wantError:         false,
			wantReady:         true, // all conditions are true
		},
	}

	g := NewGomegaWithT(t)
	r := SubscriptionReconciler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := tt.giveSubscription
			condition := tt.giveCondition
			statusChanged, err := r.replaceStatusCondition(subscription, condition)
			if tt.wantError {
				g.Expect(err).Should(HaveOccurred())
			} else {
				g.Expect(err).ShouldNot(HaveOccurred())
			}
			g.Expect(statusChanged).To(BeEquivalentTo(tt.wantStatusChanged))
			g.Expect(subscription.Status.Conditions).To(ContainElement(condition))
			g.Expect(subscription.Status.Ready).To(BeEquivalentTo(tt.wantReady))
		})
	}
}
