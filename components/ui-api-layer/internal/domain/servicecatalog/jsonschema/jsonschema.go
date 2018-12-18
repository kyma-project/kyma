package jsonschema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

func Unpack(raw []byte) (*gqlschema.JSON, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	schema := &gqlschema.JSON{}
	var err error

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	_, err = base64.StdEncoding.Decode(decoded, raw)
	if err != nil {
		*schema, err = extract(raw)
	} else {
		// We need to trim all null characters, because json.Unmarshal is failing when we give
		// data of undesirable length to base64.StdEncoding.DecodedLen function
		decoded = bytes.Trim(decoded, "\x00")
		*schema, err = extract(decoded)
	}

	if len(*schema) == 0 || err != nil {
		return nil, err
	}
	return schema, nil
}

func extract(raw []byte) (map[string]interface{}, error) {
	extracted := make(map[string]interface{})

	err := json.Unmarshal(raw, &extracted)
	if err != nil {
		return nil, errors.Wrap(err, "while extracting creation parameter schema")
	}
	if isEmpty(extracted) {
		return nil, nil
	}

	return extracted, nil
}

func isEmpty(schema map[string]interface{}) bool {
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
