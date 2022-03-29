package beb

import (
	"testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func Test_isSubscriptionStatusEqual(t *testing.T) {
	testCases := []struct {
		name                string
		subscriptionStatus1 eventingv1alpha1.SubscriptionStatus
		subscriptionStatus2 eventingv1alpha1.SubscriptionStatus
		wantEqualStatus     bool
	}{
		{
			name: "should not be equal if the conditions are not equal",
			subscriptionStatus1: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
				},
				Ready: true,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the ready status is not equal",
			subscriptionStatus1: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: false,
			},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if all the fields are equal",
			subscriptionStatus1: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready:       true,
				APIRuleName: "APIRule",
			},
			subscriptionStatus2: eventingv1alpha1.SubscriptionStatus{
				Conditions: []eventingv1alpha1.Condition{
					{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready:       true,
				APIRuleName: "APIRule",
			},
			wantEqualStatus: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotEqualStatus := isSubscriptionStatusEqual(tc.subscriptionStatus1, tc.subscriptionStatus2); tc.wantEqualStatus != gotEqualStatus {
				t.Errorf("The Subsciption Status are not equal, want: %v but got: %v", tc.wantEqualStatus, gotEqualStatus)
			}
		})
	}
}
