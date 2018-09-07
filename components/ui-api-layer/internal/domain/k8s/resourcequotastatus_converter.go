package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"k8s.io/api/core/v1"
)

type resourceQuotaStatusConverter struct{}

func (*resourceQuotaStatusConverter) ToGQL(status map[string]map[v1.ResourceName][]string) []gqlschema.ExceededQuota {
	quotas := make([]gqlschema.ExceededQuota, 0)
	for quotaName, resourceTypes := range status {
		resourcesNeeds := make([]gqlschema.ResourcesRequests, 0)
		for resourceName, messages := range resourceTypes {
			resourcesNeeds = append(resourcesNeeds, gqlschema.ResourcesRequests{
				ResourceType:       string(resourceName),
				DemandingResources: messages,
			})
		}
		quotas = append(quotas, gqlschema.ExceededQuota{Name: quotaName, ResourcesRequests: resourcesNeeds})
	}
	return quotas
}
