package ui

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestServiceClassConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := backendModuleConverter{}
		name := "test-name"

		item := v1alpha1.BackendModule{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		expected := gqlschema.BackendModule{
			Name: name,
		}

		result, err := converter.ToGQL(&item)

		assert.Equal(t, err, nil)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &backendModuleConverter{}
		_, err := converter.ToGQL(&v1alpha1.BackendModule{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &backendModuleConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}
