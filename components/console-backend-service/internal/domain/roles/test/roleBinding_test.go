package test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/rbac/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRoleBindingsService_Query(t *testing.T) {
	const namespace = "default"

	t.Run("Should filter by namespace", func(t *testing.T) {
		binding1 := createMockRoleBinding("role binding1 1", namespace, "role")
		binding2 := createMockRoleBinding("role binding1 2", "other-namespace", "role")
		service := setupServiceWithObjects(t, runtime.Object(binding1), runtime.Object(binding2))

		result, err := service.RoleBindingsQuery(context.Background(), namespace)

		require.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, binding1.Name, result[0].Name)
		assert.Equal(t, binding1.Namespace, result[0].Namespace)
	})
}

func TestRoleBindingsService_Create(t *testing.T) {
	const namespace = "default"
	const name = "binding"
	input := gqlschema.RoleBindingInput{
		RoleName: "role",
		RoleKind: "Role",
		Subjects: nil,
	}

	t.Run("Should create", func(t *testing.T) {
		service := setupServiceWithObjects(t)
		result, err := service.CreateRoleBinding(context.Background(), namespace, name, input)

		require.NoError(t, err)
		assert.Equal(t, result.Name, name)
	})

	t.Run("Should return error on duplicate", func(t *testing.T) {
		binding1 := createMockRoleBinding(name, namespace, "role")
		service := setupServiceWithObjects(t, runtime.Object(binding1))

		_, err := service.CreateRoleBinding(context.Background(), namespace, name, input)

		require.Error(t, err)
	})
}

func TestRoleBindingsService_Delete(t *testing.T) {
	const namespace = "default"
	const name = "binding"

	t.Run("Should delete", func(t *testing.T) {
		binding := createMockRoleBinding(name, namespace, "role")
		service := setupServiceWithObjects(t, runtime.Object(binding))

		result, err := service.DeleteRoleBinding(context.Background(), namespace, name)

		require.NoError(t, err)
		assert.Equal(t, result.Name, name)
	})

	t.Run("Should return error on not found", func(t *testing.T) {
		binding := createMockRoleBinding(name, "other-namespace", "role")
		service := setupServiceWithObjects(t, runtime.Object(binding))

		_, err := service.DeleteRoleBinding(context.Background(), namespace, name)

		require.Error(t, err)
	})
}

func createMockRoleBinding(name, namespace, roleName string) *v1.RoleBinding {
	return &v1.RoleBinding{
		TypeMeta: k8sMeta.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subjects: nil,
		RoleRef: v1.RoleRef{
			Kind: "Role",
			Name: roleName,
		},
	}
}
