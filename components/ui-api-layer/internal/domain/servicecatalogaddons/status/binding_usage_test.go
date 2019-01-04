package status

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestBindingUsageExtractor_Status(t *testing.T) {
	// GIVEN
	ext := BindingUsageExtractor{}
	for tn, tc := range map[string]struct {
		given    []v1alpha1.ServiceBindingUsageCondition
		expected gqlschema.ServiceBindingUsageStatus
	}{
		"ReadyStatus": {
			given: []v1alpha1.ServiceBindingUsageCondition{
				{
					Status: v1alpha1.ConditionTrue,
					Type:   v1alpha1.ServiceBindingUsageReady,
				},
			},
			expected: gqlschema.ServiceBindingUsageStatus{
				Type: gqlschema.ServiceBindingUsageStatusTypeReady,
			},
		},
		"FailedStatus": {
			given: []v1alpha1.ServiceBindingUsageCondition{
				{
					Status:  v1alpha1.ConditionFalse,
					Reason:  "error",
					Message: "supa error",
					Type:    v1alpha1.ServiceBindingUsageReady,
				},
			},
			expected: gqlschema.ServiceBindingUsageStatus{
				Type:    gqlschema.ServiceBindingUsageStatusTypeFailed,
				Reason:  "error",
				Message: "supa error",
			},
		},
		"EmptyStatus": {
			given: []v1alpha1.ServiceBindingUsageCondition{},
			expected: gqlschema.ServiceBindingUsageStatus{
				Type: gqlschema.ServiceBindingUsageStatusTypePending,
			},
		},
		"UnknownStatus": {
			given: []v1alpha1.ServiceBindingUsageCondition{
				{
					Type: "different",
				},
			},
			expected: gqlschema.ServiceBindingUsageStatus{
				Type: gqlschema.ServiceBindingUsageStatusTypeUnknown,
			},
		},
	} {

		t.Run(tn, func(t *testing.T) {
			// WHEN
			result := ext.Status(tc.given)
			// THEN
			assert.Equal(t, tc.expected, result)
		})
	}
}
