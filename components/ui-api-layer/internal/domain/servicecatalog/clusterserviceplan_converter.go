package servicecatalog

import (
	"encoding/base64"
	"encoding/json"

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
		unpackedSchema, err := p.unpackInstanceCreateParameterSchema(item.Spec.ServiceInstanceCreateParameterSchema.Raw)
		if err != nil {
			return nil, p.wrapConversionError(err, item.Name)
		}
		instanceCreateParameterSchema = &unpackedSchema
	}

	displayName := resource.ToStringPtr(externalMetadata["displayName"])
	plan := gqlschema.ClusterServicePlan{
		Name:                           item.Name,
		ExternalName:                   item.Spec.ExternalName,
		DisplayName:                    displayName,
		Description:                    item.Spec.Description,
		RelatedClusterServiceClassName: item.Spec.ClusterServiceClassRef.Name,
		InstanceCreateParameterSchema:  instanceCreateParameterSchema,
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

func (p *clusterServicePlanConverter) unpackInstanceCreateParameterSchema(raw []byte) (gqlschema.JSON, error) {
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

func (p *clusterServicePlanConverter) extractInstanceCreateSchema(raw []byte) (map[string]interface{}, error) {
	extracted := make(map[string]interface{})

	err := json.Unmarshal(raw, &extracted)
	if err != nil {
		return nil, errors.Wrap(err, "while extracting instance creation parameter schema")
	}

	return extracted, nil
}
