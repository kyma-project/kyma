package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/magiconair/properties/assert"
	"k8s.io/api/core/v1"
)

func TestResourceQuotaStatusConverter_ToGQL(t *testing.T) {
	// given
	conv := resourceQuotaStatusConverter{}

	// when
	exQuotas := conv.ToGQL(fixStatusesMap())

	// then
	assert.Equal(t, exQuotas, fixExceededQuotas())
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
			Name: "rq-a",
			ResourcesRequests: []gqlschema.ResourcesRequests{
				{
					ResourceType: string(v1.ResourceLimitsMemory),
					DemandingResources: []string{
						"fix-a", "fix-b",
					},
				},
			},
		},
		{
			Name: "rq-b",
			ResourcesRequests: []gqlschema.ResourcesRequests{
				{
					ResourceType: string(v1.ResourceRequestsMemory),
					DemandingResources: []string{
						"fix-c",
					},
				},
			},
		},
	}
}
