package resource

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func ExtractRawToMap(fieldName string, raw []byte) (map[string]interface{}, error) {
	extracted := make(map[string]interface{})
	err := json.Unmarshal(raw, &extracted)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting raw value of field %s", fieldName)
	}

	return extracted, nil
}

func ToStringPtr(val interface{}) *string {
	var result string

	if val != nil {
		valStr, ok := val.(string)
		if !ok {
			return nil
		}

		result = valStr
	}

	return &result
}
