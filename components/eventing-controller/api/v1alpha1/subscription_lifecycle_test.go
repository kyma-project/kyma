package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_InitializeConditions(t *testing.T) {
	var tests = []struct {
		name            string
		givenConditions []Condition
	}{
		{
			name: "Conditions empty",
			givenConditions: func() []Condition {
				return makeConditions()
			}(),
		},
		{
			name: "Conditions partially initialized",
			givenConditions: func() []Condition {
				// on purpose we only set one condition
				return []Condition{
					{
						Type:               ConditionSubscribed,
						LastTransitionTime: metav1.Now(),
						Status:             corev1.ConditionUnknown,
					},
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			g := NewGomegaWithT(t)
			s := SubscriptionStatus{}
			s.Conditions = tt.givenConditions
			wantConditionTypes := []ConditionType{ConditionSubscribed, ConditionSubscriptionActive, ConditionAPIRuleStatus}

			// then
			s.InitializeConditions()

			//when
			g.Expect(s.Conditions).To(HaveLen(len(wantConditionTypes)))
			foundConditionTypes := make([]ConditionType, 0)
			for _, condition := range s.Conditions {
				g.Expect(condition.Status).To(BeEquivalentTo(corev1.ConditionUnknown))
				foundConditionTypes = append(foundConditionTypes, condition.Type)
			}
			g.Expect(wantConditionTypes).To(ConsistOf(foundConditionTypes))
		})
	}
}

func Test_IsReady(t *testing.T) {
	testCases := []struct {
		name            string
		givenConditions []Condition
		wantReadyStatus bool
	}{
		{
			name:            "should not be ready if conditions are nil",
			givenConditions: nil,
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if conditions are empty",
			givenConditions: []Condition{{}},
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if only ConditionSubscribed is available and true",
			givenConditions: []Condition{{Type: ConditionSubscribed, Status: corev1.ConditionTrue}},
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if only ConditionSubscriptionActive is available and true",
			givenConditions: []Condition{{Type: ConditionSubscriptionActive, Status: corev1.ConditionTrue}},
			wantReadyStatus: false,
		},
		{
			name:            "should not be ready if only ConditionAPIRuleStatus is available and true",
			givenConditions: []Condition{{Type: ConditionAPIRuleStatus, Status: corev1.ConditionTrue}},
			wantReadyStatus: false,
		},
		{
			name: "should not be ready if all conditions are unknown",
			givenConditions: []Condition{
				{Type: ConditionSubscribed, Status: corev1.ConditionUnknown},
				{Type: ConditionSubscriptionActive, Status: corev1.ConditionUnknown},
				{Type: ConditionAPIRuleStatus, Status: corev1.ConditionUnknown},
			},
			wantReadyStatus: false,
		},
		{
			name: "should not be ready if all conditions are false",
			givenConditions: []Condition{
				{Type: ConditionSubscribed, Status: corev1.ConditionFalse},
				{Type: ConditionSubscriptionActive, Status: corev1.ConditionFalse},
				{Type: ConditionAPIRuleStatus, Status: corev1.ConditionFalse},
			},
			wantReadyStatus: false,
		},
		{
			name: "should be ready if all conditions are true",
			givenConditions: []Condition{
				{Type: ConditionSubscribed, Status: corev1.ConditionTrue},
				{Type: ConditionSubscriptionActive, Status: corev1.ConditionTrue},
				{Type: ConditionAPIRuleStatus, Status: corev1.ConditionTrue},
			},
			wantReadyStatus: true,
		},
	}

	status := SubscriptionStatus{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status.Conditions = tc.givenConditions
			if gotReadyStatus := status.IsReady(); tc.wantReadyStatus != gotReadyStatus {
				t.Errorf("Subscription status is not valid, want: %v but got: %v", tc.wantReadyStatus, gotReadyStatus)
			}
		})
	}
}
