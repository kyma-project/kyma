package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLimitRange_Query(t *testing.T) {
	const namespace = "default"

	t.Run("Should return limitRanges when there are some", func(t *testing.T) {
		limit1 := createMockLimitRange("limit1", namespace)
		limit2 := createMockLimitRange("limit2", namespace)
		service := setupServiceWithObjects(t, limit1, limit2)

		result, _ := service.LimitRangesQuery(context.Background(), namespace)

		assert.Equal(t, 2, len(result))
	})

	t.Run("Should return empty array if limitRange is not found", func(t *testing.T) {
		service := setupServiceWithObjects(t)

		result, _ := service.LimitRangesQuery(context.Background(), namespace)

		assert.Equal(t, 0, len(result))
	})
}

func TestLimitRange_JsonField(t *testing.T) {
	const namespace = "default"

	t.Run("JsonField returns no error", func(t *testing.T) {
		limit1 := createMockLimitRange("limit1", namespace)

		service := setupServiceWithObjects(t, limit1)

		_, err := service.LimitRangeJSONfield(context.Background(), limit1)

		require.NoError(t, err)
	})

}

func createMockLimitRange(name, namespace string) *v1.LimitRange {
	return &v1.LimitRange{
		TypeMeta: k8sMeta.TypeMeta{
			Kind:       "LimitRange",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: k8sMeta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
