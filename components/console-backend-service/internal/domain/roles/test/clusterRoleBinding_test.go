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

func TestClusterRoleBindingsService_Query(t *testing.T) {
	t.Run("Should return bindings", func(t *testing.T) {
		binding1 := createMockClusterRoleBinding("clusterRole binding1 1", "clusterRole")
		binding2 := createMockClusterRoleBinding("clusterRole binding1 2", "clusterRole")
		service := setupServiceWithObjects(t, runtime.Object(binding1), runtime.Object(binding2))

		result, err := service.ClusterRoleBindingsQuery(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestClusterRoleBindingsService_Create(t *testing.T) {
	const name = "binding"
	input := gqlschema.ClusterRoleBindingInput{
		RoleName: "clusterRole",
		Subjects: nil,
	}

	t.Run("Should create", func(t *testing.T) {
		service := setupServiceWithObjects(t)
		result, err := service.CreateClusterRoleBinding(context.Background(), name, input)

		require.NoError(t, err)
		assert.Equal(t, result.Name, name)
	})

	t.Run("Should return error on duplicate", func(t *testing.T) {
		binding1 := createMockClusterRoleBinding(name, "clusterRole")
		service := setupServiceWithObjects(t, runtime.Object(binding1))

		_, err := service.CreateClusterRoleBinding(context.Background(), name, input)

		require.Error(t, err)
	})
}

func TestClusterRoleBindingsService_Delete(t *testing.T) {
	const name = "binding"

	t.Run("Should delete", func(t *testing.T) {
		binding := createMockClusterRoleBinding(name, "clusterRole")
		service := setupServiceWithObjects(t, runtime.Object(binding))

		result, err := service.DeleteClusterRoleBinding(context.Background(), name)

		require.NoError(t, err)
		assert.Equal(t, result.Name, name)
	})

	t.Run("Should return error on not found", func(t *testing.T) {
		binding := createMockClusterRoleBinding("other binding", "clusterRole")
		service := setupServiceWithObjects(t, runtime.Object(binding))

		_, err := service.DeleteClusterRoleBinding(context.Background(), name)

		require.Error(t, err)
	})
}

func createMockClusterRoleBinding(name, clusterRoleName string) *v1.ClusterRoleBinding {
	return &v1.ClusterRoleBinding{
		TypeMeta: k8sMeta.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: k8sMeta.ObjectMeta{
			Name: name,
		},
		Subjects: nil,
		RoleRef: v1.RoleRef{
			Kind: "ClusterRole",
			Name: clusterRoleName,
		},
	}
}
