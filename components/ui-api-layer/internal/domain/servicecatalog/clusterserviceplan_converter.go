package servicecatalog

import (
	"encoding/base64"
	"encoding/json"

	"bytes"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/resource"
	"github.com/pkg/errors"
)

type clusterServicePlanConverter struct{}

func (p *clusterServicePlanConverter) ToGQL(item *v1beta1.ClusterServicePlan) (*gqlschema.ClusterServicePlan, error) {
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
		unpackedSchema, err := p.unpackCreateParameterSchema(item.Spec.ServiceInstanceCreateParameterSchema.Raw)
		if err != nil {
			return nil, p.wrapConversionError(err, item.Name)
		}
		if unpackedSchema != nil && len(unpackedSchema) != 0 {
			instanceCreateParameterSchema = &unpackedSchema
		}
	}
	var bindingCreateParameterSchema *gqlschema.JSON
	if item.Spec.ServiceBindingCreateParameterSchema != nil {
		unpackedSchema, err := p.unpackCreateParameterSchema(item.Spec.ServiceBindingCreateParameterSchema.Raw)
		if err != nil {
			return nil, errors.Wrapf(err, "while unpacking service binding create parameter schema from ClusterServicePlan [%s]", item.Name)
		}
		if unpackedSchema != nil && len(unpackedSchema) != 0 {
			bindingCreateParameterSchema = &unpackedSchema
		}
	}

	displayName := resource.ToStringPtr(externalMetadata["displayName"])
	plan := gqlschema.ClusterServicePlan{
		Name:                           item.Name,
		ExternalName:                   item.Spec.ExternalName,
		DisplayName:                    displayName,
		Description:                    item.Spec.Description,
		RelatedClusterServiceClassName: item.Spec.ClusterServiceClassRef.Name,
		InstanceCreateParameterSchema:  instanceCreateParameterSchema,
		BindingCreateParameterSchema:   bindingCreateParameterSchema,
	}

	return &plan, nil
}

func (c *clusterServicePlanConverter) ToGQLs(in []*v1beta1.ClusterServicePlan) ([]gqlschema.ClusterServicePlan, error) {
	var result []gqlschema.ClusterServicePlan
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

func (p *clusterServicePlanConverter) wrapConversionError(err error, name string) error {
	return errors.Wrapf(err, "while converting item %s to ClusterServicePlan", name)
}

func (p *clusterServicePlanConverter) unpackCreateParameterSchema(raw []byte) (gqlschema.JSON, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	_, err := base64.StdEncoding.Decode(decoded, raw)
	if err != nil {
		return p.extractCreateParameterSchema(raw)
	}

	// We need to trim all null characters, because json.Unmarshal is failing when we give
	// data of undesirable length to base64.StdEncoding.DecodedLen function
	decoded = bytes.Trim(decoded, "\x00")
	return p.extractCreateParameterSchema(decoded)
}

func (p *clusterServicePlanConverter) extractCreateParameterSchema(raw []byte) (map[string]interface{}, error) {
	extracted := make(map[string]interface{})

	err := json.Unmarshal(raw, &extracted)
	if err != nil {
		return nil, errors.Wrap(err, "while extracting creation parameter schema")
	}
	if p.isEmptyCreateParameterSchema(extracted) {
		return nil, nil
	}

	return extracted, nil
}

func (p *clusterServicePlanConverter) isEmptyCreateParameterSchema(schema map[string]interface{}) bool {
	if schema["properties"] != nil {
		if properties, ok := schema["properties"].(map[string]interface{}); ok && len(properties) != 0 {
			return false
		}
	}
	if schema["$ref"] != nil {
		if ref, ok := schema["$ref"].(string); ok && ref != "" {
			return false
		}
	}
	return true
}
