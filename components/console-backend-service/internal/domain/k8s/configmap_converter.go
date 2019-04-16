package k8s

import (
	"bytes"
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/state"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type configMapConverter struct {
	extractor state.ContainerExtractor
}

func (c *configMapConverter) ToGQL(in *v1.ConfigMap) (*gqlschema.ConfigMap, error) {
	if in == nil {
		return nil, nil
	}

	gqlJSON, err := c.configMapToGQLJSON(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s `%s` to it's json representation", pretty.ConfigMap, in.Name)
	}

	return &gqlschema.ConfigMap{
		Name:              in.Name,
		Namespace:         in.Namespace,
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            in.Labels,
		JSON:              gqlJSON,
	}, nil
}

func (c *configMapConverter) ToGQLs(in []*v1.ConfigMap) ([]gqlschema.ConfigMap, error) {
	var result []gqlschema.ConfigMap
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

func (c *configMapConverter) configMapToGQLJSON(in *v1.ConfigMap) (gqlschema.JSON, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s `%s`", pretty.ConfigMap, in.Name)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to map", pretty.ConfigMap, in.Name)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to GQL JSON", pretty.ConfigMap, in.Name)
	}

	return result, nil
}

func (c *configMapConverter) GQLJSONToConfigMap(in gqlschema.JSON) (v1.ConfigMap, error) {
	var buf bytes.Buffer
	in.MarshalGQL(&buf)
	bufBytes := buf.Bytes()
	result := v1.ConfigMap{}
	err := json.Unmarshal(bufBytes, &result)
	if err != nil {
		return v1.ConfigMap{}, errors.Wrapf(err, "while unmarshalling GQL JSON of %s", pretty.ConfigMap)
	}

	return result, nil
}
