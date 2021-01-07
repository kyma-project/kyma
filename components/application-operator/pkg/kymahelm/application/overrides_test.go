package application

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestOverridesMap_MergeLabelOverrides(t *testing.T) {
	tests := []struct {
		name           string
		mergeTarget    map[string]interface{}
		mergeSource    StringMap
		expectedResult map[string]interface{}
	}{
		// Add test cases.
		{"should merge simple properties to target map",
			map[string]interface{}{"simple": "simpleValue"},
			StringMap{"override.another": "anotherValue"},
			map[string]interface{}{
				"simple":  "simpleValue",
				"another": "anotherValue",
			},
		},
		{"should unwind and merge properties to target map",
			map[string]interface{}{"simple": "simpleValue"},
			StringMap{"override.flatten.another": "anotherValue"},
			map[string]interface{}{
				"simple": "simpleValue",
				"flatten": map[string]interface{}{
					"another": "anotherValue",
				},
			},
		},
		{"should ignore source maps with empty keys but merge empty values",
			map[string]interface{}{"simple": "simpleValue"},
			StringMap{
				"override.emptyValue":         "",
				"override.complex.emptyValue": "",
				"":                            "anotherValue",
			},
			map[string]interface{}{
				"simple":     "simpleValue",
				"emptyValue": "",
				"complex": map[string]interface{}{
					"emptyValue": "",
				},
			},
		},
		{"should merge properties with the same keys",
			map[string]interface{}{},
			StringMap{
				"override.key.complex": "",
				"override.key.friend":  "",
			},
			map[string]interface{}{
				"key": map[string]interface{}{
					"complex": "",
					"friend":  "",
				},
			},
		},
		{"should ignore invalid dot properties",
			map[string]interface{}{},
			StringMap{
				"override.key.complex...":    "",
				"override........key.friend": "",
				"override........":           "",
				"........":                   "",
			},
			map[string]interface{}{
				"key": map[string]interface{}{
					"complex": "",
					"friend":  "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MergeLabelOverrides(tt.mergeSource, tt.mergeTarget)
			assert.Equal(t, tt.expectedResult, tt.mergeTarget)
		})
	}
}
