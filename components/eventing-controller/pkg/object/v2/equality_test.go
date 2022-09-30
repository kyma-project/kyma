package v2

import (
	"testing"

	"github.com/stretchr/testify/require"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

func Test_isSubscriptionStatusEqual(t *testing.T) {
	testCases := []struct {
		name                string
		subscriptionStatus1 eventingv1alpha2.SubscriptionStatus
		subscriptionStatus2 eventingv1alpha2.SubscriptionStatus
		wantEqualStatus     bool
	}{
		{
			name: "should not be equal if the conditions are not equal",
			subscriptionStatus1: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionFalse},
				},
				Ready: true,
			},
			wantEqualStatus: false,
		},
		{
			name: "should not be equal if the ready status is not equal",
			subscriptionStatus1: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
			},
			subscriptionStatus2: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: false,
			},
			wantEqualStatus: false,
		},
		{
			name: "should be equal if all the fields are equal",
			subscriptionStatus1: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
				Backend: eventingv1alpha2.Backend{
					APIRuleName: "APIRule",
				},
			},
			subscriptionStatus2: eventingv1alpha2.SubscriptionStatus{
				Conditions: []eventingv1alpha2.Condition{
					{Type: eventingv1alpha2.ConditionSubscribed, Status: corev1.ConditionTrue},
				},
				Ready: true,
				Backend: eventingv1alpha2.Backend{
					APIRuleName: "APIRule",
				},
			},
			wantEqualStatus: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotEqualStatus := IsSubscriptionStatusEqual(tc.subscriptionStatus1, tc.subscriptionStatus2)
			require.Equal(t, tc.wantEqualStatus, gotEqualStatus)
		})
	}
}
