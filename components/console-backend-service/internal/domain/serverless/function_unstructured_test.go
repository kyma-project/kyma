package serverless

import (
	"testing"

	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionUnstructuredExtractor_Do(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		extractor := newFunctionUnstructuredExtractor()

		obj := testingUtils.NewUnstructured(v1alpha1.GroupVersion.String(), "Function", map[string]interface{}{
			"name":      "ExampleName",
			"namespace": "ExampleNamespace",
		}, nil, nil)

		expected := &v1alpha1.Function{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Function",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ExampleName",
				Namespace: "ExampleNamespace",
			},
		}

		result, err := extractor.do(obj)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		extractor := newFunctionUnstructuredExtractor()

		result, err := extractor.do(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid type", func(t *testing.T) {
		extractor := newFunctionUnstructuredExtractor()

		result, err := extractor.do(new(struct{}))
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
