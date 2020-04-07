package extractor_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTriggerUnstructuredExtractor_Do(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		extractor := extractor.TriggerUnstructuredExtractor{}
		obj := testingUtils.NewUnstructured("eventing.knative.dev/v1alpha1", "Trigger", map[string]interface{}{
			"name":      "ExampleName",
			"namespace": "ExampleNamespace",
		}, nil, nil)
		expected := &v1alpha1.Trigger{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Trigger",
				APIVersion: "eventing.knative.dev/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ExampleName",
				Namespace: "ExampleNamespace",
			},
		}

		result, err := extractor.Do(obj)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		extractor := extractor.TriggerUnstructuredExtractor{}

		result, err := extractor.Do(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid type", func(t *testing.T) {
		extractor := extractor.TriggerUnstructuredExtractor{}

		result, err := extractor.Do(new(struct{}))
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
