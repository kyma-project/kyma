package utils

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestOverridesMap_MergeLabelOverrides(t *testing.T) {
	tests := []struct {
		name           string
		source         InterfaceMap
		lookedFor      StringMap
		expectedResult bool
	}{
		{"should return true when all source entries present present in string map",
			map[string]interface{}{
				"simple": map[string]interface{}{
					"multiple": map[string]interface{}{
						"prop1": "value1",
						"prop2": "value2",
						"prop3": "value3",
					},
				},
			},
			StringMap{
				"simple.multiple.prop1": "value1",
				"simple.multiple.prop2": "value2",
				"simple.multiple.prop3": "value3",
			},
			true,
		},
		{"should return false when source does not contain all entries",
			map[string]interface{}{
				"simple": map[string]interface{}{
					"multiple": map[string]interface{}{
						"prop1": "value1",
						"prop2": "value2",
					},
				},
			},
			StringMap{
				"simple.multiple.prop1": "value1",
				"simple.multiple.prop2": "value2",
				"simple.multiple.prop3": "value3",
			},
			false,
		},
		{"should return false when source is empty",
			map[string]interface{}{},
			StringMap{
				"simple.multiple.prop1": "value1",
			},
			false,
		},
		{"should return false when target is empty",
			map[string]interface{}{},
			StringMap{
				"simple.multiple.prop1": "value1",
			},
			false,
		},
		{"should return false when lookedFor is empty",
			map[string]interface{}{
				"simple": map[string]interface{}{
					"multiple": map[string]interface{}{
						"prop1": "value1",
						"prop2": "value2",
					},
				},
			},
			StringMap{},
			false,
		},
		{"should return true when both empty",
			map[string]interface{}{},
			StringMap{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stringMap := NewStringMap(tt.source)
			actualResult := stringMap.ContainsAll(tt.lookedFor)
			assert.Equal(t, tt.expectedResult, actualResult)
		})
	}
}
