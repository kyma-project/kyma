package extractor_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterAssetGroupUnstructuredExtractor_Do(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		extractor := extractor.ClusterAssetGroupUnstructuredExtractor{}
		obj := testingUtils.NewUnstructured(v1beta1.GroupVersion.String(), "ClusterAssetGroup", map[string]interface{}{
			"name": "ExampleName",
		}, nil, nil)
		expected := &v1beta1.ClusterAssetGroup{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterAssetGroup",
				APIVersion: v1beta1.GroupVersion.String(),
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
		extractor := extractor.ClusterAssetGroupUnstructuredExtractor{}

		result, err := extractor.Do(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid type", func(t *testing.T) {
		extractor := extractor.ClusterAssetGroupUnstructuredExtractor{}

		result, err := extractor.Do(new(struct{}))
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
