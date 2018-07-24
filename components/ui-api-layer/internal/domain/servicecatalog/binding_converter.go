package servicecatalog

import (
	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type serviceBindingConverter struct {
	extractor status.BindingExtractor
}

func (c *serviceBindingConverter) ToCreateOutputGQL(in *api.ServiceBinding) *gqlschema.CreateServiceBindingOutput {
	if in == nil {
		return nil
	}

	return &gqlschema.CreateServiceBindingOutput{
		Name:                in.Name,
		Environment:         in.Namespace,
		ServiceInstanceName: in.Spec.ServiceInstanceRef.Name,
	}
}

func (c *serviceBindingConverter) ToGQL(in *api.ServiceBinding) *gqlschema.ServiceBinding {
	if in == nil {
		return nil
	}

	return &gqlschema.ServiceBinding{
		Name:                in.Name,
		ServiceInstanceName: in.Spec.ServiceInstanceRef.Name,
		Environment:         in.Namespace,
		SecretName:          in.Spec.SecretName,
		Status:              c.extractor.Status(in.Status.Conditions),
	}
}

func (c *serviceBindingConverter) ToGQLs(in []*api.ServiceBinding) []gqlschema.ServiceBinding {
	var result []gqlschema.ServiceBinding
	for _, item := range in {
		converted := c.ToGQL(item)
		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result
}
