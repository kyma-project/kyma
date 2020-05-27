package resource_test

import (
	v1 "k8s.io/api/core/v1"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFromUnstructured(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		obj := testingUtils.NewUnstructured("v1", "Pod", map[string]interface{}{
			"name": "ExampleName",
		}, nil, nil)
		expected := &v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ExampleName",
			},
		}

		result := &v1.Pod{}
		err := resource.FromUnstructured(obj, result)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		var result *v1.Pod
		err := resource.FromUnstructured(nil, result)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestToUnstructured(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		obj := &v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ExampleName",
			},
		}
		expected := testingUtils.NewUnstructured("v1", "Pod",
			map[string]interface{}{
				"name":              "ExampleName",
				"creationTimestamp": nil,
			},
			map[string]interface{}{
				"containers": nil,
			},
			map[string]interface{}{},
		)

		result, err := resource.ToUnstructured(obj)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		result, err := resource.ToUnstructured(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
