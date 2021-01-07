package utils

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestMergeMaps(t *testing.T) {
	t.Run("should add/override values from overrides map to original map", func(t *testing.T) {
		//given
		original := map[string]interface{}{
			"key1": "value1",
			"key2": 2,
			"key3": map[string]interface{}{
				"key4": "value4",
			},
		}

		overrides := map[string]interface{}{
			"key2": 1,
			"key3": map[string]interface{}{
				"key4": 4,
				"key5": "value5",
			},
		}

		expected := map[string]interface{}{
			"key1": "value1",
			"key2": 1,
			"key3": map[string]interface{}{
				"key4": 4,
				"key5": "value5",
			},
		}

		//when
		MergeMaps(original, overrides)

		//then
		assert.Equal(t, original, expected)
	})
}
