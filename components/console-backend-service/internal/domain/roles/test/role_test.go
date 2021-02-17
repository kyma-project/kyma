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

func TestRolesService_Query(t *testing.T) {
	const namespace = "default"

	t.Run("Should filter by namespace", func(t *testing.T) {
		role1 := createMockRole("role 1", namespace)
		role2 := createMockRole("role 2", "other-namespace")
		service := setupServiceWithObjects(t, runtime.Object(role1), runtime.Object(role2))

		result, err := service.RolesQuery(context.Background(), namespace)

		require.NoError(t, err)

		assert.Equal(t, 1, len(result))
		assert.Equal(t, role1.Name, result[0].Name)
		assert.Equal(t, role1.Namespace, result[0].Namespace)
	})

	t.Run("Should filter by name", func(t *testing.T) {
		const name = "role 1"
		role1 := createMockRole(name, namespace)
		role2 := createMockRole("role 2", namespace)
		service := setupServiceWithObjects(t, runtime.Object(role1), runtime.Object(role2))

		result, err := service.RoleQuery(context.Background(), namespace, name)

		require.NoError(t, err)

		assert.Equal(t, name, result.Name)
	})

	t.Run("Should return error if role is not found", func(t *testing.T) {
		role := createMockRole("role 1", namespace)
		service := setupServiceWithObjects(t, role)

		_, err := service.RoleQuery(context.Background(), namespace, "role 2")

		require.Error(t, err)
	})
}

func createMockRole(name, namespace string) *v1.Role {
	return &v1.Role{
		TypeMeta: k8sMeta.TypeMeta{
			Kind:       "Role",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Rules: nil,
	}
}
