package resource

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func ToJson(obj runtime.Object) (gqlschema.JSON, error) {
	if obj == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %T", obj)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %T to map", obj)
	}

	result := gqlschema.JSON{}
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling `%T` to GQL JSON", obj)
	}

	return result, nil
}
