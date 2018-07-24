package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"k8s.io/api/core/v1"
)

type limitRangeConverter struct{}

func (lr *limitRangeConverter) ToGQL(in *v1.LimitRange) *gqlschema.LimitRange {
	if in == nil {
		return nil
	}
	out := &gqlschema.LimitRange{
		Name:   in.Name,
		Limits: make([]gqlschema.LimitRangeItem, 0, len(in.Spec.Limits)),
	}

	for _, limitRange := range in.Spec.Limits {
		out.Limits = append(out.Limits, lr.limitsToGQL(limitRange))
	}

	return out
}

func (lr *limitRangeConverter) ToGQLs(in []*v1.LimitRange) []gqlschema.LimitRange {
	if in == nil {
		return nil
	}
	result := make([]gqlschema.LimitRange, 0)
	for _, limitRange := range in {
		if lr := lr.ToGQL(limitRange); lr != nil {
			result = append(result, *lr)
		}
	}
	return result
}

func (lr *limitRangeConverter) limitsToGQL(item v1.LimitRangeItem) gqlschema.LimitRangeItem {
	return gqlschema.LimitRangeItem{
		LimitType:      gqlschema.LimitType(item.Type),
		DefaultRequest: lr.extractResourceValues(item.DefaultRequest),
		Default:        lr.extractResourceValues(item.Default),
		Max:            lr.extractResourceValues(item.Max),
	}
}

func (lr *limitRangeConverter) extractResourceValues(item v1.ResourceList) gqlschema.ResourceType {
	rt := gqlschema.ResourceType{}
	if item, ok := item[v1.ResourceCPU]; ok {
		rt.Cpu = lr.stringPtr(item.String())
	}
	if item, ok := item[v1.ResourceMemory]; ok {
		rt.Memory = lr.stringPtr(item.String())
	}

	return rt
}

func (*limitRangeConverter) stringPtr(str string) *string {
	return &str
}
