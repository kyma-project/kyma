package servicecatalog

import (
	"encoding/json"

	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

func (c *serviceBindingConverter) ToGQL(in *api.ServiceBinding) (*gqlschema.ServiceBinding, error) {
	if in == nil {
		return nil, nil
	}

	params, err := c.extractParameters(in.Spec.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting parameters from service binding [name: %s][environment: %s]", in.Name, in.Namespace)
	}

	return &gqlschema.ServiceBinding{
		Name:                in.Name,
		ServiceInstanceName: in.Spec.ServiceInstanceRef.Name,
		Environment:         in.Namespace,
		SecretName:          in.Spec.SecretName,
		Status:              c.extractor.Status(in.Status.Conditions),
		Parameters:          params,
	}, nil
}

func (c *serviceBindingConverter) ToGQLs(in []*api.ServiceBinding) (gqlschema.ServiceBindings, error) {
	var result gqlschema.ServiceBindings
	for _, item := range in {
		converted, err := c.ToGQL(item)
		if err != nil {
			return gqlschema.ServiceBindings{}, errors.Wrapf(err, "while converting service binding [name: %s][environment: %s]", item.Name, item.Namespace)
		}
		if converted != nil {
			c.addStat(converted.Status.Type, &result.Stats)
			result.Items = append(result.Items, *converted)
		}
	}
	return result, nil
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

func (*serviceBindingConverter) extractParameters(ext *runtime.RawExtension) (map[string]interface{}, error) {
	if ext == nil {
		return nil, nil
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(ext.Raw, &result)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling binding parameters")
	}

	return result, nil
}
