package resource_watcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNamespaceService_GetNamespaces(t *testing.T) {
	fixNamespace1 := fixNamespace("include1", true)
	fixNamespace2 := fixNamespace("include2", false)
	fixNamespace3 := fixNamespace(excludedNamespace1, false)
	fixNamespace4 := fixNamespace(excludedNamespace2, false)

	t.Run("Success", func(t *testing.T) {
		service := fixNamespacesService(fixNamespace1, fixNamespace2, fixNamespace3, fixNamespace4)
		namespaces, err := service.GetNamespaces()

		require.NoError(t, err)
		assert.Len(t, namespaces, 1)
		assert.Exactly(t, namespaces[0], "include2")
	})
}

func TestNamespaceService_IsExcludedNamespace(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		service := fixNamespacesService()
		assert.True(t, service.IsExcludedNamespace(excludedNamespace1))
	})

	t.Run("False", func(t *testing.T) {
		service := fixNamespacesService()
		assert.False(t, service.IsExcludedNamespace("foo"))
	})
}

func fixNamespacesService(objects ...runtime.Object) *NamespaceService {
	client := fixFakeClientset(objects...)
	return NewNamespaceService(client.CoreV1(), Config{
		ExcludedNamespaces: excludedNamespaces,
	})
}
