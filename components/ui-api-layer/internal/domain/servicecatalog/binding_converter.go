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

func (c *serviceBindingConverter) ToGQLs(in []*api.ServiceBinding) gqlschema.ServiceBindings {
	var result gqlschema.ServiceBindings
	for _, item := range in {
		converted := c.ToGQL(item)
		if converted != nil {
			c.addStat(converted.Status.Type, &result.Stats)
			result.ServiceBindings = append(result.ServiceBindings, *converted)
		}
	}
	return result
}

func (*serviceBindingConverter) addStat(statusType gqlschema.ServiceBindingStatusType, stats *gqlschema.ServiceBindingsStats) {
	switch statusType {
	case gqlschema.ServiceBindingStatusTypeReady:
		stats.Ready += 1
	case gqlschema.ServiceBindingStatusTypeFailed:
		stats.Failed += 1
	case gqlschema.ServiceBindingStatusTypePending:
		stats.Pending += 1
	case gqlschema.ServiceBindingStatusTypeUnknown:
		stats.Unknown += 1
	}
}
