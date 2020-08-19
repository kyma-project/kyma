package resource_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"

	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFromUnstructured(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		apiVersion := "serverless.kyma-project.io"
		obj := testingUtils.NewUnstructured(apiVersion, "Function", map[string]interface{}{
			"name": "ExampleName",
		}, nil, nil)
		expected := &v1alpha1.Function{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Function",
				APIVersion: apiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ExampleName",
			},
		}

		result := &v1alpha1.Function{}
		err := resource.FromUnstructured(obj, result)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		var result *v1alpha1.Function
		err := resource.FromUnstructured(nil, result)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestToUnstructured(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		apiVersion := "serverless.kyma-project.io"
		obj := &v1alpha1.Function{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Function",
				APIVersion: apiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ExampleName",
			},
		}
		expected := testingUtils.NewUnstructured(apiVersion, "Function",
			map[string]interface{}{
				"name":              "ExampleName",
				"creationTimestamp": nil,
			},
			map[string]interface{}{
				"source":    "",
				"resources": map[string]interface{}{},
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
