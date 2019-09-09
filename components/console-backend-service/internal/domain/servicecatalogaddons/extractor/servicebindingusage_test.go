package extractor

import (
	"testing"

	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBindingUsageUnstructuredExtractor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		extractor := BindingUsageUnstructuredExtractor{}
		obj := testingUtils.NewUnstructured(v1alpha1.SchemeGroupVersion.String(), "ServiceBindingUsage", map[string]interface{}{
			"name": "ExampleName",
		}, nil, nil)
		expected := &v1alpha1.ServiceBindingUsage{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceBindingUsage",
				APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ExampleName",
			},
		}

		result, err := extractor.Do(obj)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		extractor := BindingUsageUnstructuredExtractor{}

		result, err := extractor.Do(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid type", func(t *testing.T) {
		extractor := BindingUsageUnstructuredExtractor{}

		result, err := extractor.Do(new(struct{}))
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
