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
			wantConditionTypes := []ConditionType{ConditionSubscribed, ConditionSubscriptionActive}

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
