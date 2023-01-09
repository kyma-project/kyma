package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNestedSet(t *testing.T) {
	testSet := NewNestedSet()

	testItems := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	t.Run("Add keys", func(t *testing.T) {
		for k, v := range testItems {
			testSet.AddKeyWithValue(k, v)
		}

		t.Log("Add keys to the list")
		require.ElementsMatch(t, []string{"key1", "key2"}, testSet.ListKeys())
	})

	t.Run("Check key with values", func(t *testing.T) {
		t.Log("check if key:value exists")
		require.True(t, testSet.HasKeyWithValue("key1", "value1"))

		t.Log("fail if key exists but not the value")
		require.False(t, testSet.HasKeyWithValue("key2", "value3"))

		t.Log("fail if checking unrelated key/value")
		require.False(t, testSet.HasKeyWithValue("key1", "value2"))

		t.Log("fail if key doesn't exists")
		require.False(t, testSet.HasKeyWithValue("key3", "value2"))

	})
}
