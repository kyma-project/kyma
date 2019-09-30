package main

import (
	"fmt"
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
		{
			value:     "/",
			namespace: defaultNamespace,
			name:      "",
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("should parse \"%s\" namespaced name", test.value), func(t *testing.T) {
			namespaceName := parseNamespacedName(test.value)
			assert.Equal(t, test.namespace, namespaceName.Namespace)
			assert.Equal(t, test.name, namespaceName.Name)
		})
	}
}
