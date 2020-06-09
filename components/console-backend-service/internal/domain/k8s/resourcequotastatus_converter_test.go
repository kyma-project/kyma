package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestResourceQuotaStatusConverter_ToGQL(t *testing.T) {
	// given
	conv := resourceQuotaStatusConverter{}
	for tn, tc := range map[string]struct {
		InputMap       map[string]map[v1.ResourceName][]string
		ExceededQuotas []*gqlschema.ExceededQuota
	}{
		"success": {
			InputMap: map[string]map[v1.ResourceName][]string{
				"rq-a": {
					v1.ResourceLimitsMemory: {
						"fix-a", "fix-b",
					},
				},
			},
			ExceededQuotas: []*gqlschema.ExceededQuota{
				{
					QuotaName:    "rq-a",
					ResourceName: string(v1.ResourceLimitsMemory),
					AffectedResources: []string{
						"fix-a", "fix-b",
					},
				},
			},
		},
		"nil input": {
			InputMap:       nil,
			ExceededQuotas: []*gqlschema.ExceededQuota{},
		},
		"empty input": {
			InputMap:       map[string]map[v1.ResourceName][]string{},
			ExceededQuotas: []*gqlschema.ExceededQuota{},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			status := conv.ToGQL(tc.InputMap)
			assert.Equal(t, status, tc.ExceededQuotas)
		})
	}
}
