package servicecatalog

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/jsonschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
)

type servicePlanConverter struct{}

func (p *servicePlanConverter) ToGQL(item *v1beta1.ServicePlan) (*gqlschema.ServicePlan, error) {
	if item == nil {
		return nil, nil
	}

	var externalMetadata map[string]interface{}
	var err error
	if item.Spec.ExternalMetadata != nil {
		externalMetadata, err = resource.ExtractRawToMap("ExternalMetadata", item.Spec.ExternalMetadata.Raw)
		if err != nil {
			return nil, p.wrapConversionError(err, item.Name)
		}
	}

	var instanceCreateParameterSchema *gqlschema.JSON
	if item.Spec.InstanceCreateParameterSchema != nil {
		instanceCreateParameterSchema, err = jsonschema.Unpack(item.Spec.InstanceCreateParameterSchema.Raw)
		if err != nil {
			return nil, errors.Wrapf(err, "while unpacking service instance create parameter schema from ServicePlan [%s]", item.Name)
		}
	}
	var bindingCreateParameterSchema *gqlschema.JSON
	if item.Spec.ServiceBindingCreateParameterSchema != nil {
		bindingCreateParameterSchema, err = jsonschema.Unpack(item.Spec.ServiceBindingCreateParameterSchema.Raw)
		if err != nil {
			return nil, errors.Wrapf(err, "while unpacking service binding create parameter schema from ServicePlan [%s]", item.Name)
		}
	}

	displayName := resource.ToStringPtr(externalMetadata["displayName"])
	plan := gqlschema.ServicePlan{
		Name:                          item.Name,
		Namespace:                     item.Namespace,
		ExternalName:                  item.Spec.ExternalName,
		DisplayName:                   displayName,
		Description:                   item.Spec.Description,
		RelatedServiceClassName:       item.Spec.ServiceClassRef.Name,
		InstanceCreateParameterSchema: instanceCreateParameterSchema,
		BindingCreateParameterSchema:  bindingCreateParameterSchema,
	}

	return &plan, nil
}

func (c *servicePlanConverter) ToGQLs(in []*v1beta1.ServicePlan) ([]*gqlschema.ServicePlan, error) {
	var result []*gqlschema.ServicePlan
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, converted)
		}
	}
	return result, nil
}

func (p *servicePlanConverter) wrapConversionError(err error, name string) error {
	return errors.Wrapf(err, "while converting item %s to ServicePlan", name)
}
