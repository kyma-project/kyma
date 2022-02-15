package beb

import (
	"testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func Test_conditionsEquals(t *testing.T) {
	testCases := []struct {
		name            string
		conditionsSet1  []eventingv1alpha1.Condition
		conditionsSet2  []eventingv1alpha1.Condition
		wantEqualStatus bool
	}{
		{
			name: "should not be equal if the number of conditions are not equal",
			conditionsSet1: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
			},
			conditionsSet2:  []eventingv1alpha1.Condition{},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if the conditions are the same",
			conditionsSet1: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: true,
		},
		{
			name: "should not be equal if the condition types are different",
			conditionsSet1: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionWebhookCallStatus, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionSubscriptionActive, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the condition types are the same but the status is different",
			conditionsSet1: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
			},
			conditionsSet2: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionFalse},
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the condition types are different but the status is the same",
			conditionsSet1: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionFalse},
			},
			conditionsSet2: []eventingv1alpha1.Condition{
				{Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			wantEqualStatus: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotEqualStatus := conditionsEquals(tc.conditionsSet1, tc.conditionsSet2); tc.wantEqualStatus != gotEqualStatus {
				t.Errorf("The list of conditions are not equal, want: %v but got: %v", tc.wantEqualStatus, gotEqualStatus)
			}
		})
	}
}

func Test_conditionEquals(t *testing.T) {
	testCases := []struct {
		name            string
		condition1      eventingv1alpha1.Condition
		condition2      eventingv1alpha1.Condition
		wantEqualStatus bool
	}{
		{
			name: "should not be equal if the types are the same but the status is different",
			condition1: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue,
			},

			condition2: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionUnknown,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the types are different but the status is the same",
			condition1: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue,
			},

			condition2: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionAPIRuleStatus, Status: corev1.ConditionTrue,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the message fields are different",
			condition1: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue, Message: "",
			},

			condition2: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue, Message: "some message",
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the reason fields are different",
			condition1: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue, Reason: eventingv1alpha1.ConditionReasonSubscriptionDeleted,
			},

			condition2: eventingv1alpha1.Condition{
				Type: eventingv1alpha1.ConditionSubscribed, Status: corev1.ConditionTrue, Reason: eventingv1alpha1.ConditionReasonSubscriptionActive,
			},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if all the fields are the same",
			condition1: eventingv1alpha1.Condition{
				Type:    eventingv1alpha1.ConditionAPIRuleStatus,
				Status:  corev1.ConditionFalse,
				Reason:  eventingv1alpha1.ConditionReasonAPIRuleStatusNotReady,
				Message: "API Rule is not ready",
			},
			condition2: eventingv1alpha1.Condition{
				Type:    eventingv1alpha1.ConditionAPIRuleStatus,
				Status:  corev1.ConditionFalse,
				Reason:  eventingv1alpha1.ConditionReasonAPIRuleStatusNotReady,
				Message: "API Rule is not ready",
			},
			wantEqualStatus: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if gotEqualStatus := conditionEquals(tc.condition1, tc.condition2); tc.wantEqualStatus != gotEqualStatus {
				t.Errorf("The conditions are not equal, want: %v but got: %v", tc.wantEqualStatus, gotEqualStatus)
			}
		})
	}
}

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
