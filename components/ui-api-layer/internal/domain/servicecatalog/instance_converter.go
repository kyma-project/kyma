package servicecatalog

import (
	"fmt"
	"strings"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type instanceConverter struct {
	extractor status.InstanceExtractor
}

func (c *instanceConverter) ToGQL(in *v1beta1.ServiceInstance) (*gqlschema.ServiceInstance, error) {
	if in == nil {
		return nil, nil
	}

	instanceLabels := c.extractLabels(in)
	servicePlanSpec, err := c.extractServicePlanSpec(in.Spec.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting servicePlanSpec for %s `%s`", pretty.ServiceInstance, in.Name)
	}

	var servicePlanName *string
	if in.Spec.ClusterServicePlanRef != nil {
		servicePlanName = &in.Spec.ClusterServicePlanRef.Name
	}

	var serviceClassName *string
	if in.Spec.ClusterServiceClassRef != nil {
		serviceClassName = &in.Spec.ClusterServiceClassRef.Name
	}

	instance := gqlschema.ServiceInstance{
		Name:                    in.Name,
		Environment:             in.Namespace,
		ServicePlanName:         servicePlanName,
		ServicePlanDisplayName:  in.Spec.ClusterServicePlanExternalName,
		ServiceClassName:        serviceClassName,
		ServiceClassDisplayName: in.Spec.ClusterServiceClassExternalName,
		ServicePlanSpec:         servicePlanSpec,
		Labels:                  instanceLabels,
		Status:                  c.ServiceStatusToGQLStatus(c.extractor.Status(*in)),
		CreationTimestamp:       in.CreationTimestamp.Time,
	}

	return &instance, nil
}

func (c *instanceConverter) ToGQLs(in []*v1beta1.ServiceInstance) ([]gqlschema.ServiceInstance, error) {
	var result []gqlschema.ServiceInstance
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *instanceConverter) GQLCreateInputToInstanceCreateParameters(in *gqlschema.ServiceInstanceCreateInput) *instanceCreateParameters {
	if in == nil {
		return nil
	}

	var parameterSchema map[string]interface{}
	if in.ParameterSchema != nil {
		parameterSchema = *in.ParameterSchema
	}

	var labels []string
	for _, label := range in.Labels {
		labels = append(labels, label)
	}

	parameters := instanceCreateParameters{
		Name:                     in.Name,
		Namespace:                in.Environment,
		Labels:                   labels,
		ExternalServicePlanName:  in.ExternalPlanName,
		ExternalServiceClassName: in.ExternalServiceClassName,
		Schema: parameterSchema,
	}

	return &parameters
}

func (c *instanceConverter) ServiceStatusTypeToGQLStatusType(in status.ServiceInstanceStatusType) gqlschema.InstanceStatusType {
	if &in == nil {
		return ""
	}

	switch in {
	case status.ServiceInstanceStatusTypeRunning:
		return gqlschema.InstanceStatusTypeRunning
	case status.ServiceInstanceStatusTypeProvisioning:
		return gqlschema.InstanceStatusTypeProvisioning
	case status.ServiceInstanceStatusTypeDeprovisioning:
		return gqlschema.InstanceStatusTypeDeprovisioning
	case status.ServiceInstanceStatusTypeFailed:
		return gqlschema.InstanceStatusTypeFailed
	default:
		return gqlschema.InstanceStatusTypePending
	}
}

func (c *instanceConverter) GQLStatusTypeToServiceStatusType(in gqlschema.InstanceStatusType) status.ServiceInstanceStatusType {
	if &in == nil {
		return ""
	}

	switch in {
	case gqlschema.InstanceStatusTypeRunning:
		return status.ServiceInstanceStatusTypeRunning
	case gqlschema.InstanceStatusTypeProvisioning:
		return status.ServiceInstanceStatusTypeProvisioning
	case gqlschema.InstanceStatusTypeDeprovisioning:
		return status.ServiceInstanceStatusTypeDeprovisioning
	case gqlschema.InstanceStatusTypeFailed:
		return status.ServiceInstanceStatusTypeFailed
	default:
		return status.ServiceInstanceStatusTypePending
	}
}

func (c *instanceConverter) GQLStatusToServiceStatus(in *gqlschema.ServiceInstanceStatus) *status.ServiceInstanceStatus {
	if in == nil {
		return nil
	}

	return &status.ServiceInstanceStatus{
		Type:    c.GQLStatusTypeToServiceStatusType(in.Type),
		Reason:  in.Reason,
		Message: in.Message,
	}
}

func (c *instanceConverter) ServiceStatusToGQLStatus(in status.ServiceInstanceStatus) gqlschema.ServiceInstanceStatus {
	return gqlschema.ServiceInstanceStatus{
		Type:    c.ServiceStatusTypeToGQLStatusType(in.Type),
		Reason:  in.Reason,
		Message: in.Message,
	}
}

func (c *instanceConverter) extractLabels(in *v1beta1.ServiceInstance) []string {
	if in == nil || len(in.Annotations["tags"]) == 0 {
		return []string{}
	}

	return strings.Split(in.Annotations["tags"], ",")
}

func (c *instanceConverter) populateLabels(inputLabels interface{}) ([]string, error) {
	if inputLabels == nil {
		return []string{}, nil
	}

	items, ok := inputLabels.([]interface{})
	if !ok {
		return []string{}, fmt.Errorf("Incorrect items type %T: should be []interface{}", inputLabels)
	}

	var labels []string
	for _, item := range items {
		label, ok := item.(string)
		if !ok {
			return []string{}, fmt.Errorf("Incorrect item type %T: should be string", inputLabels)
		}
		labels = append(labels, label)
	}

	return labels, nil
}

func (c *instanceConverter) extractServicePlanSpec(raw *runtime.RawExtension) (*gqlschema.JSON, error) {
	if raw == nil {
		return nil, nil
	}

	extracted, err := resource.ExtractRawToMap("ServicePlanSpec", raw.Raw)
	if err != nil {
		return nil, err
	}

	result := make(gqlschema.JSON)
	for k, v := range extracted {
		result[k] = v
	}

	return &result, err
}
