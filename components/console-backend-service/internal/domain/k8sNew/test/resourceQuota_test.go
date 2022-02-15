package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceQuota_Query(t *testing.T) {
	const namespace = "default"

	t.Run("Should return resourceQuotas when there are some", func(t *testing.T) {
		res1 := createMockResourceQuota("resource1", namespace)
		res2 := createMockResourceQuota("resource2", namespace)
		service := setupServiceWithObjects(t, res1, res2)

		result, _ := service.ResourceQuotasQuery(context.Background(), namespace)

		assert.Equal(t, 2, len(result))
	})

	t.Run("Should return empty array if limitRange is not found", func(t *testing.T) {
		service := setupServiceWithObjects(t)

		result, _ := service.ResourceQuotasQuery(context.Background(), namespace)

		assert.Equal(t, 0, len(result))
	})
}

func TestResourceQuota_JsonField(t *testing.T) {
	const namespace = "default"

	t.Run("JsonField returns no error", func(t *testing.T) {
		res1 := createMockResourceQuota("res1", namespace)

		service := setupServiceWithObjects(t, res1)

		_, err := service.ResourceQuotaJSONfield(context.Background(), res1)

		require.NoError(t, err)
	})

}

func createMockResourceQuota(name, namespace string) *v1.ResourceQuota {
	return &v1.ResourceQuota{
		TypeMeta: k8sMeta.TypeMeta{
			Kind:       "ResourceQuota",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
