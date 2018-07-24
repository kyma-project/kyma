package kubeless

import (
	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type functionConverter struct{}

func (c *functionConverter) ToGQL(in *v1beta1.Function) *gqlschema.Function {
	if in == nil {
		return nil
	}

	labels := make(gqlschema.JSON)
	for k, v := range in.Labels {
		labels[k] = v
	}

	return &gqlschema.Function{
		Name:              in.Name,
		Trigger:           in.Spec.Type,
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            labels,
		Environment:       in.Namespace,
	}
}

func (c *functionConverter) ToGQLs(in []*v1beta1.Function) []gqlschema.Function {
	var result []gqlschema.Function
	for _, item := range in {
		converted := c.ToGQL(item)
		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result
}
