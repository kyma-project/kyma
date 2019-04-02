package k8s

import (
	"bytes"
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type secretConverter struct{}

func (s *secretConverter) ToGQL(in *v1.Secret) (*gqlschema.Secret, error) {
	if in == nil {
		return nil, nil
	}

	gqlJSON, err := s.secretToGQLJSON(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s `%s` to it's json representation", pretty.Secret, in.Name)
	}

	out := &gqlschema.Secret{
		Name:         in.Name,
		Namespace:    in.ObjectMeta.Namespace,
		CreationTime: in.ObjectMeta.CreationTimestamp.Time,
		Type:         string(in.Type),
		JSON:         gqlJSON,
	}

	out.Data = make(gqlschema.JSON)
	for k, v := range in.Data {
		out.Data[k] = string(v)
	}

	out.Labels = make(gqlschema.JSON)
	for k, v := range in.ObjectMeta.Labels {
		out.Labels[k] = v
	}

	out.Annotations = make(gqlschema.JSON)

	for k, v := range in.ObjectMeta.Annotations {
		out.Annotations[k] = v
	}

	return out, nil
}

func (s *secretConverter) ToGQLs(in []*v1.Secret) ([]gqlschema.Secret, error) {
	var result []gqlschema.Secret
	for _, u := range in {
		converted, err := s.ToGQL(u)
		if err != nil {
			return nil, err
		}
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *secretConverter) GQLJSONToSecret(in gqlschema.JSON) (v1.Secret, error) {
	var buf bytes.Buffer
	in.MarshalGQL(&buf)
	bufBytes := buf.Bytes()
	result := v1.Secret{}
	err := json.Unmarshal(bufBytes, &result)
	if err != nil {
		return v1.Secret{}, errors.Wrapf(err, "while unmarshalling GQL JSON of %s", pretty.Secret)
	}

	return result, nil
}

func (c *secretConverter) secretToGQLJSON(in *v1.Secret) (gqlschema.JSON, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s `%s`", pretty.Secret, in.Name)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to map", pretty.Secret, in.Name)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s `%s` to GQL JSON", pretty.Secret, in.Name)
	}

	return result, nil
}
