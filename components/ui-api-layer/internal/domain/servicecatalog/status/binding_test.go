package status

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestBindingExtractor_Status(t *testing.T) {
	// GIVEN
	ext := BindingExtractor{}
	for tn, tc := range map[string]struct {
		given    []v1beta1.ServiceBindingCondition
		expected gqlschema.ServiceBindingStatus
	}{
		"ReadyStatus": {
			given: []v1beta1.ServiceBindingCondition{
				{
					Status: v1beta1.ConditionTrue,
					Type:   v1beta1.ServiceBindingConditionReady,
				},
			},
			expected: gqlschema.ServiceBindingStatus{
				Type: gqlschema.ServiceBindingStatusTypeReady,
			},
		},
		"FailedStatus": {
			given: []v1beta1.ServiceBindingCondition{
				{
					Status:  v1beta1.ConditionFalse,
					Reason:  "error",
					Message: "supa error",
					Type:    v1beta1.ServiceBindingConditionReady,
				},
			},
			expected: gqlschema.ServiceBindingStatus{
				Type:    gqlschema.ServiceBindingStatusTypeFailed,
				Reason:  "error",
				Message: "supa error",
			},
		},
		"EmptyStatus": {
			given: []v1beta1.ServiceBindingCondition{},
			expected: gqlschema.ServiceBindingStatus{
				Type: gqlschema.ServiceBindingStatusTypePending,
			},
		},
		"UnknownStatus": {
			given: []v1beta1.ServiceBindingCondition{
				{
					Type: "different",
				},
			},
			expected: gqlschema.ServiceBindingStatus{
				Type: gqlschema.ServiceBindingStatusTypeUnknown,
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
