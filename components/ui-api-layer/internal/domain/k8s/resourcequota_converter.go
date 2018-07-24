package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"k8s.io/api/core/v1"
)

type resourceQuotaConverter struct{}

func (c *resourceQuotaConverter) ToGQL(in *v1.ResourceQuota) *gqlschema.ResourceQuota {
	if in == nil {
		return nil
	}

	out := &gqlschema.ResourceQuota{
		Name: in.Name,
		Pods: c.extractValue(in, v1.ResourcePods),
		Limits: gqlschema.ResourceValues{
			Memory: c.extractValue(in, v1.ResourceLimitsMemory),
			Cpu:    c.extractValue(in, v1.ResourceLimitsCPU),
		},
		Requests: gqlschema.ResourceValues{
			Memory: c.extractValue(in, v1.ResourceRequestsMemory),
			Cpu:    c.extractValue(in, v1.ResourceRequestsCPU),
		},
	}
	return out
}

func (c *resourceQuotaConverter) extractValue(in *v1.ResourceQuota, resourceName v1.ResourceName) *string {
	val, exists := in.Spec.Hard[resourceName]
	if !exists {
		return nil
	}
	formattedVal := val.String()
	return &formattedVal
}

func (c *resourceQuotaConverter) ToGQLs(in []*v1.ResourceQuota) []gqlschema.ResourceQuota {
	result := make([]gqlschema.ResourceQuota, 0)
	for _, rq := range in {
		converted := c.ToGQL(rq)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
