package convert_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/convert"
	"testing"

	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionConvert_UnstructuredToFunction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		apiVersion := "serverless.kyma-project.io" // v1alpha1.SchemeGroupVersion.String()
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

		result, err := convert.UnstructuredToFunction(obj)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		result, err := convert.UnstructuredToFunction(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFunctionConvert_FunctionToUnstructured(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		apiVersion := "serverless.kyma-project.io" // v1alpha1.SchemeGroupVersion.String()
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
			map[string]interface{} {
				"name": "ExampleName",
				"creationTimestamp": nil,
			},
			map[string]interface{} {
				"function": "",
				"functionContentType": "",
				"runtime": "",
				"size": "",
			},
			map[string]interface{} {},
		)

		result, err := convert.FunctionToUnstructured(obj)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		result, err := convert.FunctionToUnstructured(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}