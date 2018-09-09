package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestResourceQuotaStatusConverter_ToGQL(t *testing.T) {
	// given
	conv := resourceQuotaStatusConverter{}
	fixQuotas := fixExceededQuotas()

	// when
	exQuotas := conv.ToGQL(fixStatusesMap())

	// then
	assert.Contains(t, exQuotas, fixQuotas[0])
	assert.Contains(t, exQuotas, fixQuotas[1])
}

func fixStatusesMap() map[string]map[v1.ResourceName][]string {
	return map[string]map[v1.ResourceName][]string{
		"rq-a": {
			v1.ResourceLimitsMemory: {
				"fix-a", "fix-b",
			},
		},
		"rq-b": {
			v1.ResourceRequestsMemory: {
				"fix-c",
			},
		},
	}
}

func fixExceededQuotas() []gqlschema.ExceededQuota {
	return []gqlschema.ExceededQuota{
		{
			QuotaName:    "rq-a",
			ResourceName: string(v1.ResourceLimitsMemory),
			AffectedResources: []string{
				"fix-a", "fix-b",
			},
		},
		{
			QuotaName:    "rq-b",
			ResourceName: string(v1.ResourceRequestsMemory),
			AffectedResources: []string{
				"fix-c",
			},
		},
	}
}
