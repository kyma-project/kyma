package overrides

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"

	"github.com/bmizerany/assert"
)

func TestOverridesMap_MergeLabelOverrides(t *testing.T) {
	tests := []struct {
		name           string
		mergeTarget    map[string]interface{}
		mergeSource    utils.StringMap
		expectedResult map[string]interface{}
	}{
		// Add test cases.
		{"should merge simple properties to target map",
			map[string]interface{}{"simple": "simpleValue"},
			utils.StringMap{overridePrefix + ".another": "anotherValue"},
			map[string]interface{}{
				"simple":  "simpleValue",
				"another": "anotherValue",
				"overrides": map[string]interface{}{
					"another": "anotherValue",
				},
			},
		},
		{"should ignore properties without override prefix",
			map[string]interface{}{"simple": "simpleValue"},
			utils.StringMap{
				overridePrefix + ".another": "anotherValue",
				"no.override.prop":          "anotherValue",
			},
			map[string]interface{}{
				"simple":  "simpleValue",
				"another": "anotherValue",
				"overrides": map[string]interface{}{
					"another": "anotherValue",
				},
			},
		},
		{"should unwind and merge properties to target map",
			map[string]interface{}{"simple": "simpleValue"},
			utils.StringMap{overridePrefix + ".flatten.another": "anotherValue"},
			map[string]interface{}{
				"simple": "simpleValue",
				"flatten": map[string]interface{}{
					"another": "anotherValue",
				},
				"overrides": map[string]interface{}{
					"flatten": map[string]interface{}{
						"another": "anotherValue",
					},
				},
			},
		},
		{"should ignore source maps with empty keys but merge empty values",
			map[string]interface{}{"simple": "simpleValue"},
			utils.StringMap{
				overridePrefix + ".emptyValue":         "",
				overridePrefix + ".complex.emptyValue": "",
				"":                                     "anotherValue",
			},
			map[string]interface{}{
				"simple":     "simpleValue",
				"emptyValue": "",
				"complex": map[string]interface{}{
					"emptyValue": "",
				},
				"overrides": map[string]interface{}{
					"emptyValue": "",
					"complex": map[string]interface{}{
						"emptyValue": "",
					},
				},
			},
		},
		{"should merge properties with the same keys",
			map[string]interface{}{},
			utils.StringMap{
				overridePrefix + ".key.complex": "",
				overridePrefix + ".key.friend":  "",
			},
			map[string]interface{}{
				"key": map[string]interface{}{
					"complex": "",
					"friend":  "",
				},
				"overrides": map[string]interface{}{
					"key": map[string]interface{}{
						"complex": "",
						"friend":  "",
					},
				},
			},
		},
		{"should ignore invalid dot properties",
			map[string]interface{}{},
			utils.StringMap{
				overridePrefix + ".key.complex...":    "",
				overridePrefix + "........key.friend": "",
				overridePrefix + "........":           "",
				"........":                            "",
			},
			map[string]interface{}{
				"key": map[string]interface{}{
					"complex": "",
					"friend":  "",
				},
				"overrides": map[string]interface{}{
					"key": map[string]interface{}{
						"complex": "",
						"friend":  "",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flatMap := NewFlatOverridesMap(tt.mergeSource)
			MergeLabelOverrides(flatMap, tt.mergeTarget)
			assert.Equal(t, tt.expectedResult, tt.mergeTarget)
		})
	}
}
