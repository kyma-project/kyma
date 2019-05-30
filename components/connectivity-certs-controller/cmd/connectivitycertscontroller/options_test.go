package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNamespacedName(t *testing.T) {
	testCases := []struct {
		value     string
		namespace string
		name      string
	}{
		{
			value:     "kyma-integration/ca-secret",
			namespace: "kyma-integration",
			name:      "ca-secret",
		},
		{
			value:     "ca-secret",
			namespace: defaultNamespace,
			name:      "ca-secret",
		},
		{
			value:     "/ca-secret",
			namespace: defaultNamespace,
			name:      "ca-secret",
		},
		{
			value:     "ca-secret/",
			namespace: defaultNamespace,
			name:      "ca-secret",
		},
	}

	t.Run("should parse namespaced name", func(t *testing.T) {
		for _, test := range testCases {
			namespaceName := parseNamespacedName(test.value)
			assert.Equal(t, test.namespace, namespaceName.Namespace)
			assert.Equal(t, test.name, namespaceName.Name)
		}
	})

}
