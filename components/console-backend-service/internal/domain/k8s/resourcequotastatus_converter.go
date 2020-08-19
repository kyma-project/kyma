package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type resourceQuotaStatusConverter struct{}

func (*resourceQuotaStatusConverter) ToGQL(status map[string]map[v1.ResourceName][]string) []*gqlschema.ExceededQuota {
	quotas := make([]*gqlschema.ExceededQuota, 0)
	for quotaName, resourceTypes := range status {
		for resourceName, affectedResources := range resourceTypes {
			quotas = append(quotas, &gqlschema.ExceededQuota{QuotaName: quotaName, AffectedResources: affectedResources, ResourceName: string(resourceName)})
		}
	}
	return quotas
}
