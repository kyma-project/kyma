package servicecatalog

import (
	"encoding/base64"
	"encoding/json"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/resource"
	"github.com/pkg/errors"
)

type planConverter struct{}

func (p *planConverter) ToGQL(item *v1beta1.ClusterServicePlan) (*gqlschema.ServicePlan, error) {
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
		ExternalName:                  item.Spec.ExternalName,
		DisplayName:                   displayName,
		Description:                   item.Spec.Description,
		RelatedServiceClassName:       item.Spec.ClusterServiceClassRef.Name,
		InstanceCreateParameterSchema: instanceCreateParameterSchema,
	}

	return &plan, nil
}

func (c *planConverter) ToGQLs(in []*v1beta1.ClusterServicePlan) ([]gqlschema.ServicePlan, error) {
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

func (p *planConverter) wrapConversionError(err error, name string) error {
	return errors.Wrapf(err, "while converting item %s to ServicePlan", name)
}

func (p *planConverter) unpackInstanceCreateParameterSchema(raw []byte) (gqlschema.JSON, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	//TODO: Change it when fix for helm broker will be ready
	encoded := p.omitQuotationMarksIfShould(raw)
	if len(encoded) == 0 {
		return nil, nil
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
	_, err := base64.StdEncoding.Decode(decoded, encoded)
	if err != nil {
		return p.extractInstanceCreateSchema(raw)
	}

	decoded = p.removeNullCharactersFromEndOfArray(decoded)
	return p.extractInstanceCreateSchema(decoded)
}

// TODO: Figure out why the instanceCreateParameterSchema has quotation marks
func (p *planConverter) omitQuotationMarksIfShould(input []byte) []byte {
	const quotationMarkChar byte = 34
	inputLength := len(input)

	var result []byte
	if input[inputLength-1] != quotationMarkChar {
		return input
	}

	result = append(result, input[1:len(input)-1]...)
	return result
}

// TODO: Investigate why instanceCreateParameterSchema has null character at the end
func (p *planConverter) removeNullCharactersFromEndOfArray(input []byte) []byte {
	const nullChar byte = 0

	sliceEnd := len(input)
	for i := sliceEnd - 1; input[i] == nullChar; i-- {
		sliceEnd = i
	}

	return input[:sliceEnd]
}

func (p *planConverter) extractInstanceCreateSchema(raw []byte) (map[string]interface{}, error) {
	extracted := make(map[string]interface{})

	err := json.Unmarshal(raw, &extracted)
	if err != nil {
		return nil, errors.Wrap(err, "while extracting instance creation parameter schema")
	}

	return extracted, nil
}
