package servicecatalog

import (
	"encoding/base64"
	"encoding/json"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/resource"
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
	if item.Spec.ServiceInstanceCreateParameterSchema != nil {
		unpackedSchema, err := p.unpackInstanceCreateParameterSchema(item.Spec.ServiceInstanceCreateParameterSchema.Raw)
		if err != nil {
			return nil, p.wrapConversionError(err, item.Name)
		}
		instanceCreateParameterSchema = &unpackedSchema
	}

	displayName := resource.ToStringPtr(externalMetadata["displayName"])
	plan := gqlschema.ServicePlan{
		Name:                          item.Name,
		Environment:                   item.Namespace,
		ExternalName:                  item.Spec.ExternalName,
		DisplayName:                   displayName,
		Description:                   item.Spec.Description,
		RelatedServiceClassName:       item.Spec.ServiceClassRef.Name,
		InstanceCreateParameterSchema: instanceCreateParameterSchema,
	}

	return &plan, nil
}

func (c *servicePlanConverter) ToGQLs(in []*v1beta1.ServicePlan) ([]gqlschema.ServicePlan, error) {
	var result []gqlschema.ServicePlan
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

func (p *servicePlanConverter) wrapConversionError(err error, name string) error {
	return errors.Wrapf(err, "while converting item %s to ServicePlan", name)
}

func (p *servicePlanConverter) unpackInstanceCreateParameterSchema(raw []byte) (gqlschema.JSON, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	_, err := base64.StdEncoding.Decode(decoded, raw)
	if err != nil {
		return p.extractInstanceCreateSchema(raw)
	}

	return p.extractInstanceCreateSchema(decoded)
}

func (p *servicePlanConverter) extractInstanceCreateSchema(raw []byte) (map[string]interface{}, error) {
	extracted := make(map[string]interface{})

	err := json.Unmarshal(raw, &extracted)
	if err != nil {
		return nil, errors.Wrap(err, "while extracting instance creation parameter schema")
	}

	return extracted, nil
}
