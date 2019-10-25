package extractor_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/extractor"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiRuleUnstructuredExtractor_Do(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		extractor := extractor.ApiRuleUnstructuredExtractor{}
		obj := testingUtils.NewUnstructured(v1alpha1.GroupVersion.String(), "APIRule", map[string]interface{}{
			"name": "ExampleName",
		}, nil, nil)
		expected := &v1alpha1.APIRule{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIRule",
				APIVersion: "gateway.kyma-project.io/v1alpha1",
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
		extractor := extractor.ApiRuleUnstructuredExtractor{}

		result, err := extractor.Do(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid type", func(t *testing.T) {
		extractor := extractor.ApiRuleUnstructuredExtractor{}

		result, err := extractor.Do(new(struct{}))
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
