package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type resourceQuotaConverter struct{}

func (c *resourceQuotaConverter) ToGQL(in *v1.ResourceQuota) (*gqlschema.ResourceQuota, error) {
	if in == nil {
		return nil, nil
	}

	out := &gqlschema.ResourceQuota{
		Name: in.Name,
		Pods: c.extractValue(in, v1.ResourcePods),
		Limits: gqlschema.ResourceValues{
			Memory: c.extractValue(in, v1.ResourceLimitsMemory),
			CPU:    c.extractValue(in, v1.ResourceLimitsCPU),
		},
		Requests: gqlschema.ResourceValues{
			Memory: c.extractValue(in, v1.ResourceRequestsMemory),
			CPU:    c.extractValue(in, v1.ResourceRequestsCPU),
		},
	}
	return out, nil
}

func (c *resourceQuotaConverter) extractValue(in *v1.ResourceQuota, resourceName v1.ResourceName) *string {
	val, exists := in.Spec.Hard[resourceName]
	if !exists {
		return nil
	}
	formattedVal := val.String()
	return &formattedVal
}

func (c *resourceQuotaConverter) ToGQLs(in []*v1.ResourceQuota) ([]gqlschema.ResourceQuota, error) {
	result := make([]gqlschema.ResourceQuota, 0)
	for _, rq := range in {
		converted, err := c.ToGQL(rq)
		if err != nil {
			return nil, err
		}
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
