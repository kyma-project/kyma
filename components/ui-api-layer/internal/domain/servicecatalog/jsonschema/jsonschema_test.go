package jsonschema

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchema(t *testing.T) {
	for tn, tc := range map[string]struct {
		given    map[string]interface{}
		expected interface{}
	}{
		"With properties": {
			given: map[string]interface{}{
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"field": "value",
				},
			},
			expected: &gqlschema.JSON{
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"field": "value",
				},
			},
		},
		"With ref": {
			given: map[string]interface{}{
				"additionalProperties": false,
				"$ref":                 "reference",
			},
			expected: &gqlschema.JSON{
				"additionalProperties": false,
				"$ref":                 "reference",
			},
		},
		"Empty properties": {
			given: map[string]interface{}{
				"additionalProperties": false,
				"properties":           map[string]interface{}{},
			},
			expected: nil,
		},
		"Empty ref": {
			given: map[string]interface{}{
				"additionalProperties": false,
				"$ref":                 "",
			},
			expected: nil,
		},
		"Empty properties and ref": {
			given: map[string]interface{}{
				"additionalProperties": false,
				"properties":           map[string]interface{}{},
				"$ref":                 "",
			},
			expected: nil,
		},
		"Without properties and ref": {
			given: map[string]interface{}{
				"field": "value",
				"field2": map[string]interface{}{
					"subField": "value",
				},
			},
			expected: nil,
		},
		"Empty": {
			given:    map[string]interface{}{},
			expected: nil,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			parameterSchemaBytes, err := json.Marshal(tc.given)
			encodedParameterSchemaBytes := make([]byte, base64.StdEncoding.EncodedLen(len(parameterSchemaBytes)))
			base64.StdEncoding.Encode(encodedParameterSchemaBytes, parameterSchemaBytes)
			require.NoError(t, err)

			result, err := Unpack(encodedParameterSchemaBytes)
			assert.Nil(t, err)

			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
