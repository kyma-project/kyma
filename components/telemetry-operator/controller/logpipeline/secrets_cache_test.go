package logpipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestSecretsCache(t *testing.T) {
	cache := newSecretsCache()

	require.Empty(t, cache.get(types.NamespacedName{Name: "dummy"}))

	cache.set(types.NamespacedName{Name: "dummy"}, "references-dummy")

	require.Len(t, cache.get(types.NamespacedName{Name: "dummy"}), 1)
	require.Contains(t, cache.get(types.NamespacedName{Name: "dummy"}), "references-dummy")
}
