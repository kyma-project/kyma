package k8s

import (
	"bytes"
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/types"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/extractor"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type resourceConverter struct{}

func (c *resourceConverter) GQLJSONToResource(in gqlschema.JSON) (types.Resource, error) {
	var buf bytes.Buffer
	in.MarshalGQL(&buf)

	resourceMeta, err := extractor.ExtractResourceMeta(in)
	if err != nil {
		return types.Resource{}, err
	}

	return types.Resource{
		APIVersion: resourceMeta.APIVersion,
		Name:       resourceMeta.Name,
		Namespace:  resourceMeta.Namespace,
		Kind:       resourceMeta.Kind,
		Body:       buf.Bytes(),
	}, nil
}

func (c *resourceConverter) BodyToGQLJSON(in []byte) (gqlschema.JSON, error) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(in, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s body to map", pretty.Resource)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s body to GQL JSON", pretty.Resource)
	}

	return result, nil
}
