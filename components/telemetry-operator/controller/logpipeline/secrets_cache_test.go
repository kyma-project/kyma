package logpipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetEmpty(t *testing.T) {
	cache := newSecretsCache()
	name := types.NamespacedName{Name: "secret-1"}

	require.Empty(t, cache.get(name))
}

func TestAddOrUpdate(t *testing.T) {
	cache := newSecretsCache()
	name := types.NamespacedName{Name: "secret-1"}
	cache.addOrUpdate(name, "pipeline-1")

	require.Len(t, cache.get(name), 1)
	require.Contains(t, cache.get(name), "pipeline-1")
}

func TestAddOrUpdateMultiple(t *testing.T) {
	cache := newSecretsCache()
	name := types.NamespacedName{Name: "secret-1"}
	cache.addOrUpdate(name, "pipeline-1")
	name2 := types.NamespacedName{Name: "secret-2"}
	cache.addOrUpdate(name2, "pipeline-2a")
	cache.addOrUpdate(name2, "pipeline-2b")

	require.Len(t, cache.get(name), 1)
	require.Len(t, cache.get(name2), 2)
	require.Contains(t, cache.get(name), "pipeline-1")
	require.Contains(t, cache.get(name2), "pipeline-2a")
	require.Contains(t, cache.get(name2), "pipeline-2b")
}

func TestDelete(t *testing.T) {
	cache := newSecretsCache()
	name := types.NamespacedName{Name: "secret-1"}
	cache.addOrUpdate(name, "pipeline-1")
	require.Len(t, cache.get(name), 1)

	cache.delete(name, "pipeline-1")
	require.Empty(t, cache.get(name))
}

func TestDeleteMultiple(t *testing.T) {
	cache := newSecretsCache()
	name := types.NamespacedName{Name: "secret-1"}
	cache.addOrUpdate(name, "pipeline-1")
	name2 := types.NamespacedName{Name: "secret-2"}
	cache.addOrUpdate(name2, "pipeline-2a")
	cache.addOrUpdate(name2, "pipeline-2b")

	cache.delete(name, "pipeline-1")
	require.Empty(t, cache.get(name))
	require.Len(t, cache.get(name2), 2)

	cache.delete(name2, "pipeline-2a")
	require.NotEmpty(t, cache.get(name2))
	require.Len(t, cache.get(name2), 1)
	require.Contains(t, cache.get(name2), "pipeline-2b")
}

func TestDeleteNonExistingIsNoOp(t *testing.T) {
	cache := newSecretsCache()
	name := types.NamespacedName{Name: "secret-1"}
	cache.delete(name, "pipeline-1")
}
