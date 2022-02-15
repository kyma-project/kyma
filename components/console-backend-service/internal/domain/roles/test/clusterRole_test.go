package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/rbac/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestClusterRolesService_Query(t *testing.T) {
	t.Run("Should return existing roles", func(t *testing.T) {
		clusterRole1 := createMockClusterRole("role 1")
		clusterRole2 := createMockClusterRole("role 2")
		service := setupServiceWithObjects(t, runtime.Object(clusterRole1), runtime.Object(clusterRole2))

		result, err := service.ClusterRolesQuery(context.Background())

		require.NoError(t, err)

		assert.Equal(t, 2, len(result))
	})

	t.Run("Should filter by name", func(t *testing.T) {
		const name = "clusterRole 1"
		clusterRole1 := createMockClusterRole(name)
		clusterRole2 := createMockClusterRole("clusterRole 2")
		service := setupServiceWithObjects(t, runtime.Object(clusterRole1), runtime.Object(clusterRole2))

		result, err := service.ClusterRoleQuery(context.Background(), name)

		require.NoError(t, err)

		assert.Equal(t, name, result.Name)
	})

	t.Run("Should return error if clusterRole is not found", func(t *testing.T) {
		clusterRole := createMockClusterRole("clusterRole 1")
		service := setupServiceWithObjects(t, clusterRole)

		_, err := service.ClusterRoleQuery(context.Background(), "clusterRole 2")

		require.Error(t, err)
	})
}

func createMockClusterRole(name string) *v1.ClusterRole {
	return &v1.ClusterRole{
		TypeMeta: k8sMeta.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: name,
		},
		Rules: nil,
	}
}
